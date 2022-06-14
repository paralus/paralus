package service

import (
	"context"
	"fmt"

	"github.com/casbin/casbin/v2"
	"github.com/paralus/paralus/internal/dao"
	"github.com/paralus/paralus/internal/models"
	authzpbv1 "github.com/paralus/paralus/proto/types/authz"
	"github.com/uptrace/bun"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthzService interface {
	Enforce(context.Context, *authzpbv1.EnforceRequest) (*authzpbv1.BoolReply, error)
	ListPolicies(context.Context, *authzpbv1.Policy) (*authzpbv1.Policies, error)
	CreatePolicies(context.Context, *authzpbv1.Policies) (*authzpbv1.BoolReply, error)
	DeletePolicies(context.Context, *authzpbv1.Policy) (*authzpbv1.BoolReply, error)
	ListUserGroups(context.Context, *authzpbv1.UserGroup) (*authzpbv1.UserGroups, error)
	CreateUserGroups(ctx context.Context, p *authzpbv1.UserGroups) (*authzpbv1.BoolReply, error)
	DeleteUserGroups(ctx context.Context, p *authzpbv1.UserGroup) (*authzpbv1.BoolReply, error)
	ListRolePermissionMappings(ctx context.Context, p *authzpbv1.FilteredRolePermissionMapping) (*authzpbv1.RolePermissionMappingList, error)
	CreateRolePermissionMappings(ctx context.Context, p *authzpbv1.RolePermissionMappingList) (*authzpbv1.BoolReply, error)
	DeleteRolePermissionMappings(ctx context.Context, p *authzpbv1.FilteredRolePermissionMapping) (*authzpbv1.BoolReply, error)
}

type authzService struct {
	db           *bun.DB
	enforcer     *casbin.CachedEnforcer
	mappingCache map[string][]rpmUrlAction
}

func NewAuthzService(db *bun.DB, en *casbin.CachedEnforcer) AuthzService {
	en.EnableCache(false) // disables caching in casbin
	return &authzService{
		db:           db,
		enforcer:     en,
		mappingCache: make(map[string][]rpmUrlAction),
	}
}

const (
	groupGtype = "g2"
	roleGtype  = "g"
)

type rpmUrlAction struct {
	url     string
	methods []string
}

func processRpms(rpm models.ResourcePermission) []rpmUrlAction {
	urls := []rpmUrlAction{}
	for _, rurl := range rpm.ResourceUrls {
		tmp := rpmUrlAction{
			url:     rpm.BaseUrl,
			methods: make([]string, 0),
		}
		if url, ok := rurl["url"].(string); ok {
			tmp.url = tmp.url + url
		}
		if methods, ok := rurl["methods"].([]interface{}); ok {
			for _, method := range methods {
				if met, ok := method.(string); ok {
					tmp.methods = append(tmp.methods, met)
				}
			}
		}
		urls = append(urls, tmp)
	}
	for _, raurl := range rpm.ResourceActionUrls {
		tmp := rpmUrlAction{
			url:     rpm.BaseUrl,
			methods: make([]string, 0),
		}
		if url, ok := raurl["url"].(string); ok {
			tmp.url = tmp.url + url
		}
		if methods, ok := raurl["methods"].([]interface{}); ok {
			for _, method := range methods {
				if met, ok := method.(string); ok {
					tmp.methods = append(tmp.methods, met)
				}
			}
		}
		urls = append(urls, tmp)
	}
	return urls
}

func (s *authzService) cacheResourceRolePermissions(ctx context.Context) error {
	var items []models.ResourcePermission
	entities, err := dao.ListAll(ctx, s.db, &items)
	if err != nil {
		return err
	}
	if rpms, ok := entities.(*[]models.ResourcePermission); ok {
		for _, rpm := range *rpms {
			if perm, ok := s.mappingCache[rpm.Name]; ok {
				s.mappingCache[rpm.Name] = append(perm, processRpms(rpm)...)
			} else {
				s.mappingCache[rpm.Name] = processRpms(rpm)
			}
		}
	}
	return nil
}

func (s *authzService) toPolicies(policies [][]string) *authzpbv1.Policies {
	if len(policies) == 0 {
		return &authzpbv1.Policies{}
	}

	res := &authzpbv1.Policies{}
	res.Policies = make([]*authzpbv1.Policy, len(policies))
	for i := range policies {
		res.Policies[i] = &authzpbv1.Policy{
			Sub:  policies[i][0],
			Ns:   policies[i][1],
			Proj: policies[i][2],
			Org:  policies[i][3],
			Obj:  policies[i][4],
		}
	}
	return res
}

func (s *authzService) fromPolicies(policies *authzpbv1.Policies) ([][]string, error) {
	res := [][]string{}
	for i, p := range policies.GetPolicies() {
		rule := []string{p.GetSub(), p.GetNs(), p.GetProj(), p.GetOrg(), p.GetObj()}
		for _, field := range rule {
			if field == "" {
				return res, fmt.Errorf(fmt.Sprintf("index %d: policy elements do not meet definition", i))
			}
		}
		res = append(res, rule)
	}

	return res, nil
}

func (s *authzService) toUserGroups(ug [][]string) *authzpbv1.UserGroups {
	if len(ug) == 0 {
		return &authzpbv1.UserGroups{}
	}

	res := &authzpbv1.UserGroups{}
	res.UserGroups = make([]*authzpbv1.UserGroup, len(ug))
	for i := range ug {
		res.UserGroups[i] = &authzpbv1.UserGroup{
			User: ug[i][0],
			Grp:  ug[i][1],
		}
	}

	return res
}

func (s *authzService) fromUserGroups(ugs *authzpbv1.UserGroups) ([][]string, error) {
	res := [][]string{}
	for i, p := range ugs.GetUserGroups() {
		rule := []string{p.GetUser(), p.GetGrp()}
		for _, field := range rule {
			if field == "" {
				return res, fmt.Errorf(fmt.Sprintf("index %d: request elements do not meet definition", i))
			}
		}
		res = append(res, rule)
	}

	return res, nil
}

func (s *authzService) toRolePermissionMappingList(r [][]string) *authzpbv1.RolePermissionMappingList {
	if len(r) == 0 {
		return &authzpbv1.RolePermissionMappingList{}
	}

	rpms := make(map[string][]string)
	for _, rpm := range r {
		rpms[rpm[1]] = append(rpms[rpm[1]], rpm[0])
	}

	res := &authzpbv1.RolePermissionMappingList{}
	res.RolePermissionMappingList = make([]*authzpbv1.RolePermissionMapping, len(rpms))
	var i int
	for k, v := range rpms {
		i++
		res.RolePermissionMappingList[i] = &authzpbv1.RolePermissionMapping{
			Role:       k,
			Permission: v,
		}
	}

	return res
}

func (s *authzService) fromRolePermissionMappingList(ctx context.Context, r *authzpbv1.RolePermissionMappingList) ([][]string, error) {
	if len(s.mappingCache) == 0 {
		s.cacheResourceRolePermissions(ctx)
	}

	res := [][]string{}
	for i, mapping := range r.GetRolePermissionMappingList() {
		for _, permission := range mapping.GetPermission() {
			rules := [][]string{}
			for _, rpm := range s.mappingCache[permission] {
				for _, method := range rpm.methods {
					rule := []string{rpm.url, mapping.GetRole(), method}
					for _, field := range rule {
						if field == "" {
							return res, fmt.Errorf(fmt.Sprintf("index %d: mapping elements do not meet definition", i))
						}
					}
					rules = append(rules, rule)
				}
			}
			res = append(res, rules...)
		}
	}
	return res, nil
}

func (s *authzService) Enforce(ctx context.Context, req *authzpbv1.EnforceRequest) (*authzpbv1.BoolReply, error) {
	var param interface{}
	params := make([]interface{}, 0, len(req.Params))
	for index := range req.Params {
		param = req.Params[index]
		params = append(params, param)
	}

	res, err := s.enforcer.Enforce(params...)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	return &authzpbv1.BoolReply{Res: res}, nil
}

func (s *authzService) ListPolicies(ctx context.Context, p *authzpbv1.Policy) (*authzpbv1.Policies, error) {
	return s.toPolicies(s.enforcer.GetFilteredPolicy(0, p.GetSub(), p.GetNs(), p.GetProj(), p.GetOrg(), p.GetObj())), nil
}

func (s *authzService) CreatePolicies(ctx context.Context, p *authzpbv1.Policies) (*authzpbv1.BoolReply, error) {
	if len(p.GetPolicies()) == 0 {
		return &authzpbv1.BoolReply{Res: false}, nil
	}
	policies, err := s.fromPolicies(p)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	// err could be from db, policy assertions; dispatcher, watcher updates (not pertinent)
	res, err := s.enforcer.AddPolicies(policies)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	// s.enforcer.InvalidateCache()
	return &authzpbv1.BoolReply{Res: res}, nil
}

func (s *authzService) DeletePolicies(ctx context.Context, p *authzpbv1.Policy) (*authzpbv1.BoolReply, error) {
	// err could be from db, policy assertions, cache; dispatcher, watcher updates (not pertinent)
	res, err := s.enforcer.RemoveFilteredPolicy(0, p.GetSub(), p.GetNs(), p.GetProj(), p.GetOrg(), p.GetObj())
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	// s.enforcer.InvalidateCache()
	return &authzpbv1.BoolReply{Res: res}, nil
}

func (s *authzService) ListUserGroups(ctx context.Context, p *authzpbv1.UserGroup) (*authzpbv1.UserGroups, error) {
	return s.toUserGroups(s.enforcer.GetFilteredNamedGroupingPolicy(groupGtype, 0, p.GetUser(), p.GetGrp())), nil
}

func (s *authzService) CreateUserGroups(ctx context.Context, p *authzpbv1.UserGroups) (*authzpbv1.BoolReply, error) {
	if len(p.GetUserGroups()) == 0 {
		return &authzpbv1.BoolReply{Res: false}, nil
	}

	ugs, err := s.fromUserGroups(p)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	// err could be from db, policy assertions; dispatcher, watcher updates (not pertinent)
	res, err := s.enforcer.AddNamedGroupingPolicies(groupGtype, ugs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	// s.enforcer.InvalidateCache()
	return &authzpbv1.BoolReply{Res: res}, nil
}

func (s *authzService) DeleteUserGroups(ctx context.Context, p *authzpbv1.UserGroup) (*authzpbv1.BoolReply, error) {
	// err could be from db, policy assertions, cache; dispatcher, watcher updates (not pertinent)
	res, err := s.enforcer.RemoveFilteredNamedGroupingPolicy(groupGtype, 0, p.GetUser(), p.GetGrp())
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	// s.enforcer.InvalidateCache()
	return &authzpbv1.BoolReply{Res: res}, nil
}

// NOTE: might need identifier per permission in list if inheritance is needed
func (s *authzService) ListRolePermissionMappings(ctx context.Context, p *authzpbv1.FilteredRolePermissionMapping) (*authzpbv1.RolePermissionMappingList, error) {
	// TODO: Change list of urls to permissions  (many to one)
	return s.toRolePermissionMappingList(s.enforcer.GetFilteredNamedGroupingPolicy(roleGtype, 0, p.GetRole())), nil
}

func (s *authzService) CreateRolePermissionMappings(ctx context.Context, p *authzpbv1.RolePermissionMappingList) (*authzpbv1.BoolReply, error) {
	if len(p.GetRolePermissionMappingList()) == 0 {
		return &authzpbv1.BoolReply{Res: false}, nil
	}

	rpms, err := s.fromRolePermissionMappingList(ctx, p)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	res, err := s.enforcer.AddNamedGroupingPolicies(roleGtype, rpms)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	// s.enforcer.InvalidateCache()
	return &authzpbv1.BoolReply{Res: res}, nil
}

func (s *authzService) DeleteRolePermissionMappings(ctx context.Context, p *authzpbv1.FilteredRolePermissionMapping) (*authzpbv1.BoolReply, error) {
	res, err := s.enforcer.RemoveFilteredNamedGroupingPolicy(roleGtype, 1, p.GetRole())
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	// s.enforcer.InvalidateCache()
	return &authzpbv1.BoolReply{Res: res}, nil
}
