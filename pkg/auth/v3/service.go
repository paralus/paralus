package authv3

import (
	"context"
	"errors"
	"net/http"
	"strings"

	rpcv3 "github.com/RafayLabs/rcloud-base/proto/rpc/user"
	authzv1 "github.com/RafayLabs/rcloud-base/proto/types/authz"
	commonv3 "github.com/RafayLabs/rcloud-base/proto/types/commonpb/v3"
	"github.com/spacemonkeygo/httpsig"
)

var (
	// ErrInvalidAPIKey is returned when api key is invalid
	ErrInvalidAPIKey = errors.New("invalid api key")
	// ErrInvalidSignature is returns when signature is invalid
	ErrInvalidSignature = errors.New("invalid signature")
)

func (ac *authContext) IsRequestAllowed(ctx context.Context, httpreq *http.Request, req *commonv3.IsRequestAllowedRequest) (*commonv3.IsRequestAllowedResponse, error) {
	res := &commonv3.IsRequestAllowedResponse{
		Status:      commonv3.RequestStatus_Unknown,
		SessionData: &commonv3.SessionData{},
	}

	// Authenticate request
	err, succ := ac.authenticate(ctx, httpreq, req, res)
	if err != nil {
		return nil, err
	}
	// Don't bother checking authorization if athentication failed
	if !succ {
		return res, nil
	}

	// Authorize request
	err = ac.authorize(ctx, req, res)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// authenticate validate whether the request is from a legitimate user
// and populate relevant information in res.
func (ac *authContext) authenticate(ctx context.Context, httpreq *http.Request, req *commonv3.IsRequestAllowedRequest, res *commonv3.IsRequestAllowedResponse) (error, bool) {
	if len(req.XApiKey) > 0 && len(req.XSessionToken) == 0 {
		resp, err := ac.ks.GetByKey(ctx, &rpcv3.ApiKeyRequest{
			Id: req.XApiKey,
		})
		if err != nil {
			_log.Infow("unable to get api key", "key", req.XApiKey, "error", err)
			return ErrInvalidAPIKey, false
		}
		var kg httpsig.KeyGetterFunc = func(id string) interface{} {
			return []byte(resp.Secret)
		}

		verifier := httpsig.NewVerifier(kg)
		verifier.SetRequiredHeaders([]string{"content-md5", "date", "host", "nonce"})
		err = verifier.Verify(httpreq)
		if err != nil {
			return ErrInvalidSignature, false
		}
		res.Status = commonv3.RequestStatus_RequestAllowed
		res.SessionData.Username = resp.Name
		res.SessionData.Account = resp.AccountID.String()
	} else {
		tsr := ac.kc.V0alpha2Api.ToSession(ctx).
			XSessionToken(req.GetXSessionToken()).
			Cookie(req.GetCookie())
		session, _, err := ac.kc.V0alpha2Api.ToSessionExecute(tsr)
		if err != nil {
			// '401 Unauthorized' if the credentials are invalid or no credentials were sent.
			if strings.Contains(err.Error(), "401 Unauthorized") {
				res.Status = commonv3.RequestStatus_RequestNotAuthenticated
				res.Reason = "no or invalid credentials"
				return nil, false
			} else {
				return err, false
			}
		}
		if session.GetActive() {
			res.Status = commonv3.RequestStatus_RequestAllowed
			res.SessionData.Account = session.Identity.GetId()

			// TODO: Better way to access traits
			t := session.Identity.Traits.(map[string]interface{})
			res.SessionData.Username = t["email"].(string)
		} else {
			res.Status = commonv3.RequestStatus_RequestNotAuthenticated
			res.Reason = "no active session"
		}
	}
	return nil, true
}

// authorize performs authorization of the request
func (ac *authContext) authorize(ctx context.Context, req *commonv3.IsRequestAllowedRequest, res *commonv3.IsRequestAllowedResponse) error {
	// user,namespace,project,org,url(perm),method
	// ones that don't have value should be "*"
	proj := req.Project
	if proj == "" {
		proj = "*"
	}
	org := req.Org
	if org == "" {
		org = "*"
	}
	er := authzv1.EnforceRequest{
		Params: []string{"u:" + res.SessionData.Username, "*", proj, org, req.Url, req.Method},
	}
	authenticated, err := ac.as.Enforce(ctx, &er)

	if err != nil {
		return err
	}
	if !authenticated.Res {
		res.Status = commonv3.RequestStatus_RequestMethodOrURLNotAllowed
		res.Reason = "not authorized to perform action"
		return nil
	}
	return nil
}
