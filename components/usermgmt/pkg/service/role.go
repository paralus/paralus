package service

import (
	"context"
	"fmt"
	"time"

	v3 "github.com/RafaySystems/rcloud-base/components/common/proto/types/commonpb/v3"
	"github.com/RafaySystems/rcloud-base/components/usermgmt/pkg/internal/models"
	"github.com/RafaySystems/rcloud-base/components/common/pkg/persistence/provider/pg"
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

// TODO: This right now just lets us create roles, we will have to make sure the mapping happens
func (s *roleService) Create(ctx context.Context, role *userv3.Role) (*userv3.Role, error) {

	partnerId, _ := uuid.Parse(role.GetMetadata().GetPartner())
	organizationId, _ := uuid.Parse(role.GetMetadata().GetOrganization())
	// TODO: check if a role with the same 'name' already exists and fail if so
	// TODO: we should be specifying names instead of ids for partner and org
	// TODO: create vs apply difference like in kubectl??
	//convert v3 spec to internal models
	rle := models.Role{
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
	entity, err := s.dao.Create(ctx, &rle)
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

	if rle, ok := entity.(*models.Role); ok {
		labels := make(map[string]string)
		labels["organization"] = rle.OrganizationId.String()

		role.Metadata = &v3.Metadata{
			Name:         rle.Name,
			Description:  rle.Description,
			Id:           rle.ID.String(),
			Organization: rle.OrganizationId.String(),
			Partner:      rle.PartnerId.String(),
			Labels:       labels,
			ModifiedAt:   timestamppb.New(rle.ModifiedAt),
		}
		role.Spec = &userv3.RoleSpec{
			IsGlobal: rle.IsGlobal,
			Scope:    rle.Scope,
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

	if rle, ok := entity.(*models.Role); ok {
		labels := make(map[string]string)
		labels["organization"] = rle.OrganizationId.String()

		role.Metadata = &v3.Metadata{
			Name:         rle.Name,
			Description:  rle.Description,
			Id:           rle.ID.String(),
			Organization: rle.OrganizationId.String(),
			Partner:      rle.PartnerId.String(),
			Labels:       labels,
			ModifiedAt:   timestamppb.New(rle.ModifiedAt),
		}
		role.Spec = &userv3.RoleSpec{
			IsGlobal: rle.IsGlobal,
			Scope:    rle.Scope,
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

	if rle, ok := entity.(*models.Role); ok {
		//update role details
		rle.Name = role.Metadata.Name
		rle.Description = role.Metadata.Description
		rle.IsGlobal = role.Spec.IsGlobal
		rle.Scope = role.Spec.Scope
		rle.ModifiedAt = time.Now()

		_, err = s.dao.Update(ctx, id, rle)
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
			IsGlobal: rle.IsGlobal,
			Scope:    rle.Scope,
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
	if rle, ok := entity.(*models.Role); ok {
		err = s.dao.Delete(ctx, id, rle)
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
		role.Metadata.Id = rle.ID.String()
		role.Metadata.Name = rle.Name
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
		var rles []models.Role
		entities, err := s.dao.List(ctx, uuid.NullUUID{UUID: partId, Valid: true}, uuid.NullUUID{UUID: orgId, Valid: true}, &rles)
		if err != nil {
			return roleList, err
		}
		if rles, ok := entities.(*[]models.Role); ok {
			for _, rle := range *rles {
				labels := make(map[string]string)
				labels["organization"] = rle.OrganizationId.String()
				labels["partner"] = rle.PartnerId.String()

				role.Metadata = &v3.Metadata{
					Name:         rle.Name,
					Description:  rle.Description,
					Id:           rle.ID.String(),
					Organization: rle.OrganizationId.String(),
					Partner:      rle.PartnerId.String(),
					Labels:       labels,
					ModifiedAt:   timestamppb.New(rle.ModifiedAt),
				}
				role.Spec = &userv3.RoleSpec{
					// IsGlobal: rle.IsGlobal,
					Scope: rle.Scope,
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
