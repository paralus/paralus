package providers

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	kclient "github.com/ory/kratos-client-go"
)

// IdentityPublicMetadata is an extra information of the
// user. Checkout
// https://www.ory.sh/docs/kratos/manage-identities/managing-users-identities-metadata
// for more information about Ory Kratos identity metadata.
type IdentityPublicMetadata struct {
	// Indicate identity is created with auto generated password.
	ForceReset bool
}

type kratosAuthProvider struct {
	kc *kclient.APIClient
}
type AuthProvider interface {
	// create new user
	Create(context.Context, string, map[string]interface{}, bool) (string, error) // returns id,error
	// update user
	Update(context.Context, string, map[string]interface{}, bool) error
	// get recovery link for user
	GetRecoveryLink(context.Context, string) (string, error)
	// delete user
	Delete(context.Context, string) error
	// Get Public metadata of Kratos id.
	GetPublicMetadata(context.Context, string) (*IdentityPublicMetadata, error)
}

func NewKratosAuthProvider(kc *kclient.APIClient) AuthProvider {
	return &kratosAuthProvider{kc: kc}
}

func (k *kratosAuthProvider) Create(ctx context.Context, password string, traits map[string]interface{}, forceReset bool) (string, error) {
	cib := kclient.NewCreateIdentityBody("default", traits)

	cib.Credentials = kclient.NewIdentityWithCredentials()
	cib.Credentials.SetPassword(kclient.IdentityWithCredentialsPassword{
		Config: &kclient.IdentityWithCredentialsPasswordConfig{
			Password: kclient.PtrString(password),
		},
	})
	ipm := IdentityPublicMetadata{
		ForceReset: forceReset,
	}
	cib.SetMetadataPublic(ipm)
	ir, hr, err := k.kc.IdentityApi.CreateIdentity(ctx).CreateIdentityBody(*cib).Execute()
	if err != nil {
		fmt.Println(hr)
		return "", err
	}
	return ir.Id, nil
}

func (k *kratosAuthProvider) Update(ctx context.Context, id string, traits map[string]interface{}, forceReset bool) error {
	uib := kclient.NewUpdateIdentityBody("default", "active", traits)
	ipm := IdentityPublicMetadata{
		ForceReset: forceReset,
	}
	uib.SetMetadataPublic(ipm)

	_, hr, err := k.kc.IdentityApi.UpdateIdentity(ctx, id).UpdateIdentityBody(*uib).Execute()
	if err != nil {
		fmt.Println(hr)
	}
	return err
}

func (k *kratosAuthProvider) GetRecoveryLink(ctx context.Context, id string) (string, error) {
	rlb := kclient.NewCreateRecoveryLinkForIdentityBody(id)
	rl, _, err := k.kc.IdentityApi.CreateRecoveryLinkForIdentity(ctx).CreateRecoveryLinkForIdentityBody(*rlb).Execute()
	if err != nil {
		return "", err
	}
	return rl.RecoveryLink, nil
}

func (k *kratosAuthProvider) Delete(ctx context.Context, id string) error {
	hr, err := k.kc.IdentityApi.DeleteIdentity(ctx, id).Execute()
	if err != nil {
		fmt.Println(hr)
	}
	return err
}

func (k *kratosAuthProvider) GetPublicMetadata(ctx context.Context, id string) (*IdentityPublicMetadata, error) {
	identity, res, err := k.kc.IdentityApi.GetIdentity(ctx, id).Execute()
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, errors.New("failed to get identity")
	}
	ipm := &IdentityPublicMetadata{}
	if identity.HasMetadataPublic() {
		meta := identity.GetMetadataPublic()
		if m, ok := meta.(map[string]interface{}); ok {
			fr, ok := m["ForceReset"].(bool)
			if ok {
				ipm.ForceReset = fr
			}
		}
	}
	return ipm, nil
}
