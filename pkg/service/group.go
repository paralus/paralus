package service

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/RafaySystems/rcloud-base/internal/dao"
	"github.com/RafaySystems/rcloud-base/internal/models"
	authzv1 "github.com/RafaySystems/rcloud-base/proto/types/authz"
	v3 "github.com/RafaySystems/rcloud-base/proto/types/commonpb/v3"
	userv3 "github.com/RafaySystems/rcloud-base/proto/types/userpb/v3"
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
	db  *bun.DB
	azc AuthzService
}

// NewGroupService return new group service
func NewGroupService(db *bun.DB, azc AuthzService) GroupService {
	return &groupService{db: db, azc: azc}
}

func (s *groupService) deleteGroupRoleRelaitons(ctx context.Context, db bun.IDB, groupId uuid.UUID, group *userv3.Group) (*userv3.Group, error) {
	// delete previous entries
	// TODO: single delete command
	err := dao.DeleteX(ctx, db, "group_id", groupId, &models.GroupRole{})
	if err != nil {
		return &userv3.Group{}, err
	}
	err = dao.DeleteX(ctx, db, "group_id", groupId, &models.ProjectGroupRole{})
	if err != nil {
		return &userv3.Group{}, err
	}
	err = dao.DeleteX(ctx, db, "group_id", groupId, &models.ProjectGroupNamespaceRole{})
	if err != nil {
		return &userv3.Group{}, err
	}

	_, err = s.azc.DeletePolicies(ctx, &authzv1.Policy{Sub: "g:" + group.GetMetadata().GetName()})
	if err != nil {
		return &userv3.Group{}, fmt.Errorf("unable to delete group-role relations from authz; %v", err)
	}
	return group, nil
}

// Map roles to groups
func (s *groupService) createGroupRoleRelations(ctx context.Context, db bun.IDB, group *userv3.Group, ids parsedIds) (*userv3.Group, error) {
	// TODO: add transactions
	projectNamespaceRoles := group.GetSpec().GetProjectNamespaceRoles()

	var pgnrs []models.ProjectGroupNamespaceRole
	var pgrs []models.ProjectGroupRole
	var grs []models.GroupRole
	var ps []*authzv1.Policy
	for _, pnr := range projectNamespaceRoles {
		role := pnr.GetRole()
		entity, err := dao.GetIdByName(ctx, db, role, &models.Role{})
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
			projectId, err := dao.GetProjectId(ctx, db, project)
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
			})
		case project != "":
			projectId, err := dao.GetProjectId(ctx, db, project)
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
			})
		}
	}
	if len(pgnrs) > 0 {
		_, err := dao.Create(ctx, db, &pgnrs)
		if err != nil {
			return &userv3.Group{}, err
		}
	}
	if len(pgrs) > 0 {
		_, err := dao.Create(ctx, db, &pgrs)
		if err != nil {
			return &userv3.Group{}, err
		}
	}
	if len(grs) > 0 {
		_, err := dao.Create(ctx, db, &grs)
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

func (s *groupService) deleteGroupAccountRelations(ctx context.Context, db bun.IDB, groupId uuid.UUID, group *userv3.Group) (*userv3.Group, error) {
	err := dao.DeleteX(ctx, db, "group_id", groupId, &models.GroupAccount{})
	if err != nil {
		return &userv3.Group{}, fmt.Errorf("unable to delete user; %v", err)
	}

	_, err = s.azc.DeleteUserGroups(ctx, &authzv1.UserGroup{Grp: "g:" + group.GetMetadata().GetName()})
	if err != nil {
		return &userv3.Group{}, fmt.Errorf("unable to delete group-user relations from authz; %v", err)
	}
	return group, nil
}

// Update the users(account) mapped to each group
func (s *groupService) createGroupAccountRelations(ctx context.Context, db bun.IDB, groupId uuid.UUID, group *userv3.Group) (*userv3.Group, error) {
	// TODO: add transactions
	var grpaccs []models.GroupAccount
	var ugs []*authzv1.UserGroup
	for _, account := range unique(group.GetSpec().GetUsers()) {
		// FIXME: do combined lookup
		entity, err := dao.GetIdByTraits(ctx, db, account, &models.KratosIdentities{})
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
	_, err := dao.Create(ctx, db, &grpaccs)
	if err != nil {
		return &userv3.Group{}, err
	}

	// TODO: revert our db inserts if this fails
	// Just FYI, the success can be false if we delete the db directly but casbin has it available internally
	_, err = s.azc.CreateUserGroups(ctx, &authzv1.UserGroups{UserGroups: ugs})
	if err != nil {
		return &userv3.Group{}, fmt.Errorf("unable to create mapping in authz; %v", err)
	}

	return group, nil
}

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
		group, err = s.createGroupAccountRelations(ctx, tx, grp.ID, group)
		if err != nil {
			tx.Rollback()
			return &userv3.Group{}, err
		}

		group, err = s.createGroupRoleRelations(ctx, tx, group, parsedIds{Id: grp.ID, Partner: partnerId, Organization: organizationId})
		if err != nil {
			tx.Rollback()
			return &userv3.Group{}, err
		}

		err = tx.Commit()
		if err != nil {
			tx.Rollback()
			_log.Warn("unable to commit changes", err)
		}
		return group, nil
	}
	tx.Rollback()
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
		group, err = s.deleteGroupAccountRelations(ctx, tx, grp.ID, group)
		if err != nil {
			tx.Rollback()
			return &userv3.Group{}, err
		}
		group, err = s.createGroupAccountRelations(ctx, tx, grp.ID, group)
		if err != nil {
			tx.Rollback()
			return &userv3.Group{}, err
		}
		group, err = s.deleteGroupRoleRelaitons(ctx, tx, grp.ID, group)
		if err != nil {
			tx.Rollback()
			return &userv3.Group{}, err
		}
		group, err = s.createGroupRoleRelations(ctx, tx, group, parsedIds{Id: grp.ID, Partner: partnerId, Organization: organizationId})
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

		group, err = s.deleteGroupRoleRelaitons(ctx, s.db, grp.ID, group)
		if err != nil {
			tx.Rollback()
			return &userv3.Group{}, err
		}
		group, err = s.deleteGroupAccountRelations(ctx, s.db, grp.ID, group)
		if err != nil {
			tx.Rollback()
			return &userv3.Group{}, err
		}
		err = dao.Delete(ctx, s.db, grp.ID, grp)
		if err != nil {
			tx.Rollback()
			return &userv3.Group{}, err
		}

		err = tx.Commit()
		if err != nil {
			tx.Rollback()
			_log.Warn("unable to commit changes", err)
		}
		return group, nil
	}

	return &userv3.Group{}, fmt.Errorf("unable to delete group")
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
		orgId, err := dao.GetOrganizationId(ctx, s.db, group.Metadata.Organization)
		if err != nil {
			return groupList, err
		}
		partId, err := dao.GetPartnerId(ctx, s.db, group.Metadata.Partner)
		if err != nil {
			return groupList, err
		}
		var grps []models.Group
		entities, err := dao.List(ctx, s.db, uuid.NullUUID{UUID: partId, Valid: true}, uuid.NullUUID{UUID: orgId, Valid: true}, &grps)
		if err != nil {
			return groupList, err
		}
		if grps, ok := entities.(*[]models.Group); ok {
			for _, grp := range *grps {
				entry := &userv3.Group{Metadata: group.GetMetadata()}
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

	} else {
		return groupList, fmt.Errorf("missing organization id in metadata")
	}
	return groupList, nil
}
