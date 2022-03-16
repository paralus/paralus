package service

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/RafaySystems/rcloud-base/internal/dao"
	"github.com/RafaySystems/rcloud-base/internal/models"
	"github.com/RafaySystems/rcloud-base/internal/persistence/provider/pg"
	authzv1 "github.com/RafaySystems/rcloud-base/proto/types/authz"
	v3 "github.com/RafaySystems/rcloud-base/proto/types/commonpb/v3"
	rolev3 "github.com/RafaySystems/rcloud-base/proto/types/rolepb/v3"
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
	// create role
	Create(context.Context, *rolev3.Role) (*rolev3.Role, error)
	// get role by id
	GetByID(context.Context, *rolev3.Role) (*rolev3.Role, error)
	// get role by name
	GetByName(context.Context, *rolev3.Role) (*rolev3.Role, error)
	// create or update role
	Update(context.Context, *rolev3.Role) (*rolev3.Role, error)
	// delete role
	Delete(context.Context, *rolev3.Role) (*rolev3.Role, error)
	// list roles
	List(context.Context, *rolev3.Role) (*rolev3.RoleList, error)
}

// roleService implements RoleService
type roleService struct {
	db  *bun.DB
	azc AuthzService
}

// NewRoleService return new role service
func NewRoleService(db *bun.DB, azc AuthzService) RoleService {
	return &roleService{db: db, azc: azc}
}

func (s *roleService) getPartnerOrganization(ctx context.Context, db bun.IDB, role *rolev3.Role) (uuid.UUID, uuid.UUID, error) {
	partner := role.GetMetadata().GetPartner()
	org := role.GetMetadata().GetOrganization()
	partnerId, err := pg.GetPartnerId(ctx, s.db, partner)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	organizationId, err := pg.GetOrganizationId(ctx, s.db, org)
	if err != nil {
		return partnerId, uuid.Nil, err
	}
	return partnerId, organizationId, nil

}

func (s *roleService) deleteRolePermissionMapping(ctx context.Context, db bun.IDB, rleId uuid.UUID, role *rolev3.Role) (*rolev3.Role, error) {
	err := pg.DeleteX(ctx, s.db, "resource_role_id", rleId, &models.ResourceRolePermission{})
	if err != nil {
		return &rolev3.Role{}, err
	}

	drpm := authzv1.FilteredRolePermissionMapping{Role: role.GetMetadata().GetName()}
	success, err := s.azc.DeleteRolePermissionMappings(ctx, &drpm)
	if err != nil {
		return &rolev3.Role{}, fmt.Errorf("unable to delete mapping from authz; %v", err)
	}
	if !success.Res {
		fmt.Println("No roles deleted") // TODO: maybe it should not return false?
	}

	return role, nil
}

func (s *roleService) createRolePermissionMapping(ctx context.Context, db bun.IDB, role *rolev3.Role, ids parsedIds) (*rolev3.Role, error) {
	perms := role.GetSpec().GetRolepermissions()

	var items []models.ResourceRolePermission
	for _, p := range perms {
		entity, err := pg.GetIdByName(ctx, s.db, p, &models.ResourcePermission{})
		if err != nil {
			return role, fmt.Errorf("unable to find role permission '%v'", p)
		}
		if prm, ok := entity.(*models.ResourcePermission); ok {
			items = append(items, models.ResourceRolePermission{
				ResourceRoleId:       ids.Id,
				ResourcePermissionId: prm.ID,
			})
		} else {
			return role, fmt.Errorf("unable to find role permission '%v'", p)
		}
	}
	if len(items) > 0 {
		_, err := pg.Create(ctx, s.db, &items)
		if err != nil {
			return role, err
		}

		crpm := authzv1.RolePermissionMappingList{
			RolePermissionMappingList: []*authzv1.RolePermissionMapping{{
				Role:       role.GetMetadata().GetName(),
				Permission: role.Spec.Rolepermissions,
			}},
		}
		success, err := s.azc.CreateRolePermissionMappings(ctx, &crpm)
		if err != nil || !success.Res {
			return role, fmt.Errorf("unable to create mapping in authz; %v", err)
		}
	}
	return role, nil
}

func (s *roleService) Create(ctx context.Context, role *rolev3.Role) (*rolev3.Role, error) {
	partnerId, organizationId, err := s.getPartnerOrganization(ctx, s.db, role)
	if err != nil {
		return nil, fmt.Errorf("unable to get partner and org id")
	}
	r, _ := pg.GetIdByNamePartnerOrg(ctx, s.db, role.GetMetadata().GetName(), uuid.NullUUID{UUID: partnerId, Valid: true}, uuid.NullUUID{UUID: organizationId, Valid: true}, &models.Role{})
	if r != nil {
		return nil, fmt.Errorf("role '%v' already exists", role.GetMetadata().GetName())
	}

	scope := role.GetSpec().GetScope()
	// since this is purely additional metadata at this point, we
	// can kinda treat it as optional, and so we are allowing empty
	// TODO: check if "" is valid
	if !contains([]string{"system", "organization", "project", ""}, strings.ToLower(scope)) {
		return nil, fmt.Errorf("unknown scope '%v'", scope)
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
		Scope:          strings.ToLower(scope),
	}

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return &rolev3.Role{}, err
	}

	entity, err := pg.Create(ctx, tx, &rle)
	if err != nil {
		tx.Rollback()
		return &rolev3.Role{}, err
	}

	//update v3 spec
	if createdRole, ok := entity.(*models.Role); ok {
		role, err = s.createRolePermissionMapping(ctx, tx, role, parsedIds{Id: createdRole.ID, Partner: partnerId, Organization: organizationId})
		if err != nil {
			tx.Rollback()
			return &rolev3.Role{}, err
		}
	} else {
		tx.Rollback()
		return &rolev3.Role{}, fmt.Errorf("unable to create role '%v'", role.GetMetadata().GetName())
	}

	err = tx.Commit()
	if err != nil {
		fmt.Println("unable to commit changes", err)
	}
	return role, nil

}

func (s *roleService) GetByID(ctx context.Context, role *rolev3.Role) (*rolev3.Role, error) {
	id := role.GetMetadata().GetId()
	uid, err := uuid.Parse(id)
	if err != nil {
		return &rolev3.Role{}, err
	}
	entity, err := pg.GetByID(ctx, s.db, uid, &models.Role{})
	if err != nil {
		return &rolev3.Role{}, err
	}

	if rle, ok := entity.(*models.Role); ok {
		role, err = s.toV3Role(ctx, s.db, role, rle)
		if err != nil {
			return &rolev3.Role{}, err
		}
		return role, nil
	}
	return role, nil

}

func (s *roleService) GetByName(ctx context.Context, role *rolev3.Role) (*rolev3.Role, error) {
	name := role.GetMetadata().GetName()
	partnerId, organizationId, err := s.getPartnerOrganization(ctx, s.db, role)
	if err != nil {
		return nil, fmt.Errorf("unable to get partner and org id")
	}
	entity, err := pg.GetByNamePartnerOrg(ctx, s.db, name, uuid.NullUUID{UUID: partnerId, Valid: true}, uuid.NullUUID{UUID: organizationId, Valid: true}, &models.Role{})
	if err != nil {
		return &rolev3.Role{}, err
	}

	if rle, ok := entity.(*models.Role); ok {
		role, err = s.toV3Role(ctx, s.db, role, rle)
		if err != nil {
			return &rolev3.Role{}, err
		}
	} else {
		return nil, fmt.Errorf("unable to find role")
	}
	return role, nil

}

func (s *roleService) Update(ctx context.Context, role *rolev3.Role) (*rolev3.Role, error) {
	partnerId, organizationId, err := s.getPartnerOrganization(ctx, s.db, role)
	if err != nil {
		return nil, fmt.Errorf("unable to get partner and org id")
	}

	name := role.GetMetadata().GetName()
	entity, err := pg.GetByNamePartnerOrg(ctx, s.db, name, uuid.NullUUID{UUID: partnerId, Valid: true}, uuid.NullUUID{UUID: organizationId, Valid: true}, &models.Role{})
	if err != nil {
		return role, fmt.Errorf("unable to find role '%v'", name)
	}

	if rle, ok := entity.(*models.Role); ok {
		//update role details
		rle.Name = role.Metadata.Name
		rle.Description = role.Metadata.Description
		rle.IsGlobal = role.Spec.IsGlobal
		rle.Scope = role.Spec.Scope
		rle.ModifiedAt = time.Now()

		tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
		if err != nil {
			return &rolev3.Role{}, err
		}

		_, err = pg.Update(ctx, tx, rle.ID, rle)
		if err != nil {
			tx.Rollback()
			return &rolev3.Role{}, err
		}

		role, err = s.deleteRolePermissionMapping(ctx, tx, rle.ID, role)
		if err != nil {
			tx.Rollback()
			return &rolev3.Role{}, err
		}

		role, err = s.createRolePermissionMapping(ctx, tx, role, parsedIds{Id: rle.ID, Partner: partnerId, Organization: organizationId})
		if err != nil {
			tx.Rollback()
			return &rolev3.Role{}, err
		}

		//update spec and status
		role.Spec = &rolev3.RoleSpec{
			IsGlobal: rle.IsGlobal,
			Scope:    rle.Scope,
		}

		err = tx.Commit()
		if err != nil {
			fmt.Println("unable to commit changes", err)
		}
		return role, nil
	}
	return &rolev3.Role{}, fmt.Errorf("unable to update role '%v'", role.GetMetadata().GetName())

}

func (s *roleService) Delete(ctx context.Context, role *rolev3.Role) (*rolev3.Role, error) {
	name := role.GetMetadata().GetName()
	partnerId, organizationId, err := s.getPartnerOrganization(ctx, s.db, role)
	if err != nil {
		return &rolev3.Role{}, fmt.Errorf("unable to get partner and org id; %v", err)
	}

	entity, err := pg.GetByNamePartnerOrg(ctx, s.db, name, uuid.NullUUID{UUID: partnerId, Valid: true}, uuid.NullUUID{UUID: organizationId, Valid: true}, &models.Role{})
	if err != nil {
		return &rolev3.Role{}, err
	}

	if rle, ok := entity.(*models.Role); ok {

		tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
		if err != nil {
			return &rolev3.Role{}, err
		}

		role, err = s.deleteRolePermissionMapping(ctx, tx, rle.ID, role)
		if err != nil {
			tx.Rollback()
			return &rolev3.Role{}, err
		}

		err = pg.Delete(ctx, s.db, rle.ID, rle)
		if err != nil {
			tx.Rollback()
			return &rolev3.Role{}, err
		}

		err = tx.Commit()
		if err != nil {
			fmt.Println("unable to commit changes", err)
		}
		return role, nil
	}

	return &rolev3.Role{}, fmt.Errorf("unable to delete role '%v'", role.GetMetadata().GetName())
}

func (s *roleService) toV3Role(ctx context.Context, db bun.IDB, role *rolev3.Role, rle *models.Role) (*rolev3.Role, error) {
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
	entities, err := dao.GetRolePermissions(ctx, s.db, rle.ID)
	if err != nil {
		return role, err
	}
	permissions := []string{}
	for _, p := range entities {
		permissions = append(permissions, p.Name)
	}

	role.Spec = &rolev3.RoleSpec{
		IsGlobal:        rle.IsGlobal,
		Scope:           rle.Scope,
		Rolepermissions: permissions,
	}
	return role, nil
}

func (s *roleService) List(ctx context.Context, role *rolev3.Role) (*rolev3.RoleList, error) {
	var roles []*rolev3.Role
	roleList := &rolev3.RoleList{
		ApiVersion: apiVersion,
		Kind:       roleListKind,
		Metadata: &v3.ListMetadata{
			Count: 0,
		},
	}
	if len(role.Metadata.Organization) > 0 {
		orgId, err := pg.GetOrganizationId(ctx, s.db, role.Metadata.Organization)
		if err != nil {
			return roleList, err
		}
		partId, err := pg.GetPartnerId(ctx, s.db, role.Metadata.Partner)
		if err != nil {
			return roleList, err
		}
		var rles []models.Role
		entities, err := pg.List(ctx, s.db, uuid.NullUUID{UUID: partId, Valid: true}, uuid.NullUUID{UUID: orgId, Valid: true}, &rles)
		if err != nil {
			return roleList, err
		}
		if rles, ok := entities.(*[]models.Role); ok {
			for _, rle := range *rles {
				entry := &rolev3.Role{Metadata: role.GetMetadata()}
				entry, err = s.toV3Role(ctx, s.db, entry, &rle)
				if err != nil {
					return roleList, err
				}
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
