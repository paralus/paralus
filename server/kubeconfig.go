package server

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/paralus/paralus/internal/constants"
	"github.com/paralus/paralus/pkg/query"
	"github.com/paralus/paralus/pkg/sentry/cryptoutil"
	"github.com/paralus/paralus/pkg/sentry/kubeconfig"
	"github.com/paralus/paralus/pkg/sentry/util"
	"github.com/paralus/paralus/pkg/service"
	sentryrpc "github.com/paralus/paralus/proto/rpc/sentry"
	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	sentry "github.com/paralus/paralus/proto/types/sentry"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type kubeConfigServer struct {
	bs  service.BootstrapService
	aps service.AccountPermissionService
	gps service.GroupPermissionService
	kss service.KubeconfigSettingService
	krs service.KubeconfigRevocationService
	pf  cryptoutil.PasswordFunc
	ks  service.ApiKeyService
	os  service.OrganizationService
	ps  service.PartnerService
	al  *zap.Logger
}

var _ sentryrpc.KubeConfigServiceServer = (*kubeConfigServer)(nil)

func (s *kubeConfigServer) GetForClusterSystemSession(ctx context.Context, in *sentryrpc.GetForClusterRequest) (*commonv3.HttpBody, error) {
	config, err := kubeconfig.GetConfigForCluster(ctx, s.bs, in, s.pf, s.kss, kubeconfig.ParalusSystem)
	if err != nil {
		return nil, err
	}
	return &commonv3.HttpBody{
		ContentType: "application/yaml",
		Data:        config,
	}, nil
}

func (s *kubeConfigServer) GetForClusterWebSession(ctx context.Context, in *sentryrpc.GetForClusterRequest) (*commonv3.HttpBody, error) {
	config, err := kubeconfig.GetConfigForCluster(ctx, s.bs, in, s.pf, s.kss, kubeconfig.WebShell)
	if err != nil {
		return nil, err
	}
	return &commonv3.HttpBody{
		ContentType: "application/yaml",
		Data:        config,
	}, nil
}

func (s *kubeConfigServer) GetForUser(ctx context.Context, in *sentryrpc.GetForUserRequest) (*commonv3.HttpBody, error) {
	config, err := kubeconfig.GetConfigForUser(ctx, s.bs, s.aps, s.gps, in, s.pf, s.kss, s.ks, s.os, s.ps, s.al)
	if err != nil {
		_log.Errorw("error generating kubeconfig", "error", err.Error())
		return nil, err
	}
	return &commonv3.HttpBody{
		ContentType: "application/yaml",
		Data:        config,
	}, nil
}

func (s *kubeConfigServer) RevokeKubeconfig(ctx context.Context, req *sentryrpc.RevokeKubeconfigRequest) (*sentryrpc.RevokeKubeconfigResponse, error) {
	opts := req.Opts
	accountID, err := query.GetAccountID(opts)
	if err != nil {
		return nil, err
	}
	isSSOUser := false

	// if no user scope in url revoke for current user
	if opts.UrlScope == "" {
		isSSOUser = opts.IsSSOUser
	}
	err = s.krs.Patch(ctx, &sentry.KubeconfigRevocation{
		OrganizationID: opts.Organization,
		PartnerID:      opts.Partner,
		AccountID:      accountID,
		IsSSOUser:      isSSOUser,
		RevokedAt:      timestamppb.New(time.Now()),
	})
	if err != nil {
		return nil, err
	}

	return &sentryrpc.RevokeKubeconfigResponse{}, nil
}

func (s *kubeConfigServer) GetOrganizationSetting(ctx context.Context, req *sentryrpc.GetKubeconfigSettingRequest) (*sentryrpc.GetKubeconfigSettingResponse, error) {
	opts := req.Opts
	orgID, err := util.GetOrganizationScope(opts.UrlScope)

	if err != nil {
		return nil, err
	}
	if orgID != opts.Organization {
		opts.Organization = orgID
	}
	ks, err := s.kss.Get(ctx, opts.Organization, "", false)
	if err == constants.ErrNotFound {
		// default values for ValiditySeconds and SaValiditySeconds: 8 hours
		return &sentryrpc.GetKubeconfigSettingResponse{ValiditySeconds: 28800, SaValiditySeconds: 28800}, nil
	} else if err != nil {
		return nil, err
	}

	resp := &sentryrpc.GetKubeconfigSettingResponse{
		ValiditySeconds:             ks.ValiditySeconds,
		SaValiditySeconds:           ks.SaValiditySeconds,
		EnableSessionCheck:          ks.EnableSessionCheck,
		EnablePrivateRelay:          ks.EnablePrivateRelay,
		EnforceOrgAdminSecretAccess: ks.EnforceOrgAdminSecretAccess,
		DisableWebKubectl:           ks.DisableWebKubectl,
		DisableCLIKubectl:           ks.DisableCLIKubectl,
	}
	return resp, nil
}

func (s *kubeConfigServer) GetUserSetting(ctx context.Context, req *sentryrpc.GetKubeconfigSettingRequest) (*sentryrpc.GetKubeconfigSettingResponse, error) {
	opts := req.Opts
	accountID, err := util.GetUserScope(opts.UrlScope)
	if err != nil {
		return nil, err
	}
	ks, err := s.kss.Get(ctx, opts.Organization, accountID, false)
	if err == constants.ErrNotFound {
		req.Opts.UrlScope = "organization/" + opts.Organization
		return s.GetOrganizationSetting(ctx, req)
	} else if err != nil {
		return nil, err
	}
	resp := &sentryrpc.GetKubeconfigSettingResponse{
		ValiditySeconds:             ks.ValiditySeconds,
		EnableSessionCheck:          ks.EnableSessionCheck,
		EnablePrivateRelay:          ks.EnablePrivateRelay,
		EnforceOrgAdminSecretAccess: ks.EnforceOrgAdminSecretAccess,
		DisableWebKubectl:           ks.DisableWebKubectl,
		DisableCLIKubectl:           ks.DisableCLIKubectl,
	}
	return resp, nil

}

func (s *kubeConfigServer) UpdateOrganizationSetting(ctx context.Context, req *sentryrpc.UpdateKubeconfigSettingRequest) (*sentryrpc.UpdateKubeconfigSettingResponse, error) {
	opts := req.Opts
	orgID, err := util.GetOrganizationScope(opts.UrlScope)
	if err != nil {
		return nil, err
	}
	if orgID != opts.Organization {
		return nil, fmt.Errorf("invalid request")
	}

	err = s.kss.Patch(ctx, &sentry.KubeconfigSetting{
		OrganizationID:              opts.Organization,
		PartnerID:                   opts.Partner,
		AccountID:                   "",
		ValiditySeconds:             req.ValiditySeconds,
		SaValiditySeconds:           req.SaValiditySeconds,
		EnableSessionCheck:          req.EnableSessionCheck,
		EnablePrivateRelay:          req.EnablePrivateRelay,
		EnforceOrgAdminSecretAccess: req.EnforceOrgAdminSecretAccess,
		DisableWebKubectl:           req.DisableWebKubectl,
		DisableCLIKubectl:           req.DisableCLIKubectl,
	})
	if err != nil {
		return nil, err
	}

	return &sentryrpc.UpdateKubeconfigSettingResponse{}, nil
}

func (s *kubeConfigServer) UpdateUserSetting(ctx context.Context, req *sentryrpc.UpdateKubeconfigSettingRequest) (*sentryrpc.UpdateKubeconfigSettingResponse, error) {
	opts := req.Opts
	accountID, err := util.GetUserScope(opts.UrlScope)
	if err != nil {
		return nil, err
	}
	_log.Infow("UpdateUserSetting", "req.EnforceOrgAdminSecretAccess", req.EnforceOrgAdminSecretAccess)
	_log.Infow("UpdateUserSetting", "req.DisableWebKubectl", req.DisableWebKubectl)

	err = s.kss.Patch(ctx, &sentry.KubeconfigSetting{
		OrganizationID:              opts.Organization,
		PartnerID:                   opts.Partner,
		AccountID:                   accountID,
		ValiditySeconds:             req.ValiditySeconds,
		EnableSessionCheck:          req.EnableSessionCheck,
		IsSSOUser:                   false,
		EnforceOrgAdminSecretAccess: req.EnforceOrgAdminSecretAccess,
		DisableWebKubectl:           req.DisableWebKubectl,
		DisableCLIKubectl:           req.DisableCLIKubectl,
	})
	if err != nil {
		return nil, err
	}

	return &sentryrpc.UpdateKubeconfigSettingResponse{}, nil
}

// NewKubeConfigServer returns new kube config server
func NewKubeConfigServer(bs service.BootstrapService, aps service.AccountPermissionService, gps service.GroupPermissionService, kss service.KubeconfigSettingService,
	krs service.KubeconfigRevocationService, pf cryptoutil.PasswordFunc, ksvc service.ApiKeyService, os service.OrganizationService, ps service.PartnerService, al *zap.Logger) sentryrpc.KubeConfigServiceServer {
	return &kubeConfigServer{bs, aps, gps, kss, krs, pf, ksvc, os, ps, al}
}

func checkOrgAdmin(groups []string) bool {
	orgGrp := "Organization Admins"
	sort.Strings(groups)
	indx := sort.SearchStrings(groups, orgGrp)
	if indx < len(groups) {
		if groups[indx] == orgGrp {
			return true
		}
	}
	return false
}

func (s *kubeConfigServer) RevokeKubeconfigSSO(ctx context.Context, req *sentryrpc.RevokeKubeconfigRequest) (*sentryrpc.RevokeKubeconfigResponse, error) {
	opts := req.Opts
	accountID, err := query.GetAccountID(opts)
	if err != nil {
		return nil, err
	}
	err = s.krs.Patch(ctx, &sentry.KubeconfigRevocation{
		OrganizationID: opts.Organization,
		PartnerID:      opts.Partner,
		AccountID:      accountID,
		IsSSOUser:      true,
		RevokedAt:      timestamppb.New(time.Now()),
	})
	if err != nil {
		return nil, err
	}

	/*TODO: pending with events
	revokeUser, err := kubeconfig.GetUserNameFromAccountID(ctx, accountID, opts.Organization, s.aps, opts.IsSSOUser)
	acID := accountID
	partnerID := opts.Partner
	orgID := opts.Organization
	kubeconfigRevokeEvent(ctx, "user.kubeconfig.revoke", orgID, partnerID, revokeUser, acID, opts.Username, opts.Account.String(), opts.Groups)
	*/
	return &sentryrpc.RevokeKubeconfigResponse{}, nil
}
func (s *kubeConfigServer) GetSSOUserSetting(ctx context.Context, req *sentryrpc.GetKubeconfigSettingRequest) (*sentryrpc.GetKubeconfigSettingResponse, error) {
	opts := req.Opts
	accountID, err := util.GetUserScope(opts.UrlScope)
	if err != nil {
		return nil, err
	}
	ks, err := s.kss.Get(ctx, opts.Organization, accountID, true)
	if err == constants.ErrNotFound {
		req.Opts.UrlScope = "organization/" + opts.Organization
		return s.GetOrganizationSetting(ctx, req)
	} else if err != nil {
		return nil, err
	}
	resp := &sentryrpc.GetKubeconfigSettingResponse{
		ValiditySeconds:             ks.ValiditySeconds,
		EnableSessionCheck:          ks.EnableSessionCheck,
		EnablePrivateRelay:          ks.EnablePrivateRelay,
		EnforceOrgAdminSecretAccess: ks.EnforceOrgAdminSecretAccess,
		DisableWebKubectl:           ks.DisableWebKubectl,
		DisableCLIKubectl:           ks.DisableCLIKubectl,
	}
	return resp, nil
}

func (s *kubeConfigServer) UpdateSSOUserSetting(ctx context.Context, req *sentryrpc.UpdateKubeconfigSettingRequest) (*sentryrpc.UpdateKubeconfigSettingResponse, error) {
	opts := req.Opts
	accountID, err := util.GetUserScope(opts.UrlScope)
	if err != nil {
		return nil, err
	}
	_log.Infow("UpdateSSOUserSetting", "req.EnforceOrgAdminSecretAccess", req.EnforceOrgAdminSecretAccess)
	_log.Infow("UpdateSSOUserSetting", "req.DisableWebKubectl", req.DisableWebKubectl)

	err = s.kss.Patch(ctx, &sentry.KubeconfigSetting{
		OrganizationID:              opts.Organization,
		PartnerID:                   opts.Partner,
		AccountID:                   accountID,
		ValiditySeconds:             req.ValiditySeconds,
		EnableSessionCheck:          req.EnableSessionCheck,
		IsSSOUser:                   true,
		EnforceOrgAdminSecretAccess: req.EnforceOrgAdminSecretAccess,
		DisableWebKubectl:           req.DisableWebKubectl,
		DisableCLIKubectl:           req.DisableCLIKubectl,
	})
	if err != nil {
		return nil, err
	}

	/* TODO: pending with events
	forUser, err := kubeconfig.GetUserNameFromAccountID(ctx, accountID, opts.Organization, s.aps, opts.IsSSOUser)
	acID := accountID
	partnerID := opts.Partner
	orgIDString := opts.Organization
	kubeconfigSettingEvent(ctx, "user.kubeconfig.setting", orgIDString, partnerID, forUser, acID, opts.Username, opts.Account.String(), opts.Groups, req.ValiditySeconds, req.EnableSessionCheck)
	*/

	return &sentryrpc.UpdateKubeconfigSettingResponse{}, nil
}
