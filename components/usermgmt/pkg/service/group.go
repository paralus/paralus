package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	authzrpcv1 "github.com/RafaySystems/rcloud-base/components/authz/proto/rpc/v1"
	authzv1 "github.com/RafaySystems/rcloud-base/components/authz/proto/types"
	"github.com/RafaySystems/rcloud-base/components/common/pkg/persistence/provider/pg"
	"github.com/RafaySystems/rcloud-base/components/common/pkg/utils"
	v3 "github.com/RafaySystems/rcloud-base/components/common/proto/types/commonpb/v3"
	userv3 "github.com/RafaySystems/rcloud-base/components/common/proto/types/userpb/v3"
	"github.com/RafaySystems/rcloud-base/components/usermgmt/internal/group/dao"
	"github.com/RafaySystems/rcloud-base/components/usermgmt/internal/models"
	"github.com/google/uuid"
	bun "github.com/uptrace/bun"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	groupKind     = "Group"
	groupListKind = "GroupList"
)

// GroupService is the interface for group operations
type GroupService interface {
	Close() error
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
	List(context.Context, *userv3.Group) (*userv3.GroupList, error)
}

// groupService implements GroupService
type groupService struct {
	dao  pg.EntityDAO
	gdao dao.GroupDAO
	l    utils.Lookup
	azc  authzrpcv1.AuthzClient
}

// NewGroupService return new group service
func NewGroupService(db *bun.DB, azc authzrpcv1.AuthzClient) GroupService {
	return &groupService{
		dao:  pg.NewEntityDAO(db),
		gdao: dao.NewGroupDAO(db),
		l:    utils.NewLookup(db),
		azc:  azc,
	}
}

func (s *groupService) deleteGroupRoleRelaitons(ctx context.Context, groupId uuid.UUID, group *userv3.Group) (*userv3.Group, error) {
	// delete previous entries
	// TODO: single delete command
	err := s.dao.DeleteX(ctx, "group_id", groupId, &models.GroupRole{})
	if err != nil {
		return &userv3.Group{}, err
	}
	err = s.dao.DeleteX(ctx, "group_id", groupId, &models.ProjectGroupRole{})
	if err != nil {
		return &userv3.Group{}, err
	}
	err = s.dao.DeleteX(ctx, "group_id", groupId, &models.ProjectGroupNamespaceRole{})
	if err != nil {
		return &userv3.Group{}, err
	}

	_, err = s.azc.DeletePolicies(ctx, &authzv1.Policy{Sub: "g:" + group.GetMetadata().GetName()})
	if err != nil {
		return &userv3.Group{}, fmt.Errorf("unable to delete gorup-role relations from authz; %v", err)
	}
	return group, nil
}

// Map roles to groups
func (s *groupService) createGroupRoleRelations(ctx context.Context, group *userv3.Group, ids parsedIds) (*userv3.Group, error) {
	// TODO: add transactions
	projectNamespaceRoles := group.GetSpec().GetProjectNamespaceRoles()

	var pgnrs []models.ProjectGroupNamespaceRole
	var pgrs []models.ProjectGroupRole
	var grs []models.GroupRole
	var ps []*authzv1.Policy
	for _, pnr := range projectNamespaceRoles {
		role := pnr.GetRole()
		entity, err := s.dao.GetIdByName(ctx, role, &models.Role{})
		if err != nil {
			return &userv3.Group{}, fmt.Errorf("unable to find role '%v'", role)
		}
		var roleId uuid.UUID
		if rle, ok := entity.(*models.Role); ok {
			roleId = rle.ID
		} else {
			return &userv3.Group{}, fmt.Errorf("unable to find role '%v'", role)
		}

		project := pnr.GetProject()
		org := group.GetMetadata().GetOrganization()
		namespaceId := pnr.GetNamespace() // TODO: lookup id from name
		switch {
		case namespaceId != 0:
			projectId, err := s.l.GetProjectId(ctx, project)
			if err != nil {
				return &userv3.Group{}, fmt.Errorf("unable to find project '%v'", project)
			}
			pgnr := models.ProjectGroupNamespaceRole{
				Trash:          false,
				RoleId:         roleId,
				PartnerId:      ids.Partner,
				OrganizationId: ids.Organization,
				GroupId:        ids.Id,
				ProjectId:      projectId,
				NamespaceId:    namespaceId,
				Active:         true,
			}
			pgnrs = append(pgnrs, pgnr)

			ps = append(ps, &authzv1.Policy{
				Sub:  "g:" + group.GetMetadata().GetName(),
				Ns:   strconv.FormatInt(namespaceId, 10),
				Proj: project,
				Org:  org,
				Obj:  role,
				Act:  "*",
			})
		case project != "":
			projectId, err := s.l.GetProjectId(ctx, project)
			if err != nil {
				return &userv3.Group{}, fmt.Errorf("unable to find project '%v'", project)
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
				Act:  "*",
			})
		default:
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
				Act:  "*",
			})
		}
	}
	if len(pgnrs) > 0 {
		_, err := s.dao.Create(ctx, &pgnrs)
		if err != nil {
			return &userv3.Group{}, err
		}
	}
	if len(pgrs) > 0 {
		_, err := s.dao.Create(ctx, &pgrs)
		if err != nil {
			return &userv3.Group{}, err
		}
	}
	if len(grs) > 0 {
		_, err := s.dao.Create(ctx, &grs)
		if err != nil {
			return &userv3.Group{}, err
		}
	}

	if len(ps) > 0 {
		success, err := s.azc.CreatePolicies(ctx, &authzv1.Policies{Policies: ps})
		if err != nil || !success.Res {
			return &userv3.Group{}, fmt.Errorf("unable to create mapping in authz; %v", err)
		}
	}

	return group, nil
}

func (s *groupService) deleteGroupAccountRelations(ctx context.Context, groupId uuid.UUID, group *userv3.Group) (*userv3.Group, error) {
	err := s.dao.DeleteX(ctx, "group_id", groupId, &models.GroupAccount{})
	if err != nil {
		return &userv3.Group{}, fmt.Errorf("unable to delete user; %v", err)
	}

	_, err = s.azc.DeleteUserGroups(ctx, &authzv1.UserGroup{Grp: "g:" + group.GetMetadata().GetName()})
	if err != nil {
		return &userv3.Group{}, fmt.Errorf("unable to delete gorup-user relations from authz; %v", err)
	}
	return group, nil
}

// Update the users(account) mapped to each group
func (s *groupService) createGroupAccountRelations(ctx context.Context, groupId uuid.UUID, group *userv3.Group) (*userv3.Group, error) {
	// TODO: add transactions
	var grpaccs []models.GroupAccount
	var ugs []*authzv1.UserGroup
	for _, account := range group.GetSpec().GetUsers() {
		// FIXME: do combined lookup
		entity, err := s.dao.GetIdByTraits(ctx, account, &models.KratosIdentities{})
		if err != nil {
			return &userv3.Group{}, fmt.Errorf("unable to find user '%v'", account)
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
			grpaccs = append(grpaccs, grp)
			ugs = append(ugs, &authzv1.UserGroup{
				Grp:  "g:" + group.GetMetadata().GetName(),
				User: "u:" + account,
			})
		}
	}
	if len(grpaccs) == 0 {
		return group, nil
	}
	_, err := s.dao.Create(ctx, &grpaccs)
	if err != nil {
		return &userv3.Group{}, err
	}

	// TODO: revert our db inserts if this fails
	// Just FYI, the succcess can be false if we delete the db directly but casbin has it available internally
	_, err = s.azc.CreateUserGroups(ctx, &authzv1.UserGroups{UserGroups: ugs})
	if err != nil {
		return &userv3.Group{}, fmt.Errorf("unable to create mapping in authz; %v", err)
	}

	return group, nil
}

func (s *groupService) getPartnerOrganization(ctx context.Context, group *userv3.Group) (uuid.UUID, uuid.UUID, error) {
	partner := group.GetMetadata().GetPartner()
	org := group.GetMetadata().GetOrganization()
	partnerId, err := s.l.GetPartnerId(ctx, partner)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	organizationId, err := s.l.GetOrganizationId(ctx, org)
	if err != nil {
		return partnerId, uuid.Nil, err
	}
	return partnerId, organizationId, nil

}

func (s *groupService) Create(ctx context.Context, group *userv3.Group) (*userv3.Group, error) {
	partnerId, organizationId, err := s.getPartnerOrganization(ctx, group)
	if err != nil {
		return nil, fmt.Errorf("unable to get partner and org id")
	}
	g, _ := s.dao.GetIdByNamePartnerOrg(ctx, group.GetMetadata().GetName(), uuid.NullUUID{UUID: partnerId, Valid: true}, uuid.NullUUID{UUID: organizationId, Valid: true}, &models.Group{})
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
	entity, err := s.dao.Create(ctx, &grp)
	if err != nil {
		return &userv3.Group{}, err
	}

	//update v3 spec
	if grp, ok := entity.(*models.Group); ok {
		// we can get previous group using the id, find users/roles from that and delete those
		group, err = s.createGroupAccountRelations(ctx, grp.ID, group)
		if err != nil {
			return &userv3.Group{}, err
		}

		group, err = s.createGroupRoleRelations(ctx, group, parsedIds{Id: grp.ID, Partner: partnerId, Organization: organizationId})
		if err != nil {
			return &userv3.Group{}, err
		}
		return group, nil
	}
	return &userv3.Group{}, fmt.Errorf("unable to create group")
}

func (s *groupService) toV3Group(ctx context.Context, group *userv3.Group, grp *models.Group) (*userv3.Group, error) {
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
	}
	users, err := s.gdao.GetUsers(ctx, grp.ID)
	if err != nil {
		return &userv3.Group{}, err
	}
	userNames := []string{}
	for _, u := range users {
		userNames = append(userNames, u.Traits["email"].(string))
	}

	roles, err := s.gdao.GetRoles(ctx, grp.ID)
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
	entity, err := s.dao.GetByID(ctx, uid, &models.Group{})
	if err != nil {
		return &userv3.Group{}, err
	}

	if grp, ok := entity.(*models.Group); ok {
		return s.toV3Group(ctx, group, grp)
	}
	return group, nil

}

func (s *groupService) GetByName(ctx context.Context, group *userv3.Group) (*userv3.Group, error) {
	name := group.GetMetadata().GetName()
	partnerId, organizationId, err := s.getPartnerOrganization(ctx, group)
	if err != nil {
		return nil, fmt.Errorf("unable to get partner and org id")
	}
	entity, err := s.dao.GetByNamePartnerOrg(ctx, name, uuid.NullUUID{UUID: partnerId, Valid: true}, uuid.NullUUID{UUID: organizationId, Valid: true}, &models.Group{})
	if err != nil {
		return &userv3.Group{}, err
	}

	if grp, ok := entity.(*models.Group); ok {
		return s.toV3Group(ctx, group, grp)
	}
	return group, nil

}

func (s *groupService) Update(ctx context.Context, group *userv3.Group) (*userv3.Group, error) {
	// TODO: inform when unchanged
	name := group.GetMetadata().GetName()
	partnerId, organizationId, err := s.getPartnerOrganization(ctx, group)
	if err != nil {
		return nil, fmt.Errorf("unable to get partner and org id")
	}
	entity, err := s.dao.GetByNamePartnerOrg(ctx, name, uuid.NullUUID{UUID: partnerId, Valid: true}, uuid.NullUUID{UUID: organizationId, Valid: true}, &models.Group{})
	if err != nil {
		return &userv3.Group{}, fmt.Errorf("no group found with name '%v'", name)
	}

	if grp, ok := entity.(*models.Group); ok {
		// TODO: are we not letting them update org/partner?
		grp.Name = group.Metadata.Name
		grp.Description = group.Metadata.Description
		grp.Type = group.Spec.Type
		grp.ModifiedAt = time.Now()

		// update account/role links
		group, err = s.deleteGroupAccountRelations(ctx, grp.ID, group)
		if err != nil {
			return &userv3.Group{}, err
		}
		group, err = s.createGroupAccountRelations(ctx, grp.ID, group)
		if err != nil {
			return &userv3.Group{}, err
		}
		group, err = s.deleteGroupRoleRelaitons(ctx, grp.ID, group)
		if err != nil {
			return &userv3.Group{}, err
		}
		group, err = s.createGroupRoleRelations(ctx, group, parsedIds{Id: grp.ID, Partner: partnerId, Organization: organizationId})
		if err != nil {
			return &userv3.Group{}, err
		}

		_, err = s.dao.Update(ctx, grp.ID, grp)
		if err != nil {
			return &userv3.Group{}, err
		}

		// update spec and status
		group.Spec = &userv3.GroupSpec{
			Type:                  grp.Type,
			Users:                 group.Spec.Users, // TODO: update from db resp or no update?
			ProjectNamespaceRoles: group.Spec.ProjectNamespaceRoles,
		}
	}

	return group, nil
}

func (s *groupService) Delete(ctx context.Context, group *userv3.Group) (*userv3.Group, error) {
	name := group.GetMetadata().GetName()
	partnerId, organizationId, err := s.getPartnerOrganization(ctx, group)
	if err != nil {
		return &userv3.Group{}, fmt.Errorf("unable to get partner and org id")
	}
	entity, err := s.dao.GetByNamePartnerOrg(ctx, name, uuid.NullUUID{UUID: partnerId, Valid: true}, uuid.NullUUID{UUID: organizationId, Valid: true}, &models.Group{})
	if err != nil {
		return &userv3.Group{}, err
	}
	if grp, ok := entity.(*models.Group); ok {
		group, err = s.deleteGroupRoleRelaitons(ctx, grp.ID, group)
		if err != nil {
			return &userv3.Group{}, err
		}
		group, err = s.deleteGroupAccountRelations(ctx, grp.ID, group)
		if err != nil {
			return &userv3.Group{}, err
		}
		err = s.dao.Delete(ctx, grp.ID, grp)
		if err != nil {
			return &userv3.Group{}, err
		}
	}

	return group, nil
}

func (s *groupService) List(ctx context.Context, group *userv3.Group) (*userv3.GroupList, error) {
	var groups []*userv3.Group
	groupList := &userv3.GroupList{
		ApiVersion: apiVersion,
		Kind:       groupListKind,
		Metadata: &v3.ListMetadata{
			Count: 0,
		},
	}
	if len(group.Metadata.Organization) > 0 {
		orgId, err := s.l.GetOrganizationId(ctx, group.Metadata.Organization)
		if err != nil {
			return groupList, err
		}
		partId, err := s.l.GetPartnerId(ctx, group.Metadata.Partner)
		if err != nil {
			return groupList, err
		}
		var grps []models.Group
		entities, err := s.dao.List(ctx, uuid.NullUUID{UUID: partId, Valid: true}, uuid.NullUUID{UUID: orgId, Valid: true}, &grps)
		if err != nil {
			return groupList, err
		}
		if grps, ok := entities.(*[]models.Group); ok {
			for _, grp := range *grps {
				entry := &userv3.Group{Metadata: group.GetMetadata()}
				entry, err = s.toV3Group(ctx, entry, &grp)
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

	} else {
		return groupList, fmt.Errorf("missing organization id in metadata")
	}
	return groupList, nil
}

func (s *groupService) Close() error {
	return s.dao.Close()
}
