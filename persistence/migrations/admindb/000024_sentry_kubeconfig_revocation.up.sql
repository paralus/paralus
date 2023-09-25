CREATE TABLE IF NOT EXISTS sentry_kubeconfig_revocation (
    id uuid default uuid_generate_v4() PRIMARY KEY,
    organization_id uuid NOT NULL,
    partner_id uuid NOT NULL,
    account_id uuid NOT NULL,
    revoked_at timestamp WITH time zone,
    created_at timestamp WITH time zone NOT NULL,
    is_sso_user boolean default FALSE,
    CONSTRAINT sentry_kubeconfig_revocation_acc_org_sso_key UNIQUE (organization_id, account_id, is_sso_user)
);
