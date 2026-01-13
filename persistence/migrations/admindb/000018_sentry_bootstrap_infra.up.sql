CREATE TABLE IF NOT EXISTS sentry_bootstrap_infra (
    -- database id fields
    name character varying(256) PRIMARY KEY,
    organization_id uuid,
    partner_id uuid,
    project_id uuid,
    -- paralus meta fields
    display_name character varying(256) NOT NULL,
    created_at timestamp WITH time zone NOT NULL,
    modified_at timestamp WITH time zone,
    deleted_at timestamp with time zone,
    trash boolean NOT NULL default false,
    labels jsonb NOT NULL DEFAULT '{}' ::jsonb,
    annotations jsonb NOT NULL DEFAULT '{}' ::jsonb,
    -- infra spec
    ca_cert text NOT NULL,
    ca_key text NOT NULL,
    CONSTRAINT sentry_bootstrap_infra_name_partner_id_organization_id_proj_key UNIQUE (name, partner_id, organization_id, project_id)
);
