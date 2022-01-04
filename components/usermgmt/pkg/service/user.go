package service

import (
	"context"
	"fmt"

	kclient "github.com/ory/kratos-client-go"

	v3 "github.com/RafaySystems/rcloud-base/components/common/proto/types/commonpb/v3"
	userrpcv3 "github.com/RafaySystems/rcloud-base/components/usermgmt/proto/rpc/v3"
	userv3 "github.com/RafaySystems/rcloud-base/components/usermgmt/proto/types/userpb/v3"
)

type server struct {
	kc *kclient.APIClient
}

func identityToUser(id *kclient.Identity) *userv3.User {
	traits := id.Traits.(map[string]interface{})
	return &userv3.User{
		ApiVersion: "usermgmt.k8smgmt.io/v3",
		Kind:       "User",
		Metadata: &v3.Metadata{
			Id: id.Id,
		},
		Spec: &userv3.UserSpec{
			Username: traits["email"].(string),
			FirstName: traits["first_name"].(string),
			LastName: traits["last_name"].(string),
		},
	}
}

func (s *server) CreateUser(ctx context.Context, user *userv3.User) (*userv3.User, error) {
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
	fmt.Println("Recovery link:", rl.RecoveryLink) // TODO: email the recovery link to the user
	return identityToUser(ir), nil
}
func (s *server) GetUsers(ctx context.Context, _ *userrpcv3.Empty) (*userv3.UserList, error) {
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
func (s *server) GetUser(ctx context.Context, req *userrpcv3.UserRequest) (*userv3.User, error) {
	// TODO: should it be get by id or by email? Kratos can only fileter by id
	ir, _, err := s.kc.V0alpha2Api.AdminGetIdentity(ctx, req.Userid).Execute()
	if err != nil {
		return nil, err
	}
	return identityToUser(ir), nil
}
func (s *server) UpdateUser(ctx context.Context, req *userrpcv3.PutUserRequest) (*userrpcv3.UserResponse, error) {
	uib := kclient.NewAdminUpdateIdentityBody("active", map[string]interface{}{"email": req.User.Spec.Username, "first_name": req.User.Spec.FirstName, "last_name": req.User.Spec.LastName})
	_, hr, err := s.kc.V0alpha2Api.AdminUpdateIdentity(ctx, req.Userid).AdminUpdateIdentityBody(*uib).Execute()
	if err != nil {
		fmt.Println(hr)
		// TODO: forward exact error message from kratos (eg: json schema validation)
		return nil, err
	}
	return &userrpcv3.UserResponse{Status: "OK"}, nil
}
func (s *server) DeleteUser(ctx context.Context, req *userrpcv3.UserRequest) (*userrpcv3.UserResponse, error) {
	// TODO: should it be get by id or by email? Kratos can only fileter by id
	_, err := s.kc.V0alpha2Api.AdminDeleteIdentity(ctx, req.Userid).Execute()
	if err != nil {
		return nil, err
	}
	return &userrpcv3.UserResponse{Status: "OK"}, nil
}

func NewUserServer(kc *kclient.APIClient) userrpcv3.UserServer {
	return &server{kc: kc}
}
