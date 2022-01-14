package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	kclient "github.com/ory/kratos-client-go"
	bun "github.com/uptrace/bun"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/RafaySystems/rcloud-base/components/common/pkg/persistence/provider/pg"
	v3 "github.com/RafaySystems/rcloud-base/components/common/proto/types/commonpb/v3"
	"github.com/RafaySystems/rcloud-base/components/usermgmt/pkg/internal/models"
	userrpcv3 "github.com/RafaySystems/rcloud-base/components/usermgmt/proto/rpc/v3"
	userv3 "github.com/RafaySystems/rcloud-base/components/usermgmt/proto/types/userpb/v3"
)

const (
	userKind     = "User"
	userListKind = "UserList"
)

// GroupService is the interface for group operations
type UserService interface {
	// create user
	Create(ctx context.Context, user *userv3.User) (*userv3.User, error)
	// get user by id
	GetByID(ctx context.Context, id string) (*userv3.User, error)
	// // get user by name
	// TODO: Implement GetByName
	// GetByName(ctx context.Context, name string) (*userv3.User, error)
	// create or update user
	Update(ctx context.Context, user *userv3.User) (*userv3.User, error)
	// delete user
	Delete(ctx context.Context, user *userv3.User) (*userrpcv3.DeleteUserResponse, error)
	// list users
	List(ctx context.Context, user *userv3.User) (*userv3.UserList, error)
}

type userService struct {
	kc  *kclient.APIClient
	dao pg.EntityDAO
}

func NewUserService(kc *kclient.APIClient, db *bun.DB) UserService {
	return &userService{kc: kc, dao: pg.NewEntityDAO(db)}
}

// Convert from kratos.Identity to GVK format
func identityToUser(id *kclient.Identity) *userv3.User {
	traits := id.GetTraits().(map[string]interface{})
	return &userv3.User{
		ApiVersion: "usermgmt.k8smgmt.io/v3",
		Kind:       "User",
		Metadata: &v3.Metadata{
			Id: id.Id,
		},
		Spec: &userv3.UserSpec{
			Username:  traits["email"].(string),
			FirstName: traits["first_name"].(string),
			LastName:  traits["last_name"].(string),
		},
	}
}

// Map roles to accounts
func (s *userService) updateUserRoleRelation(ctx context.Context, user *userv3.User) (*userv3.User, error) {
	accountId, _ := uuid.Parse(user.GetMetadata().GetId())
	partnerId, _ := uuid.Parse(user.GetMetadata().GetPartner())
	organizationId, _ := uuid.Parse(user.GetMetadata().GetOrganization())
	projectNamespaceRoles := user.GetSpec().GetProjectnamespaceroles()

	// TODO: add transactions
	var panrs []models.ProjectAccountNamespaceRole
	var pars []models.ProjectAccountResourcerole
	var ars []models.AccountResourcerole
	for _, pnr := range projectNamespaceRoles {
		projectId, perr := uuid.Parse(pnr.GetProject())
		namespaceId := pnr.GetNamespace()
		roleId, err := uuid.Parse(pnr.GetRole())
		if err != nil {
			return user, err
		}
		switch {
		case namespaceId != 0: // TODO: namespaceId can be zero?
			panr := models.ProjectAccountNamespaceRole{
				Name:           user.GetMetadata().GetName(),
				Description:    user.GetMetadata().GetDescription(),
				CreatedAt:      time.Now(),
				ModifiedAt:     time.Now(),
				Trash:          false,
				RoleId:         roleId,
				PartnerId:      partnerId,
				OrganizationId: organizationId,
				AccountId:      accountId,
				ProjectId:      projectId,
				NamesapceId:    namespaceId,
				Active:         true,
			}
			panrs = append(panrs, panr)
		case perr == nil: // TODO: maybe a better check?
			par := models.ProjectAccountResourcerole{
				Name:           user.GetMetadata().GetName(),
				Description:    user.GetMetadata().GetDescription(),
				CreatedAt:      time.Now(),
				ModifiedAt:     time.Now(),
				Trash:          false,
				Default:        true, // TODO: what is this for?
				RoleId:         roleId,
				PartnerId:      partnerId,
				OrganizationId: organizationId,
				AccountId:      accountId,
				ProjectId:      projectId,
				Active:         true,
			}
			pars = append(pars, par)
		default:
			ar := models.AccountResourcerole{
				Name:           user.GetMetadata().GetName(),
				Description:    user.GetMetadata().GetDescription(),
				CreatedAt:      time.Now(),
				ModifiedAt:     time.Now(),
				Trash:          false,
				Default:        true, // TODO: what is this for?
				RoleId:         roleId,
				PartnerId:      partnerId,
				OrganizationId: organizationId,
				AccountId:      accountId,
				Active:         true,
			}
			ars = append(ars, ar)
		}
	}
	if len(panrs) > 0 {
		_, err := s.dao.Create(ctx, &panrs)
		if err != nil {
			return user, err
		}
	}
	if len(pars) > 0 {
		_, err := s.dao.Create(ctx, &pars)
		if err != nil {
			return user, err
		}
	}
	if len(ars) > 0 {
		_, err := s.dao.Create(ctx, &ars)
		if err != nil {
			return user, err
		}
	}

	return user, nil
}

// Update the users(account) mapped to each group
func (s *userService) updateGroupAccountRelation(ctx context.Context, user *userv3.User) (*userv3.User, error) {
	// TODO: diff and delete the old relations
	userId, _ := uuid.Parse(user.GetMetadata().GetId())

	// TODO: add transactions
	var grpaccs []models.GroupAccount
	for _, group := range user.GetSpec().GetGroups() {
		groupId, err := uuid.Parse(group)
		if err != nil {
			return nil, err
		}
		grp := models.GroupAccount{
			Name:        user.GetMetadata().GetName(),        // TODO: what is name for relations?
			Description: user.GetMetadata().GetDescription(), // TODO: now sure what this is either
			CreatedAt:   time.Now(),
			ModifiedAt:  time.Now(),
			Trash:       false,
			AccountId:   userId,
			GroupId:     groupId,
			Active:      true,
		}
		grpaccs = append(grpaccs, grp)
	}
	if len(grpaccs) == 0 {
		return user, nil
	}
	_, err := s.dao.Create(ctx, &grpaccs)
	if err != nil {
		user.Status = &v3.Status{
			ConditionType:   "Create",
			ConditionStatus: v3.ConditionStatus_StatusFailed,
			LastUpdated:     timestamppb.Now(),
		}
		return user, err
	}

	return user, nil
}

func (s *userService) Create(ctx context.Context, user *userv3.User) (*userv3.User, error) {
	// TODO: restrict endpoint to admin
	cib := kclient.NewAdminCreateIdentityBody("default", map[string]interface{}{"email": user.Spec.Username, "first_name": user.Spec.FirstName, "last_name": user.Spec.LastName})
	ir, hr, err := s.kc.V0alpha2Api.AdminCreateIdentity(ctx).AdminCreateIdentityBody(*cib).Execute()
	if err != nil {
		fmt.Println(hr)
		// TODO: forward exact error message from kratos (eg: json schema validation)
		return nil, err
	}
	user.Metadata.Id = ir.Id
	rlb := kclient.NewAdminCreateSelfServiceRecoveryLinkBody(ir.Id)
	rl, _, err := s.kc.V0alpha2Api.AdminCreateSelfServiceRecoveryLink(ctx).AdminCreateSelfServiceRecoveryLinkBody(*rlb).Execute()
	if err != nil {
		return nil, err
	}

	user, err = s.updateUserRoleRelation(ctx, user)
	if err != nil {
		user.Status = &v3.Status{
			ConditionType:   "Create",
			ConditionStatus: v3.ConditionStatus_StatusFailed,
			LastUpdated:     timestamppb.Now(),
		}
		return user, err
	}

	user, err = s.updateGroupAccountRelation(ctx, user)
	if err != nil {
		user.Status = &v3.Status{
			ConditionType:   "Create",
			ConditionStatus: v3.ConditionStatus_StatusFailed,
			LastUpdated:     timestamppb.Now(),
		}
		return user, err
	}

	fmt.Println("Recovery link:", rl.RecoveryLink) // TODO: email the recovery link to the user
	user.Metadata = &v3.Metadata{
		Id: ir.Id,
	}
	user.Status = &v3.Status{
		ConditionType:   "StatusOK",
		ConditionStatus: v3.ConditionStatus_StatusOK,
	}

	return user, nil
}
func (s *userService) List(ctx context.Context, _ *userv3.User) (*userv3.UserList, error) {
	ir, _, err := s.kc.V0alpha2Api.AdminListIdentities(ctx).Execute()
	if err != nil {
		return nil, err
	}
	res := &userv3.UserList{}
	for _, u := range ir {
		res.Items = append(res.Items, identityToUser(&u))
	}
	return res, nil
}
func (s *userService) GetByID(ctx context.Context, id string) (*userv3.User, error) {
	// TODO: should it be get by id or by email? Kratos can only fileter by id
	ir, _, err := s.kc.V0alpha2Api.AdminGetIdentity(ctx, id).Execute()
	if err != nil {
		return nil, err
	}
	return identityToUser(ir), nil
}
func (s *userService) Update(ctx context.Context, user *userv3.User) (*userv3.User, error) {
	uib := kclient.NewAdminUpdateIdentityBody("active", map[string]interface{}{"email": user.Spec.Username, "first_name": user.Spec.FirstName, "last_name": user.Spec.LastName})
	_, hr, err := s.kc.V0alpha2Api.AdminUpdateIdentity(ctx, user.Metadata.Id).AdminUpdateIdentityBody(*uib).Execute()
	if err != nil {
		fmt.Println(hr)
		// TODO: forward exact error message from kratos (eg: json schema validation)
		return nil, err
	}
	user.Status = &v3.Status{
		ConditionType:   "StatusOK",
		ConditionStatus: v3.ConditionStatus_StatusOK,
	}

	return user, nil
}
func (s *userService) Delete(ctx context.Context, user *userv3.User) (*userrpcv3.DeleteUserResponse, error) {
	// TODO: should it be get by id or by email? Kratos can only filter by id
	_, err := s.kc.V0alpha2Api.AdminDeleteIdentity(ctx, user.Metadata.Id).Execute()
	if err != nil {
		return nil, err
	}

	return &userrpcv3.DeleteUserResponse{}, nil
}
