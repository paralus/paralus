CREATE TABLE IF NOT EXISTS sentry_bootstrap_agent_template (
    -- database id fields
    name character varying(256) PRIMARY KEY,
    organization_id uuid,
    partner_id uuid,
    project_id uuid,
    infra_ref character varying(256) NOT NULL REFERENCES sentry_bootstrap_infra(name),
    -- paralus meta fields
    display_name character varying(256) NOT NULL,
    created_at timestamp WITH time zone NOT NULL,
    modified_at timestamp WITH time zone,
    deleted_at timestamp with time zone,
    labels jsonb NOT NULL DEFAULT '{}' ::jsonb,
    annotations jsonb NOT NULL DEFAULT '{}' ::jsonb,
    trash boolean NOT NULL default false,
    -- template spec
    auto_register boolean NOT NULL DEFAULT FALSE,
    ignore_multiple_register boolean NOT NULL DEFAULT FALSE,
    auto_approve boolean NOT NULL DEFAULT FALSE,
    template_type character varying(512) NOT NULL,
    hosts jsonb NOT NULL DEFAULT '[]'::jsonb,
    token character varying(256) NOT NULL UNIQUE,
    incluster_template text NOT NULL,
    outofcluster_template text NOT NULL,
    CONSTRAINT sentry_bootstrap_agent_templa_name_partner_id_organization__key UNIQUE (name, partner_id, organization_id, project_id)
);
