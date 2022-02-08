package service

import (
	"context"
	"fmt"
	"time"

	"github.com/RafaySystems/rcloud-base/components/common/pkg/persistence/provider/pg"
	"github.com/RafaySystems/rcloud-base/components/common/pkg/utils"
	v3 "github.com/RafaySystems/rcloud-base/components/common/proto/types/commonpb/v3"
	"github.com/RafaySystems/rcloud-base/components/usermgmt/internal/models"
	"github.com/RafaySystems/rcloud-base/components/usermgmt/internal/role/dao"
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
	Create(context.Context, *userv3.Role) (*userv3.Role, error)
	// get role by id
	GetByID(context.Context, *userv3.Role) (*userv3.Role, error)
	// get role by name
	GetByName(context.Context, *userv3.Role) (*userv3.Role, error)
	// create or update role
	Update(context.Context, *userv3.Role) (*userv3.Role, error)
	// delete role
	Delete(context.Context, *userv3.Role) (*userv3.Role, error)
	// list roles
	List(context.Context, *userv3.Role) (*userv3.RoleList, error)
}

// roleService implements RoleService
type roleService struct {
	dao  pg.EntityDAO
	rdao dao.RoleDAO
	l    utils.Lookup
}

// NewRoleService return new role service
func NewRoleService(db *bun.DB) RoleService {
	return &roleService{
		dao:  pg.NewEntityDAO(db),
		rdao: dao.NewRoleDAO(db),
		l:    utils.NewLookup(db),
	}
}

func (s *roleService) getPartnerOrganization(ctx context.Context, role *userv3.Role) (uuid.UUID, uuid.UUID, error) {
	partner := role.GetMetadata().GetPartner()
	org := role.GetMetadata().GetOrganization()
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

func (s *roleService) deleteRolePermissionMapping(ctx context.Context, rleId uuid.UUID, role *userv3.Role) (*userv3.Role, error) {
	err := s.dao.DeleteX(ctx, "resource_role_id", rleId, &models.ResourceRolePermission{})
	if err != nil {
		role.Status = statusFailed(err)
		return role, err
	}

	return role, nil
}

func (s *roleService) createRolePermissionMapping(ctx context.Context, role *userv3.Role, ids parsedIds) (*userv3.Role, error) {
	perms := role.GetSpec().GetRolepermissions()

	var items []models.ResourceRolePermission
	for _, p := range perms {
		entity, err := s.dao.GetIdByName(ctx, p, &models.ResourcePermission{})
		if err != nil {
			role.Status = statusFailed(fmt.Errorf("unable to find role permission '%v'", p))
			return role, fmt.Errorf("unable to find role permission '%v'", p)
		}
		if prm, ok := entity.(*models.ResourcePermission); ok {
			items = append(items, models.ResourceRolePermission{
				ResourceRoleId:       ids.Id,
				ResourcePermissionId: prm.ID,
			})
		} else {
			role.Status = statusFailed(fmt.Errorf("unable to find role permission '%v'", p))
			return role, fmt.Errorf("unable to find role permission '%v'", p)
		}
	}
	if len(items) > 0 {
		_, err := s.dao.Create(ctx, &items)
		if err != nil {
			return role, err
		}
	}
	return role, nil
}

func (s *roleService) Create(ctx context.Context, role *userv3.Role) (*userv3.Role, error) {
	partnerId, organizationId, err := s.getPartnerOrganization(ctx, role)
	if err != nil {
		return nil, fmt.Errorf("unable to get partner and org id")
	}
	r, _ := s.dao.GetIdByNamePartnerOrg(ctx, role.GetMetadata().GetName(), uuid.NullUUID{UUID: partnerId, Valid: true}, uuid.NullUUID{UUID: organizationId, Valid: true}, &models.Role{})
	if r != nil {
		return nil, fmt.Errorf("role '%v' already exists", role.GetMetadata().GetName())
	}

	// convert v3 spec to internal models
	rle := models.Role{
		Name:           role.GetMetadata().GetName(),
		Description:    role.GetMetadata().GetDescription(),
		CreatedAt:      time.Now(),
		ModifiedAt:     time.Now(),
		Trash:          false,
		OrganizationId: organizationId,
		PartnerId:      partnerId,
		IsGlobal:       role.GetSpec().GetIsGlobal(),
		Scope:          role.GetSpec().GetScope(), // TODO: validate scope is SYSTEM/ORG/PROJECT?
	}
	entity, err := s.dao.Create(ctx, &rle)
	if err != nil {
		role.Status = statusFailed(err)
		return role, err
	}

	//update v3 spec
	if createdRole, ok := entity.(*models.Role); ok {
		role, err = s.createRolePermissionMapping(ctx, role, parsedIds{Id: createdRole.ID, Partner: partnerId, Organization: organizationId})
		if err != nil {
			role.Status = statusFailed(err)
			return role, err
		}

		if role.Status == nil {
			role.Status = statusOK()
		}
	} else {
		role.Status = statusFailed(fmt.Errorf("unable to create role"))
	}

	return role, nil

}

func (s *roleService) GetByID(ctx context.Context, role *userv3.Role) (*userv3.Role, error) {
	id := role.GetMetadata().GetId()
	uid, err := uuid.Parse(id)
	if err != nil {
		role.Status = statusFailed(err)
		return role, err
	}
	entity, err := s.dao.GetByID(ctx, uid, &models.Role{})
	if err != nil {
		role.Status = statusFailed(err)
		return role, err
	}

	if rle, ok := entity.(*models.Role); ok {
		role, err = s.toV3Role(ctx, role, rle)
		if err != nil {
			role.Status = statusFailed(err)
			return role, err
		}
		return role, nil
	}
	return role, nil

}

func (s *roleService) GetByName(ctx context.Context, role *userv3.Role) (*userv3.Role, error) {
	name := role.GetMetadata().GetName()
	partnerId, organizationId, err := s.getPartnerOrganization(ctx, role)
	if err != nil {
		return nil, fmt.Errorf("unable to get partner and org id")
	}
	entity, err := s.dao.GetByNamePartnerOrg(ctx, name, uuid.NullUUID{UUID: partnerId, Valid: true}, uuid.NullUUID{UUID: organizationId, Valid: true}, &models.Role{})
	if err != nil {
		role.Status = statusFailed(err)
		return role, err
	}

	if rle, ok := entity.(*models.Role); ok {
		role, err = s.toV3Role(ctx, role, rle)
		if err != nil {
			role.Status = statusFailed(err)
			return role, nil
		}
		role.Status = statusOK()
		return role, nil
	}
	return role, nil

}

func (s *roleService) Update(ctx context.Context, role *userv3.Role) (*userv3.Role, error) {
	partnerId, organizationId, err := s.getPartnerOrganization(ctx, role)
	if err != nil {
		return nil, fmt.Errorf("unable to get partner and org id")
	}

	id, _ := uuid.Parse(role.Metadata.Id)
	entity, err := s.dao.GetByID(ctx, id, &models.Role{})
	if err != nil {
		role.Status = statusFailed(err)
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
			role.Status = statusFailed(err)
			return role, err
		}

		role, err = s.deleteRolePermissionMapping(ctx, rle.ID, role)
		if err != nil {
			role.Status = statusFailed(err)
			return role, err
		}
		role, err = s.createRolePermissionMapping(ctx, role, parsedIds{Id: id, Partner: partnerId, Organization: organizationId})
		if err != nil {
			role.Status = statusFailed(err)
			return role, err
		}

		//update spec and status
		role.Spec = &userv3.RoleSpec{
			IsGlobal: rle.IsGlobal,
			Scope:    rle.Scope,
		}
		role.Status = statusOK()
	}

	return role, nil
}

func (s *roleService) Delete(ctx context.Context, role *userv3.Role) (*userv3.Role, error) {
	name := role.GetMetadata().GetName()
	partnerId, organizationId, err := s.getPartnerOrganization(ctx, role)
	if err != nil {
		return role, fmt.Errorf("unable to get partner and org id")
	}

	entity, err := s.dao.GetByNamePartnerOrg(ctx, name, uuid.NullUUID{UUID: partnerId, Valid: true}, uuid.NullUUID{UUID: organizationId, Valid: true}, &models.Role{})
	if err != nil {
		role.Status = statusFailed(err)
		return role, err
	}
	if rle, ok := entity.(*models.Role); ok {
		role, err = s.deleteRolePermissionMapping(ctx, rle.ID, role)
		if err != nil {
			role.Status = statusFailed(err)
			return role, err
		}

		err = s.dao.Delete(ctx, rle.ID, rle)
		if err != nil {
			role.Status = statusFailed(err)
			return role, err
		}

		//update v3 spec
		role.Metadata.Name = rle.Name
		role.Status = statusOK()
	}

	return role, nil
}

func (s *roleService) toV3Role(ctx context.Context, role *userv3.Role, rle *models.Role) (*userv3.Role, error) {
	labels := make(map[string]string)
	labels["organization"] = role.GetMetadata().GetOrganization()
	labels["partner"] = role.GetMetadata().GetPartner()

	role.ApiVersion = apiVersion
	role.Kind = roleKind
	role.Metadata = &v3.Metadata{
		Name:         rle.Name,
		Description:  rle.Description,
		Organization: role.GetMetadata().GetOrganization(),
		Partner:      role.GetMetadata().GetPartner(),
		Labels:       labels,
		ModifiedAt:   timestamppb.New(rle.ModifiedAt),
	}
	entities, err := s.rdao.GetRolePermissions(ctx, rle.ID)
	if err != nil {
		return role, err
	}
	permissions := []string{}
	for _, p := range entities {
		permissions = append(permissions, p.Name)
	}

	role.Spec = &userv3.RoleSpec{
		IsGlobal:        rle.IsGlobal,
		Scope:           rle.Scope,
		Rolepermissions: permissions,
	}
	role.Status = statusOK()
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
		orgId, err := s.l.GetOrganizationId(ctx, role.Metadata.Organization)
		if err != nil {
			return roleList, err
		}
		partId, err := s.l.GetPartnerId(ctx, role.Metadata.Partner)
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
				entry := &userv3.Role{Metadata: role.GetMetadata()}
				entry, err = s.toV3Role(ctx, entry, &rle)
				roles = append(roles, entry)
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
