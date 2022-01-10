package service

import (
	"context"
	"fmt"

	kclient "github.com/ory/kratos-client-go"

	v3 "github.com/RafaySystems/rcloud-base/components/common/proto/types/commonpb/v3"
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
	kc *kclient.APIClient
}

func NewUserService(kc *kclient.APIClient) UserService {
	return &userService{kc: kc}
}

// Convert from kratos.Identity to GVK format
func identityToUser(id *kclient.Identity) *userv3.User {
	traits := id.Traits.(map[string]interface{})
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

func (s *userService) Create(ctx context.Context, user *userv3.User) (*userv3.User, error) {
	// TODO: restrict endpoint to admin
	cib := kclient.NewAdminCreateIdentityBody("default", map[string]interface{}{"email": user.Spec.Username, "first_name": user.Spec.FirstName, "last_name": user.Spec.LastName})
	ir, hr, err := s.kc.V0alpha2Api.AdminCreateIdentity(ctx).AdminCreateIdentityBody(*cib).Execute()
	if err != nil {
		fmt.Println(hr)
		// TODO: forward exact error message from kratos (eg: json schema validation)
		return nil, err
	}
	rlb := kclient.NewAdminCreateSelfServiceRecoveryLinkBody(ir.Id)
	rl, _, err := s.kc.V0alpha2Api.AdminCreateSelfServiceRecoveryLink(ctx).AdminCreateSelfServiceRecoveryLinkBody(*rlb).Execute()
	if err != nil {
		return nil, err
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
