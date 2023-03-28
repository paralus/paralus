package service

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/paralus/paralus/internal/dao"
	"github.com/paralus/paralus/internal/models"
	"github.com/paralus/paralus/pkg/query"
	"github.com/paralus/paralus/pkg/utils"
	authzv1 "github.com/paralus/paralus/proto/types/authz"
	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	v3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	userv3 "github.com/paralus/paralus/proto/types/userpb/v3"
	bun "github.com/uptrace/bun"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	groupKind     = "Group"
	groupListKind = "GroupList"
)

// GroupService is the interface for group operations
type GroupService interface {
	// create group
	Create(context.Context, *userv3.Group) (*userv3.Group, error)
	// get group by id
	GetByID(context.Context, *userv3.Group) (*userv3.Group, error)
	// get group by name
	GetByName(context.Context, *userv3.Group) (*userv3.Group, error)
	// create or update group
	Update(context.Context, *userv3.Group) (*userv3.Group, error)
	// delete group
	Delete(context.Context, *userv3.Group) (*userv3.Group, error)
	// list groups
	List(context.Context, ...query.Option) (*userv3.GroupList, error)
}

// groupService implements GroupService
type groupService struct {
	db  *bun.DB
	azc AuthzService
	al  *zap.Logger
}

// NewGroupService return new group service
func NewGroupService(db *bun.DB, azc AuthzService, al *zap.Logger) GroupService {
	return &groupService{db: db, azc: azc, al: al}
}

// deleteGroupRoleRelaitons deletes existing group-role relations
func (s *groupService) deleteGroupRoleRelaitons(ctx context.Context, db bun.IDB, groupId uuid.UUID, group *userv3.Group) (*userv3.Group, []uuid.UUID, error) {
	// TODO: single delete command
	ids := []uuid.UUID{}
	gr := []models.GroupRole{}
	err := dao.DeleteXR(ctx, db, "group_id", groupId, &gr)
	if err != nil {
		return &userv3.Group{}, nil, err
	}
	for _, r := range gr {
		ids = append(ids, r.RoleId)
	}

	pgr := []models.ProjectGroupRole{}
	err = dao.DeleteXR(ctx, db, "group_id", groupId, &pgr)
	if err != nil {
		return &userv3.Group{}, nil, err
	}
	for _, r := range pgr {
		ids = append(ids, r.RoleId)
	}

	pgnr := []models.ProjectGroupNamespaceRole{}
	err = dao.DeleteXR(ctx, db, "group_id", groupId, &pgnr)
	if err != nil {
		return &userv3.Group{}, nil, err
	}
	for _, r := range pgnr {
		ids = append(ids, r.RoleId)
	}

	_, err = s.azc.DeletePolicies(ctx, &authzv1.Policy{Sub: "g:" + group.GetMetadata().GetName()})
	if err != nil {
		return &userv3.Group{}, nil, fmt.Errorf("unable to delete group-role relations from authz; %v", err)
	}

	return group, ids, nil
}

// Map roles to groups
func (s *groupService) createGroupRoleRelations(ctx context.Context, db bun.IDB, group *userv3.Group, ids parsedIds) (*userv3.Group, []uuid.UUID, error) {
	projectNamespaceRoles := group.GetSpec().GetProjectNamespaceRoles()

	var pgrs []models.ProjectGroupRole
	var pgnr []models.ProjectGroupNamespaceRole
	var grs []models.GroupRole
	var ps []*authzv1.Policy
	var rids []uuid.UUID
	regexc := regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)

	for _, pnr := range projectNamespaceRoles {
		role := pnr.GetRole()
		entity, err := dao.GetByName(ctx, db, role, &models.Role{})
		if err != nil {
			return &userv3.Group{}, nil, fmt.Errorf("unable to find role '%v'", role)
		}
		var roleId uuid.UUID
		var roleName string
		var scope string
		if rle, ok := entity.(*models.Role); ok {
			roleId = rle.ID
			roleName = rle.Name
			rids = append(rids, rle.ID)
			scope = strings.ToLower(rle.Scope)
		} else {
			return &userv3.Group{}, nil, fmt.Errorf("unable to find role '%v'", role)
		}

		project := pnr.GetProject()
		org := group.GetMetadata().GetOrganization()

		switch scope {
		case "system":
			gr := models.GroupRole{
				Trash:          false,
				RoleId:         roleId,
				PartnerId:      ids.Partner,
				OrganizationId: ids.Organization,
				GroupId:        ids.Id,
				Active:         true,
			}
			grs = append(grs, gr)
			ps = append(ps, &authzv1.Policy{
				Sub:  "g:" + group.GetMetadata().GetName(),
				Ns:   "*",
				Proj: "*",
				Org:  "*",
				Obj:  role,
			})
		case "organization":
			if org == "" {
				return &userv3.Group{}, nil, fmt.Errorf("no org name provided for role '%v'", roleName)
			}
			gr := models.GroupRole{
				Trash:          false,
				RoleId:         roleId,
				PartnerId:      ids.Partner,
				OrganizationId: ids.Organization,
				GroupId:        ids.Id,
				Active:         true,
			}
			grs = append(grs, gr)
			ps = append(ps, &authzv1.Policy{
				Sub:  "g:" + group.GetMetadata().GetName(),
				Ns:   "*",
				Proj: "*",
				Org:  org,
				Obj:  role,
			})
		case "project":
			if org == "" {
				return &userv3.Group{}, nil, fmt.Errorf("no org name provided for role '%v'", roleName)
			}
			if project == "" {
				return &userv3.Group{}, nil, fmt.Errorf("no project name provided for role '%v'", roleName)
			}
			projectId, err := dao.GetProjectId(ctx, s.db, project)
			if err != nil {
				return &userv3.Group{}, nil, fmt.Errorf("unable to find project '%v'", project)
			}

			pgr := models.ProjectGroupRole{
				Trash:          false,
				RoleId:         roleId,
				PartnerId:      ids.Partner,
				OrganizationId: ids.Organization,
				GroupId:        ids.Id,
				ProjectId:      projectId,
				Active:         true,
			}
			pgrs = append(pgrs, pgr)

			ps = append(ps, &authzv1.Policy{
				Sub:  "g:" + group.GetMetadata().GetName(),
				Ns:   "*",
				Proj: project,
				Org:  org,
				Obj:  role,
			})
		case "namespace":
			if org == "" {
				return &userv3.Group{}, nil, fmt.Errorf("no org name provided for role '%v'", roleName)
			}
			if project == "" {
				return &userv3.Group{}, nil, fmt.Errorf("no project name provided for role '%v'", roleName)
			}
			projectId, err := dao.GetProjectId(ctx, s.db, project)
			if err != nil {
				return &userv3.Group{}, nil, fmt.Errorf("unable to find project '%v'", project)
			}

			namespace := pnr.GetNamespace()
			match := regexc.MatchString(namespace)
			if !match {
				return &userv3.Group{}, nil, fmt.Errorf("namespace %q is invalid", namespace)
			}
			if len(namespace) < 1 || len(namespace) > 63 {
				return &userv3.Group{}, nil, fmt.Errorf("namespace %q is invalid. must be no more than 63 characters", namespace)
			}

			pgnrObj := models.ProjectGroupNamespaceRole{
				CreatedAt:      time.Now(),
				ModifiedAt:     time.Now(),
				Trash:          false,
				PartnerId:      ids.Partner,
				OrganizationId: ids.Organization,
				RoleId:         roleId,
				GroupId:        ids.Id,
				ProjectId:      projectId,
				Namespace:      namespace,
				Active:         true,
			}
			pgnr = append(pgnr, pgnrObj)

			ps = append(ps, &authzv1.Policy{
				Sub:  "g:" + group.GetMetadata().GetName(),
				Ns:   namespace,
				Proj: project,
				Org:  org,
				Obj:  role,
			})
		default:
			if err != nil {
				return group, nil, fmt.Errorf("other scoped roles are not handled")
			}
		}
	}
	if len(pgrs) > 0 {
		_, err := dao.Create(ctx, db, &pgrs)
		if err != nil {
			return &userv3.Group{}, nil, err
		}
	}
	if len(pgnr) > 0 {
		_, err := dao.Create(ctx, db, &pgnr)
		if err != nil {
			return &userv3.Group{}, nil, err
		}
	}
	if len(grs) > 0 {
		_, err := dao.Create(ctx, db, &grs)
		if err != nil {
			return &userv3.Group{}, nil, err
		}
	}

	if len(ps) > 0 {
		success, err := s.azc.CreatePolicies(ctx, &authzv1.Policies{Policies: ps})
		if err != nil || !success.Res {
			return &userv3.Group{}, nil, fmt.Errorf("unable to create mapping in authz; %v", err)
		}
	}

	return group, rids, nil
}

func (s *groupService) deleteGroupAccountRelations(ctx context.Context, db bun.IDB, groupId uuid.UUID, group *userv3.Group) (*userv3.Group, []uuid.UUID, error) {
	ga := []models.GroupAccount{}
	err := dao.DeleteXR(ctx, db, "group_id", groupId, &ga)
	if err != nil {
		return &userv3.Group{}, nil, fmt.Errorf("unable to remove user from group user; %v", err)
	}

	_, err = s.azc.DeleteUserGroups(ctx, &authzv1.UserGroup{Grp: "g:" + group.GetMetadata().GetName()})
	if err != nil {
		return &userv3.Group{}, nil, fmt.Errorf("unable to delete group-user relations from authz; %v", err)
	}

	ids := []uuid.UUID{}
	for _, r := range ga {
		ids = append(ids, r.AccountId)
	}
	return group, ids, nil
}

// Update the users(account) mapped to each group
func (s *groupService) createGroupAccountRelations(ctx context.Context, db bun.IDB, groupId uuid.UUID, group *userv3.Group) (*userv3.Group, []uuid.UUID, error) {
	var grpaccs []models.GroupAccount
	var ugs []*authzv1.UserGroup
	var uids []uuid.UUID
	for _, account := range utils.Unique(group.GetSpec().GetUsers()) {
		// FIXME: do combined lookup
		entity, err := dao.GetUserIdByEmail(ctx, db, account, &models.KratosIdentities{})
		if err != nil {
			return &userv3.Group{}, nil, fmt.Errorf("unable to find user '%v'", account)
		}
		if acc, ok := entity.(*models.KratosIdentities); ok {
			grp := models.GroupAccount{
				CreatedAt:  time.Now(),
				ModifiedAt: time.Now(),
				Trash:      false,
				AccountId:  acc.ID,
				GroupId:    groupId,
				Active:     true,
			}
			uids = append(uids, acc.ID)
			grpaccs = append(grpaccs, grp)
			ugs = append(ugs, &authzv1.UserGroup{
				Grp:  "g:" + group.GetMetadata().GetName(),
				User: "u:" + account,
			})
		}
	}
	if len(grpaccs) == 0 {
		return group, nil, nil
	}
	_, err := dao.Create(ctx, db, &grpaccs)
	if err != nil {
		return &userv3.Group{}, nil, err
	}

	_, err = s.azc.CreateUserGroups(ctx, &authzv1.UserGroups{UserGroups: ugs})
	if err != nil {
		return &userv3.Group{}, nil, fmt.Errorf("unable to create mapping in authz; %v", err)
	}

	return group, uids, nil
}

// TODO: move this to utils, make it accept two strings (names)
func (s *groupService) getPartnerOrganization(ctx context.Context, db bun.IDB, group *userv3.Group) (uuid.UUID, uuid.UUID, error) {
	partner := group.GetMetadata().GetPartner()
	org := group.GetMetadata().GetOrganization()
	partnerId, err := dao.GetPartnerId(ctx, db, partner)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	organizationId, err := dao.GetOrganizationId(ctx, db, org)
	if err != nil {
		return partnerId, uuid.Nil, err
	}
	return partnerId, organizationId, nil

}

func (s *groupService) Create(ctx context.Context, group *userv3.Group) (*userv3.Group, error) {
	partnerId, organizationId, err := s.getPartnerOrganization(ctx, s.db, group)
	if err != nil {
		return nil, fmt.Errorf("unable to get partner and org id")
	}
	g, _ := dao.GetIdByNamePartnerOrg(ctx, s.db, group.GetMetadata().GetName(), uuid.NullUUID{UUID: partnerId, Valid: true}, uuid.NullUUID{UUID: organizationId, Valid: true}, &models.Group{})
	if g != nil {
		return nil, fmt.Errorf("group '%v' already exists", group.GetMetadata().GetName())
	}
	//convert v3 spec to internal models
	grp := models.Group{
		Name:           group.GetMetadata().GetName(),
		Description:    group.GetMetadata().GetDescription(),
		CreatedAt:      time.Now(),
		ModifiedAt:     time.Now(),
		Trash:          false,
		OrganizationId: organizationId,
		PartnerId:      partnerId,
		Type:           group.GetSpec().GetType(),
	}

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return &userv3.Group{}, err
	}

	entity, err := dao.Create(ctx, tx, &grp)
	if err != nil {
		tx.Rollback() // TODO: check errors for rollback (and do what?)
		return &userv3.Group{}, err
	}

	//update v3 spec
	if grp, ok := entity.(*models.Group); ok {
		// we can get previous group using the id, find users/roles from that and delete those
		group, usersAfter, err := s.createGroupAccountRelations(ctx, tx, grp.ID, group)
		if err != nil {
			tx.Rollback()
			return &userv3.Group{}, err
		}

		group, rolesAfter, err := s.createGroupRoleRelations(ctx, tx, group, parsedIds{Id: grp.ID, Partner: partnerId, Organization: organizationId})
		if err != nil {
			tx.Rollback()
			return &userv3.Group{}, err
		}

		err = tx.Commit()
		if err != nil {
			tx.Rollback()
			_log.Warn("unable to commit changes", err)
		}

		CreateGroupAuditEvent(ctx, s.al, s.db, AuditActionCreate, group.GetMetadata().GetName(), grp.ID, []uuid.UUID{}, usersAfter, []uuid.UUID{}, rolesAfter)
		return group, nil
	}
	return &userv3.Group{}, fmt.Errorf("unable to create group")
}

func (s *groupService) toV3Group(ctx context.Context, db bun.IDB, group *userv3.Group, grp *models.Group) (*userv3.Group, error) {
	labels := make(map[string]string)
	labels["organization"] = group.GetMetadata().GetOrganization()
	labels["partner"] = group.GetMetadata().GetPartner()

	group.ApiVersion = apiVersion
	group.Kind = groupKind
	group.Metadata = &v3.Metadata{
		Name:         grp.Name,
		Description:  grp.Description,
		Organization: group.GetMetadata().GetOrganization(),
		Partner:      group.GetMetadata().GetPartner(),
		Labels:       labels,
		ModifiedAt:   timestamppb.New(grp.ModifiedAt),
		CreatedAt:    timestamppb.New(grp.CreatedAt),
	}
	users, err := dao.GetUsers(ctx, db, grp.ID)
	if err != nil {
		return &userv3.Group{}, err
	}
	userNames := []string{}
	for _, u := range users {
		userNames = append(userNames, u.Traits["email"].(string))
	}

	roles, err := dao.GetGroupRoles(ctx, db, grp.ID)
	if err != nil {
		return &userv3.Group{}, err
	}
	group.Spec = &userv3.GroupSpec{
		Type:                  grp.Type,
		Users:                 userNames,
		ProjectNamespaceRoles: roles,
	}
	return group, nil
}

func (s *groupService) GetByID(ctx context.Context, group *userv3.Group) (*userv3.Group, error) {
	id := group.GetMetadata().GetId()
	uid, err := uuid.Parse(id)
	if err != nil {
		return &userv3.Group{}, err
	}
	entity, err := dao.GetByID(ctx, s.db, uid, &models.Group{})
	if err != nil {
		return &userv3.Group{}, err
	}

	if grp, ok := entity.(*models.Group); ok {
		return s.toV3Group(ctx, s.db, group, grp)
	}
	return group, nil

}

func (s *groupService) GetByName(ctx context.Context, group *userv3.Group) (*userv3.Group, error) {
	name := group.GetMetadata().GetName()
	partnerId, organizationId, err := s.getPartnerOrganization(ctx, s.db, group)
	if err != nil {
		return nil, fmt.Errorf("unable to get partner and org id")
	}
	entity, err := dao.GetByNamePartnerOrg(ctx, s.db, name, uuid.NullUUID{UUID: partnerId, Valid: true}, uuid.NullUUID{UUID: organizationId, Valid: true}, &models.Group{})
	if err != nil {
		return &userv3.Group{}, err
	}

	if grp, ok := entity.(*models.Group); ok {
		return s.toV3Group(ctx, s.db, group, grp)
	}
	return group, nil

}

func (s *groupService) Update(ctx context.Context, group *userv3.Group) (*userv3.Group, error) {
	// TODO: inform when unchanged
	name := group.GetMetadata().GetName()
	partnerId, organizationId, err := s.getPartnerOrganization(ctx, s.db, group)
	if err != nil {
		return nil, fmt.Errorf("unable to get partner and org id")
	}
	entity, err := dao.GetByNamePartnerOrg(ctx, s.db, name, uuid.NullUUID{UUID: partnerId, Valid: true}, uuid.NullUUID{UUID: organizationId, Valid: true}, &models.Group{})
	if err != nil {
		return &userv3.Group{}, fmt.Errorf("no group found with name '%v'", name)
	}

	if grp, ok := entity.(*models.Group); ok {
		// TODO: are we not letting them update org/partner?
		grp.Name = group.Metadata.Name
		grp.Description = group.Metadata.Description
		grp.Type = group.Spec.Type
		grp.ModifiedAt = time.Now()

		tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
		if err != nil {
			return &userv3.Group{}, err
		}

		// update account/role links
		group, usersBefore, err := s.deleteGroupAccountRelations(ctx, tx, grp.ID, group)
		if err != nil {
			tx.Rollback()
			return &userv3.Group{}, err
		}
		group, usersAfter, err := s.createGroupAccountRelations(ctx, tx, grp.ID, group)
		if err != nil {
			tx.Rollback()
			return &userv3.Group{}, err
		}
		group, rolesBefore, err := s.deleteGroupRoleRelaitons(ctx, tx, grp.ID, group)
		if err != nil {
			tx.Rollback()
			return &userv3.Group{}, err
		}
		group, rolesAfter, err := s.createGroupRoleRelations(ctx, tx, group, parsedIds{Id: grp.ID, Partner: partnerId, Organization: organizationId})
		if err != nil {
			tx.Rollback()
			return &userv3.Group{}, err
		}

		_, err = dao.Update(ctx, tx, grp.ID, grp)
		if err != nil {
			tx.Rollback()
			return &userv3.Group{}, err
		}

		err = tx.Commit()
		if err != nil {
			tx.Rollback()
			_log.Warn("unable to commit changes", err)
		}

		// update spec and status
		group.Spec = &userv3.GroupSpec{
			Type:                  grp.Type,
			Users:                 group.Spec.Users, // TODO: update from db resp or no update?
			ProjectNamespaceRoles: group.Spec.ProjectNamespaceRoles,
		}

		CreateGroupAuditEvent(ctx, s.al, s.db, AuditActionUpdate, group.GetMetadata().GetName(), grp.ID, usersBefore, usersAfter, rolesBefore, rolesAfter)
	}

	return group, nil
}

func (s *groupService) Delete(ctx context.Context, group *userv3.Group) (*userv3.Group, error) {
	name := group.GetMetadata().GetName()
	partnerId, organizationId, err := s.getPartnerOrganization(ctx, s.db, group)
	if err != nil {
		return &userv3.Group{}, fmt.Errorf("unable to get partner and org id")
	}
	entity, err := dao.GetByNamePartnerOrg(ctx, s.db, name, uuid.NullUUID{UUID: partnerId, Valid: true}, uuid.NullUUID{UUID: organizationId, Valid: true}, &models.Group{})
	if err != nil {
		return &userv3.Group{}, err
	}
	if grp, ok := entity.(*models.Group); ok {

		tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
		if err != nil {
			return &userv3.Group{}, err
		}

		group, rolesBefore, err := s.deleteGroupRoleRelaitons(ctx, tx, grp.ID, group)
		if err != nil {
			tx.Rollback()
			return &userv3.Group{}, err
		}
		group, usersBefore, err := s.deleteGroupAccountRelations(ctx, tx, grp.ID, group)
		if err != nil {
			tx.Rollback()
			return &userv3.Group{}, err
		}
		err = dao.Delete(ctx, tx, grp.ID, grp)
		if err != nil {
			tx.Rollback()
			return &userv3.Group{}, err
		}

		err = tx.Commit()
		if err != nil {
			tx.Rollback()
			_log.Warn("unable to commit changes", err)
		}

		CreateGroupAuditEvent(ctx, s.al, s.db, AuditActionDelete, group.GetMetadata().GetName(), grp.ID, usersBefore, []uuid.UUID{}, rolesBefore, []uuid.UUID{})
		return group, nil
	}

	return &userv3.Group{}, fmt.Errorf("unable to delete group")
}

func (s *groupService) List(ctx context.Context, opts ...query.Option) (*userv3.GroupList, error) {
	var groups []*userv3.Group
	groupList := &userv3.GroupList{
		ApiVersion: apiVersion,
		Kind:       groupListKind,
		Metadata: &v3.ListMetadata{
			Count: 0,
		},
	}

	queryOptions := commonv3.QueryOptions{}
	for _, opt := range opts {
		opt(&queryOptions)
	}

	orgId, err := dao.GetOrganizationId(ctx, s.db, queryOptions.Organization)
	if err != nil {
		return groupList, err
	}
	partId, err := dao.GetPartnerId(ctx, s.db, queryOptions.Partner)
	if err != nil {
		return groupList, err
	}
	var grps []models.Group
	entities, err := dao.ListFiltered(ctx, s.db,
		uuid.NullUUID{UUID: partId, Valid: true}, uuid.NullUUID{UUID: orgId, Valid: true},
		uuid.NullUUID{Valid: false},
		&grps,
		queryOptions.Q,
		queryOptions.OrderBy,
		queryOptions.Order,
		int(queryOptions.Limit),
		int(queryOptions.Offset),
	)
	if err != nil {
		return groupList, err
	}
	if grps, ok := entities.(*[]models.Group); ok {
		for _, grp := range *grps {
			entry := &userv3.Group{Metadata: &commonv3.Metadata{
				Organization: queryOptions.Organization,
				Partner:      queryOptions.Partner,
			}}
			entry, err = s.toV3Group(ctx, s.db, entry, &grp)
			if err != nil {
				return groupList, err
			}
			groups = append(groups, entry)
		}

		//update the list metadata and items response
		groupList.Metadata = &v3.ListMetadata{
			Count: int64(len(groups)),
		}
		groupList.Items = groups
	}

	return groupList, nil
}
