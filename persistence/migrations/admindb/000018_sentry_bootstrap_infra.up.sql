CREATE TABLE IF NOT EXISTS sentry_bootstrap_infra (
    -- database id fields
    name character varying(256) NOT NULL,
    organization_id uuid,
    partner_id uuid,
    project_id uuid,
    -- rafay meta fields
    display_name character varying(256) NOT NULL,
    created_at timestamp WITH time zone NOT NULL,
    modified_at timestamp WITH time zone,
    deleted_at timestamp with time zone,
    trash boolean NOT NULL default false,
    labels jsonb NOT NULL DEFAULT '{}' ::jsonb,
    annotations jsonb NOT NULL DEFAULT '{}' ::jsonb,
    -- infra spec
    ca_cert text NOT NULL,
    ca_key text NOT NULL
);

ALTER TABLE sentry_bootstrap_infra OWNER TO admindbuser;

ALTER TABLE ONLY sentry_bootstrap_infra ADD CONSTRAINT sentry_bootstrap_infra_pkey PRIMARY KEY (name);

ALTER TABLE ONLY sentry_bootstrap_infra
    ADD CONSTRAINT sentry_bootstrap_infra_name_partner_id_organization_id_proj_key UNIQUE (name, partner_id, organization_id, project_id);