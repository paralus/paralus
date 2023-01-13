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
	cib := kclient.NewAdminCreateIdentityBody("default", traits)
	cib.Credentials = kclient.NewAdminIdentityImportCredentials()
	cib.Credentials.SetPassword(kclient.AdminCreateIdentityImportCredentialsPassword{
		Config: &kclient.AdminCreateIdentityImportCredentialsPasswordConfig{
			Password: kclient.PtrString(password),
		},
	})
	ipm := IdentityPublicMetadata{
		ForceReset: forceReset,
	}
	cib.SetMetadataPublic(ipm)
	ir, hr, err := k.kc.V0alpha2Api.AdminCreateIdentity(ctx).AdminCreateIdentityBody(*cib).Execute()
	if err != nil {
		fmt.Println(hr)
		return "", err
	}
	return ir.Id, nil
}

func (k *kratosAuthProvider) Update(ctx context.Context, id string, traits map[string]interface{}, forceReset bool) error {
	uib := kclient.NewAdminUpdateIdentityBody("default", "active", traits)
	ipm := IdentityPublicMetadata{
		ForceReset: forceReset,
	}
	uib.SetMetadataPublic(ipm)
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

func (k *kratosAuthProvider) GetPublicMetadata(ctx context.Context, id string) (*IdentityPublicMetadata, error) {
	identity, res, err := k.kc.V0alpha2Api.AdminGetIdentity(ctx, id).Execute()
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, errors.New("Failed to get identity")
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
