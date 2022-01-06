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
	roleKind     = "Role"
	roleListKind = "RoleList"
)

// RoleService is the interface for role operations
type RoleService interface {
	Close() error
	// create role
	Create(ctx context.Context, role *userv3.Role) (*userv3.Role, error)
	// get role by id
	GetByID(ctx context.Context, id string) (*userv3.Role, error)
	// get role by name
	GetByName(ctx context.Context, name string) (*userv3.Role, error)
	// create or update role
	Update(ctx context.Context, role *userv3.Role) (*userv3.Role, error)
	// delete role
	Delete(ctx context.Context, role *userv3.Role) (*userv3.Role, error)
	// list roles
	List(ctx context.Context, role *userv3.Role) (*userv3.RoleList, error)
}

// roleService implements RoleService
type roleService struct {
	dao pg.EntityDAO
}

// NewRoleService return new role service
func NewRoleService(db *bun.DB) RoleService {
	return &roleService{
		dao: pg.NewEntityDAO(db),
	}
}

func (s *roleService) Create(ctx context.Context, role *userv3.Role) (*userv3.Role, error) {

	partnerId, _ := uuid.Parse(role.GetMetadata().GetPartner())
	organizationId, _ := uuid.Parse(role.GetMetadata().GetOrganization())
	// TODO: find out the interaction if project key is present in the role metadata
	// TODO: check if a role with the same 'name' already exists and fail if so
	// TODO: we should be specifying names instead of ids for partner and org
	// TODO: create vs apply difference like in kubectl??
	//convert v3 spec to internal models
	grp := models.Role{
		Name:           role.GetMetadata().GetName(),
		Description:    role.GetMetadata().GetDescription(),
		CreatedAt:      time.Now(),
		ModifiedAt:     time.Now(),
		Trash:          false,
		OrganizationId: organizationId,
		PartnerId:      partnerId,
		IsGlobal:       role.GetSpec().GetIsGlobal(),
		Scope:          role.GetSpec().GetScope(),
	}
	entity, err := s.dao.Create(ctx, &grp)
	if err != nil {
		role.Status = &v3.Status{
			ConditionType:   "Create",
			ConditionStatus: v3.ConditionStatus_StatusFailed,
			LastUpdated:     timestamppb.Now(),
		}
		return role, err
	}

	//update v3 spec
	if createdRole, ok := entity.(*models.Role); ok {
		role.Metadata.Id = createdRole.ID.String()
		role.Spec = &userv3.RoleSpec{
			IsGlobal: createdRole.IsGlobal,
			Scope:    createdRole.Scope,
		}
		if role.Status != nil {
			role.Status = &v3.Status{
				ConditionType:   "Create",
				ConditionStatus: v3.ConditionStatus_StatusOK,
				LastUpdated:     timestamppb.Now(),
			}
		}
	}

	return role, nil

}

func (s *roleService) GetByID(ctx context.Context, id string) (*userv3.Role, error) {

	role := &userv3.Role{
		ApiVersion: apiVersion,
		Kind:       roleKind,
		Metadata: &v3.Metadata{
			Id: id,
		},
	}

	uid, err := uuid.Parse(id)
	if err != nil {
		role.Status = &v3.Status{
			ConditionType:   "Describe",
			ConditionStatus: v3.ConditionStatus_StatusFailed,
			LastUpdated:     timestamppb.Now(),
			Reason:          err.Error(),
		}
		return role, err
	}
	entity, err := s.dao.GetByID(ctx, uid, &models.Role{})
	if err != nil {
		role.Status = &v3.Status{
			ConditionType:   "Describe",
			ConditionStatus: v3.ConditionStatus_StatusFailed,
			LastUpdated:     timestamppb.Now(),
			Reason:          err.Error(),
		}
		return role, err
	}

	if grp, ok := entity.(*models.Role); ok {
		labels := make(map[string]string)
		labels["organization"] = grp.OrganizationId.String()

		role.Metadata = &v3.Metadata{
			Name:         grp.Name,
			Description:  grp.Description,
			Id:           grp.ID.String(),
			Organization: grp.OrganizationId.String(),
			Partner:      grp.PartnerId.String(),
			Labels:       labels,
			ModifiedAt:   timestamppb.New(grp.ModifiedAt),
		}
		role.Spec = &userv3.RoleSpec{
			IsGlobal: grp.IsGlobal,
			Scope:    grp.Scope,
		}
		role.Status = &v3.Status{
			LastUpdated:     timestamppb.Now(),
			ConditionType:   "Describe",
			ConditionStatus: v3.ConditionStatus_StatusOK,
		}

		return role, nil
	}
	return role, nil

}

func (s *roleService) GetByName(ctx context.Context, name string) (*userv3.Role, error) {
	fmt.Println("name:", name)

	role := &userv3.Role{
		ApiVersion: apiVersion,
		Kind:       roleKind,
		Metadata: &v3.Metadata{
			Name: name,
		},
	}

	entity, err := s.dao.GetByName(ctx, name, &models.Role{})
	if err != nil {
		role.Status = &v3.Status{
			ConditionType:   "Describe",
			ConditionStatus: v3.ConditionStatus_StatusFailed,
			LastUpdated:     timestamppb.Now(),
			Reason:          err.Error(),
		}
		return role, err
	}

	if grp, ok := entity.(*models.Role); ok {
		labels := make(map[string]string)
		labels["organization"] = grp.OrganizationId.String()

		role.Metadata = &v3.Metadata{
			Name:         grp.Name,
			Description:  grp.Description,
			Id:           grp.ID.String(),
			Organization: grp.OrganizationId.String(),
			Partner:      grp.PartnerId.String(),
			Labels:       labels,
			ModifiedAt:   timestamppb.New(grp.ModifiedAt),
		}
		role.Spec = &userv3.RoleSpec{
			IsGlobal: grp.IsGlobal,
			Scope:    grp.Scope,
		}
		role.Status = &v3.Status{
			LastUpdated:     timestamppb.Now(),
			ConditionType:   "Describe",
			ConditionStatus: v3.ConditionStatus_StatusOK,
		}

		return role, nil
	}
	return role, nil

}

func (s *roleService) Update(ctx context.Context, role *userv3.Role) (*userv3.Role, error) {
	// TODO: inform when unchanged

	id, _ := uuid.Parse(role.Metadata.Id)
	entity, err := s.dao.GetByID(ctx, id, &models.Role{})
	if err != nil {
		role.Status = &v3.Status{
			ConditionType:   "Update",
			ConditionStatus: v3.ConditionStatus_StatusFailed,
			LastUpdated:     timestamppb.Now(),
			Reason:          err.Error(),
		}
		return role, err
	}

	if grp, ok := entity.(*models.Role); ok {
		//update role details
		grp.Name = role.Metadata.Name
		grp.Description = role.Metadata.Description
		grp.IsGlobal = role.Spec.IsGlobal
		grp.Scope = role.Spec.Scope
		grp.ModifiedAt = time.Now()

		_, err = s.dao.Update(ctx, id, grp)
		if err != nil {
			role.Status = &v3.Status{
				ConditionType:   "Update",
				ConditionStatus: v3.ConditionStatus_StatusFailed,
				LastUpdated:     timestamppb.Now(),
				Reason:          err.Error(),
			}
			return role, err
		}

		//update spec and status
		role.Spec = &userv3.RoleSpec{
			IsGlobal: grp.IsGlobal,
			Scope:    grp.Scope,
		}
		role.Status = &v3.Status{
			ConditionType:   "Update",
			ConditionStatus: v3.ConditionStatus_StatusOK,
			LastUpdated:     timestamppb.Now(),
		}
	}

	return role, nil
}

func (s *roleService) Delete(ctx context.Context, role *userv3.Role) (*userv3.Role, error) {
	id, err := uuid.Parse(role.Metadata.Id)
	if err != nil {
		role.Status = &v3.Status{
			ConditionType:   "Delete",
			ConditionStatus: v3.ConditionStatus_StatusFailed,
			LastUpdated:     timestamppb.Now(),
			Reason:          err.Error(),
		}
		return role, err
	}
	entity, err := s.dao.GetByID(ctx, id, &models.Role{})
	if err != nil {
		role.Status = &v3.Status{
			ConditionType:   "Delete",
			ConditionStatus: v3.ConditionStatus_StatusFailed,
			LastUpdated:     timestamppb.Now(),
			Reason:          err.Error(),
		}
		return role, err
	}
	if grp, ok := entity.(*models.Role); ok {
		err = s.dao.Delete(ctx, id, grp)
		if err != nil {
			role.Status = &v3.Status{
				ConditionType:   "Delete",
				ConditionStatus: v3.ConditionStatus_StatusFailed,
				LastUpdated:     timestamppb.Now(),
				Reason:          err.Error(),
			}
			return role, err
		}
		//update v3 spec
		role.Metadata.Id = grp.ID.String()
		role.Metadata.Name = grp.Name
		role.Status = &v3.Status{
			ConditionType:   "Delete",
			ConditionStatus: v3.ConditionStatus_StatusOK,
			LastUpdated:     timestamppb.Now(),
		}
	}

	return role, nil
}

func (s *roleService) List(ctx context.Context, role *userv3.Role) (*userv3.RoleList, error) {

	var roles []*userv3.Role
	roleList := &userv3.RoleList{
		ApiVersion: apiVersion,
		Kind:       roleListKind,
		Metadata: &v3.ListMetadata{
			Count: 0,
		},
	}
	if len(role.Metadata.Organization) > 0 {
		orgId, err := uuid.Parse(role.Metadata.Organization)
		if err != nil {
			return roleList, err
		}
		partId, err := uuid.Parse(role.Metadata.Partner)
		if err != nil {
			return roleList, err
		}
		var grps []models.Role
		entities, err := s.dao.List(ctx, uuid.NullUUID{UUID: partId, Valid: true}, uuid.NullUUID{UUID: orgId, Valid: true}, &grps)
		if err != nil {
			return roleList, err
		}
		if grps, ok := entities.(*[]models.Role); ok {
			for _, grp := range *grps {
				labels := make(map[string]string)
				labels["organization"] = grp.OrganizationId.String()
				labels["partner"] = grp.PartnerId.String()

				role.Metadata = &v3.Metadata{
					Name:         grp.Name,
					Description:  grp.Description,
					Id:           grp.ID.String(),
					Organization: grp.OrganizationId.String(),
					Partner:      grp.PartnerId.String(),
					Labels:       labels,
					ModifiedAt:   timestamppb.New(grp.ModifiedAt),
				}
				role.Spec = &userv3.RoleSpec{
					// IsGlobal: grp.IsGlobal,
					Scope: grp.Scope,
				}
				roles = append(roles, role)
			}

			//update the list metadata and items response
			roleList.Metadata = &v3.ListMetadata{
				Count: int64(len(roles)),
			}
			roleList.Items = roles
		}

	} else {
		return roleList, fmt.Errorf("missing organization id in metadata")
	}
	return roleList, nil
}

func (s *roleService) Close() error {
	return s.dao.Close()
}
