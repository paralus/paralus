package providers

import (
	"context"
	"fmt"

	kclient "github.com/ory/kratos-client-go"
)

type kratosAuthProvider struct {
	kc *kclient.APIClient
}
type AuthProvider interface {
	// create new user
	Create(context.Context, map[string]interface{}) (string, error) // returns id,error
	// update user
	Update(context.Context, string, map[string]interface{}) error
	// get recovery link for user
	GetRecoveryLink(context.Context, string) (string, error)
	// delete user
	Delete(context.Context, string) error
}

func NewKratosAuthProvider(kc *kclient.APIClient) AuthProvider {
	return &kratosAuthProvider{kc: kc}
}

func (k *kratosAuthProvider) Create(ctx context.Context, traits map[string]interface{}) (string, error) {
	cib := kclient.NewAdminCreateIdentityBody("default", traits)
	ir, hr, err := k.kc.V0alpha2Api.AdminCreateIdentity(ctx).AdminCreateIdentityBody(*cib).Execute()
	if err != nil {
		fmt.Println(hr)
		return "", err
	}
	return ir.Id, nil
}

func (k *kratosAuthProvider) Update(ctx context.Context, id string, traits map[string]interface{}) error {
	uib := kclient.NewAdminUpdateIdentityBody("active", traits)
	_, hr, err := k.kc.V0alpha2Api.AdminUpdateIdentity(ctx, id).AdminUpdateIdentityBody(*uib).Execute()
	if err != nil {
		fmt.Println(hr)
	}
	return err
}

func (k *kratosAuthProvider) GetRecoveryLink(ctx context.Context, id string) (string, error) {
	rlb := kclient.NewAdminCreateSelfServiceRecoveryLinkBody(id)
	rl, _, err := k.kc.V0alpha2Api.AdminCreateSelfServiceRecoveryLink(ctx).AdminCreateSelfServiceRecoveryLinkBody(*rlb).Execute()
	if err != nil {
		return "", err
	}
	return rl.RecoveryLink, nil
}

func (k *kratosAuthProvider) Delete(ctx context.Context, id string) error {
	hr, err := k.kc.V0alpha2Api.AdminDeleteIdentity(ctx, id).Execute()
	if err != nil {
		fmt.Println(hr)
	}
	return err
}
