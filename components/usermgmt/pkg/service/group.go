package service

import (
	"context"
	"fmt"
	"time"

	v3 "github.com/RafaySystems/rcloud-base/components/common/proto/types/commonpb/v3"
	"github.com/RafaySystems/rcloud-base/components/usermgmt/pkg/internal/models"
	"github.com/RafaySystems/rcloud-base/components/usermgmt/pkg/internal/persistence/provider/pg"
	userv3 "github.com/RafaySystems/rcloud-base/components/usermgmt/proto/types/userpb/v3"
	"github.com/google/uuid"
	bun "github.com/uptrace/bun"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	apiVersion    = "system.k8smgmt.io/v3"
	groupKind     = "Group"
	groupListKind = "GroupList"
)

// GroupService is the interface for group operations
type GroupService interface {
	Close() error
	// create group
	Create(ctx context.Context, group *userv3.Group) (*userv3.Group, error)
	// get group by id
	GetByID(ctx context.Context, id string) (*userv3.Group, error)
	// get group by name
	GetByName(ctx context.Context, name string) (*userv3.Group, error)
	// create or update group
	Update(ctx context.Context, group *userv3.Group) (*userv3.Group, error)
	// delete group
	Delete(ctx context.Context, group *userv3.Group) (*userv3.Group, error)
	// list groups
	List(ctx context.Context, group *userv3.Group) (*userv3.GroupList, error)
}

// groupService implements GroupService
type groupService struct {
	dao pg.EntityDAO
}

// NewGroupService return new group service
func NewGroupService(db *bun.DB) GroupService {
	return &groupService{
		dao: pg.NewEntityDAO(db),
	}
}

// Map roles to groups
func (s *groupService) userGroupRoleRelation(ctx context.Context, group *userv3.Group) (*userv3.Group, error) {
	groupId, _ := uuid.Parse(group.GetMetadata().GetId())
	partnerId, _ := uuid.Parse(group.GetMetadata().GetPartner())
	organizationId, _ := uuid.Parse(group.GetMetadata().GetOrganization())

	// TODO: also parse out namesapce
	projectNamespaceRoles := group.GetSpec().GetProjectnamespaceroles()

	// TODO: add transactions
	var pgnrs []models.ProjectGroupNamespaceRole
	var pgrs []models.ProjectGroupRole
	var grs []models.GroupRole
	for _, pnr := range projectNamespaceRoles {
		projectId, perr := uuid.Parse(pnr.GetProject())
		namespaceId := pnr.GetNamespace()
		roleId, err := uuid.Parse(pnr.GetRole())
		if err != nil {
			return group, err
		}
		switch {
		case namespaceId != 0: // TODO: namespaceId can be zero?
			pgnr := models.ProjectGroupNamespaceRole{
				Name:           group.GetMetadata().GetName(),
				Description:    group.GetMetadata().GetDescription(),
				CreatedAt:      time.Now(),
				ModifiedAt:     time.Now(),
				Trash:          false,
				RoleId:         roleId,
				PartnerId:      partnerId,
				OrganizationId: organizationId,
				GroupId:        groupId,
				ProjectId:      projectId,
				NamesapceId:    namespaceId,
				Active:         true,
			}
			pgnrs = append(pgnrs, pgnr)
		case perr == nil: // TODO: maybe a better check?
			pgr := models.ProjectGroupRole{
				Name:           group.GetMetadata().GetName(),
				Description:    group.GetMetadata().GetDescription(),
				CreatedAt:      time.Now(),
				ModifiedAt:     time.Now(),
				Trash:          false,
				Default:        true, // TODO: what is this for?
				RoleId:         roleId,
				PartnerId:      partnerId,
				OrganizationId: organizationId,
				GroupId:        groupId,
				ProjectId:      projectId,
				Active:         true,
			}
			pgrs = append(pgrs, pgr)
		default:
			gr := models.GroupRole{
				Name:           group.GetMetadata().GetName(),
				Description:    group.GetMetadata().GetDescription(),
				CreatedAt:      time.Now(),
				ModifiedAt:     time.Now(),
				Trash:          false,
				Default:        true, // TODO: what is this for?
				RoleId:         roleId,
				PartnerId:      partnerId,
				OrganizationId: organizationId,
				GroupId:        groupId,
				Active:         true,
			}
			grs = append(grs, gr)
		}
	}
	if len(pgnrs) > 0 {
		_, err := s.dao.Create(ctx, &pgnrs)
		if err != nil {
			return group, err
		}
	}
	if len(pgrs) > 0 {
		_, err := s.dao.Create(ctx, &pgrs)
		if err != nil {
			return group, err
		}
	}
	if len(grs) > 0 {
		_, err := s.dao.Create(ctx, &grs)
		if err != nil {
			return group, err
		}
	}

	return group, nil
}

// Update the users(account) mapped to each group
func (s *groupService) updateGroupAccountRelation(ctx context.Context, group *userv3.Group) (*userv3.Group, error) {
	// TODO: use a more efficient way to update the relations
	// TODO: diff and delete the old relations
	groupId, _ := uuid.Parse(group.GetMetadata().GetId())

	// TODO: add transactions
	var grpaccs []models.GroupAccount
	for _, account := range group.GetSpec().GetUsers() {
		accountId, err := uuid.Parse(account)
		if err != nil {
			return nil, err
		}
		grp := models.GroupAccount{
			Name:        group.GetMetadata().GetName(),
			Description: group.GetMetadata().GetDescription(),
			CreatedAt:   time.Now(),
			ModifiedAt:  time.Now(),
			Trash:       false,
			AccountId:   accountId,
			GroupId:     groupId,
			Active:      true,
		}
		grpaccs = append(grpaccs, grp)
	}
	_, err := s.dao.Create(ctx, &grpaccs)
	if err != nil {
		group.Status = &v3.Status{
			ConditionType:   "Create",
			ConditionStatus: v3.ConditionStatus_StatusFailed,
			LastUpdated:     timestamppb.Now(),
		}
		return group, err
	}

	return group, nil
}

func (s *groupService) Create(ctx context.Context, group *userv3.Group) (*userv3.Group, error) {
	partnerId, _ := uuid.Parse(group.GetMetadata().GetPartner())
	organizationId, _ := uuid.Parse(group.GetMetadata().GetOrganization())
	// TODO: find out the interaction if project key is present in the group metadata
	// TODO: check if a group with the same 'name' already exists and fail if so
	// TODO: we should be specifying names instead of ids for partner and org (at least in output)
	// TODO: create vs apply difference like in kubectl??
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
		group.Status = &v3.Status{
			ConditionType:   "Create",
			ConditionStatus: v3.ConditionStatus_StatusFailed,
			LastUpdated:     timestamppb.Now(),
		}
		return group, err
	}

	//update v3 spec
	if createdGroup, ok := entity.(*models.Group); ok {
		group.Metadata.Id = createdGroup.ID.String()
		group.Spec = &userv3.GroupSpec{
			Type:  createdGroup.Type,
			Users: group.Spec.Users, // TODO: is this the right thing to do?
			Projectnamespaceroles: group.Spec.Projectnamespaceroles, // TODO: is this the right thing to do?
		}
		if group.Status != nil {
			group.Status = &v3.Status{
				ConditionType:   "Create",
				ConditionStatus: v3.ConditionStatus_StatusOK,
				LastUpdated:     timestamppb.Now(),
			}
		}
	}

	group, err = s.updateGroupAccountRelation(ctx, group)
	if err != nil {
		group.Status = &v3.Status{
			ConditionType:   "Create",
			ConditionStatus: v3.ConditionStatus_StatusFailed,
			LastUpdated:     timestamppb.Now(),
		}
		return group, err
	}

	group, err = s.userGroupRoleRelation(ctx, group)
	if err != nil {
		group.Status = &v3.Status{
			ConditionType:   "Create",
			ConditionStatus: v3.ConditionStatus_StatusFailed,
			LastUpdated:     timestamppb.Now(),
		}
		return group, err
	}

	return group, nil
}

func (s *groupService) GetByID(ctx context.Context, id string) (*userv3.Group, error) {

	group := &userv3.Group{
		ApiVersion: apiVersion,
		Kind:       groupKind,
		Metadata: &v3.Metadata{
			Id: id,
		},
	}

	uid, err := uuid.Parse(id)
	if err != nil {
		group.Status = &v3.Status{
			ConditionType:   "Describe",
			ConditionStatus: v3.ConditionStatus_StatusFailed,
			LastUpdated:     timestamppb.Now(),
			Reason:          err.Error(),
		}
		return group, err
	}
	entity, err := s.dao.GetByID(ctx, uid, &models.Group{})
	if err != nil {
		group.Status = &v3.Status{
			ConditionType:   "Describe",
			ConditionStatus: v3.ConditionStatus_StatusFailed,
			LastUpdated:     timestamppb.Now(),
			Reason:          err.Error(),
		}
		return group, err
	}

	if grp, ok := entity.(*models.Group); ok {
		labels := make(map[string]string)
		labels["organization"] = grp.OrganizationId.String()

		group.Metadata = &v3.Metadata{
			Name:         grp.Name,
			Description:  grp.Description,
			Id:           grp.ID.String(),
			Organization: grp.OrganizationId.String(),
			Partner:      grp.PartnerId.String(),
			Labels:       labels,
			ModifiedAt:   timestamppb.New(grp.ModifiedAt),
		}
		group.Spec = &userv3.GroupSpec{
			Type: grp.Type,
		}
		group.Status = &v3.Status{
			LastUpdated:     timestamppb.Now(),
			ConditionType:   "Describe",
			ConditionStatus: v3.ConditionStatus_StatusOK,
		}

		return group, nil
	}
	return group, nil

}

func (s *groupService) GetByName(ctx context.Context, name string) (*userv3.Group, error) {
	group := &userv3.Group{
		ApiVersion: apiVersion,
		Kind:       groupKind,
		Metadata: &v3.Metadata{
			Name: name,
		},
	}

	entity, err := s.dao.GetByName(ctx, name, &models.Group{})
	if err != nil {
		group.Status = &v3.Status{
			ConditionType:   "Describe",
			ConditionStatus: v3.ConditionStatus_StatusFailed,
			LastUpdated:     timestamppb.Now(),
			Reason:          err.Error(),
		}
		return group, err
	}

	if grp, ok := entity.(*models.Group); ok {
		labels := make(map[string]string)
		labels["organization"] = grp.OrganizationId.String()

		group.Metadata = &v3.Metadata{
			Name:         grp.Name,
			Description:  grp.Description,
			Id:           grp.ID.String(),
			Organization: grp.OrganizationId.String(),
			Partner:      grp.PartnerId.String(),
			Labels:       labels,
			ModifiedAt:   timestamppb.New(grp.ModifiedAt),
		}
		group.Spec = &userv3.GroupSpec{
			Type: grp.Type,
		}
		group.Status = &v3.Status{
			LastUpdated:     timestamppb.Now(),
			ConditionType:   "Describe",
			ConditionStatus: v3.ConditionStatus_StatusOK,
		}

		return group, nil
	}
	return group, nil

}

func (s *groupService) Update(ctx context.Context, group *userv3.Group) (*userv3.Group, error) {
	// TODO: inform when unchanged

	id, _ := uuid.Parse(group.Metadata.Id)
	entity, err := s.dao.GetByID(ctx, id, &models.Group{})
	if err != nil {
		group.Status = &v3.Status{
			ConditionType:   "Update",
			ConditionStatus: v3.ConditionStatus_StatusFailed,
			LastUpdated:     timestamppb.Now(),
			Reason:          err.Error(),
		}
		return group, err
	}

	if grp, ok := entity.(*models.Group); ok {
		//update group details
		grp.Name = group.Metadata.Name
		grp.Description = group.Metadata.Description
		grp.Type = group.Spec.Type
		grp.ModifiedAt = time.Now()

		_, err = s.dao.Update(ctx, id, grp)
		if err != nil {
			group.Status = &v3.Status{
				ConditionType:   "Update",
				ConditionStatus: v3.ConditionStatus_StatusFailed,
				LastUpdated:     timestamppb.Now(),
				Reason:          err.Error(),
			}
			return group, err
		}

		//update spec and status
		group.Spec = &userv3.GroupSpec{
			Type: grp.Type,
		}
		group.Status = &v3.Status{
			ConditionType:   "Update",
			ConditionStatus: v3.ConditionStatus_StatusOK,
			LastUpdated:     timestamppb.Now(),
		}
	}

	return group, nil
}

func (s *groupService) Delete(ctx context.Context, group *userv3.Group) (*userv3.Group, error) {
	id, err := uuid.Parse(group.Metadata.Id)
	if err != nil {
		group.Status = &v3.Status{
			ConditionType:   "Delete",
			ConditionStatus: v3.ConditionStatus_StatusFailed,
			LastUpdated:     timestamppb.Now(),
			Reason:          err.Error(),
		}
		return group, err
	}
	entity, err := s.dao.GetByID(ctx, id, &models.Group{})
	if err != nil {
		group.Status = &v3.Status{
			ConditionType:   "Delete",
			ConditionStatus: v3.ConditionStatus_StatusFailed,
			LastUpdated:     timestamppb.Now(),
			Reason:          err.Error(),
		}
		return group, err
	}
	if grp, ok := entity.(*models.Group); ok {
		err = s.dao.Delete(ctx, id, grp)
		if err != nil {
			group.Status = &v3.Status{
				ConditionType:   "Delete",
				ConditionStatus: v3.ConditionStatus_StatusFailed,
				LastUpdated:     timestamppb.Now(),
				Reason:          err.Error(),
			}
			return group, err
		}
		//update v3 spec
		group.Metadata.Id = grp.ID.String()
		group.Metadata.Name = grp.Name
		group.Status = &v3.Status{
			ConditionType:   "Delete",
			ConditionStatus: v3.ConditionStatus_StatusOK,
			LastUpdated:     timestamppb.Now(),
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
		orgId, err := uuid.Parse(group.Metadata.Organization)
		if err != nil {
			return groupList, err
		}
		partId, err := uuid.Parse(group.Metadata.Partner)
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
				labels := make(map[string]string)
				labels["organization"] = grp.OrganizationId.String()
				labels["partner"] = grp.PartnerId.String()

				group.Metadata = &v3.Metadata{
					Name:         grp.Name,
					Description:  grp.Description,
					Id:           grp.ID.String(),
					Organization: grp.OrganizationId.String(),
					Partner:      grp.PartnerId.String(),
					Labels:       labels,
					ModifiedAt:   timestamppb.New(grp.ModifiedAt),
				}
				group.Spec = &userv3.GroupSpec{
					Type: grp.Type,
				}
				groups = append(groups, group)
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
