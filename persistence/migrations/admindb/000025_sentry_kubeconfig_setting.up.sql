CREATE TABLE IF NOT EXISTS sentry_kubeconfig_setting (
    id uuid NOT NULL default uuid_generate_v4(),
    organization_id uuid NOT NULL,
    partner_id uuid NOT NULL,
    account_id uuid NOT NULL,
    scope character varying(256) NOT NULL,
    validity_seconds integer NOT NULL DEFAULT 0,
    sa_validity_seconds integer NOT NULL DEFAULT 0,
    created_at timestamp WITH time zone NOT NULL,
    modified_at timestamp WITH time zone,
    deleted_at timestamp WITH time zone,
    enforce_rsid boolean default false,
    disable_all_audit boolean default false,
    disable_cmd_audit boolean default false,
    is_sso_user boolean default false,
    disable_web_kubectl boolean default false,
    disable_cli_kubectl boolean default false,
    enable_privaterelay boolean default false,
    enforce_orgadmin_secret_access boolean default false
);

ALTER TABLE sentry_kubeconfig_setting OWNER TO admindbuser;

ALTER TABLE ONLY sentry_kubeconfig_setting ADD CONSTRAINT sentry_kubeconfig_setting_pkey PRIMARY KEY (id);

ALTER TABLE ONLY sentry_kubeconfig_setting
    ADD CONSTRAINT sentry_kubeconfig_setting_acc_org_sso_key UNIQUE (organization_id, account_id, is_sso_user);