package authv3

import (
	"context"
	"strings"

	commonv3 "github.com/RafaySystems/rcloud-base/proto/types/commonpb/v3"
)

func (ac *authContext) IsRequestAllowed(ctx context.Context, req *commonv3.IsRequestAllowedRequest) (*commonv3.IsRequestAllowedResponse, error) {
	res := &commonv3.IsRequestAllowedResponse{
		Status:      commonv3.RequestStatus_Unknown,
		SessionData: &commonv3.SessionData{},
	}

	// Authenticate request
	err := ac.authenticate(ctx, req, res)
	if err != nil {
		return nil, err
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
func (ac *authContext) authenticate(ctx context.Context, req *commonv3.IsRequestAllowedRequest, res *commonv3.IsRequestAllowedResponse) error {
	tsr := ac.kc.V0alpha2Api.ToSession(ctx).
		XSessionToken(req.GetXSessionToken()).
		Cookie(req.GetCookie())
	session, _, err := ac.kc.V0alpha2Api.ToSessionExecute(tsr)
	if err != nil {
		// '401 Unauthorized' if the credentials are invalid or no credentials were sent.
		if strings.Contains(err.Error(), "401 Unauthorized") {
			res.Status = commonv3.RequestStatus_RequestNotAuthenticated
			res.Reason = "no or invalid credentials"
			return nil
		} else {
			return err
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
	return nil
}

// authorize performs authorization of the request and populate
// relevant information in res.
func (ac *authContext) authorize(ctx context.Context, req *commonv3.IsRequestAllowedRequest, res *commonv3.IsRequestAllowedResponse) error {
	return nil
}
