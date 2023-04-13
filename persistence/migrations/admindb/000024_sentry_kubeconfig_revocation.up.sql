CREATE TABLE IF NOT EXISTS sentry_kubeconfig_revocation (
    id uuid NOT NULL default uuid_generate_v4(),
    organization_id uuid NOT NULL,
    partner_id uuid NOT NULL,
    account_id uuid NOT NULL,
    revoked_at timestamp WITH time zone,
    created_at timestamp WITH time zone NOT NULL,
    is_sso_user boolean default FALSE
);

ALTER TABLE ONLY sentry_kubeconfig_revocation ADD CONSTRAINT sentry_kubeconfig_revocation_pkey PRIMARY KEY (id);

ALTER TABLE ONLY sentry_kubeconfig_revocation
    ADD CONSTRAINT sentry_kubeconfig_revocation_acc_org_sso_key UNIQUE (organization_id, account_id, is_sso_user);