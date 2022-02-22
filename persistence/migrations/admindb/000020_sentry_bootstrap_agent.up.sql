CREATE TABLE IF NOT EXISTS sentry_bootstrap_agent (
    -- database id fields
    id uuid NOT NULL default uuid_generate_v4(),
    name character varying(256) NOT NULL,
    organization_id uuid,
    partner_id uuid,
    project_id uuid,
    template_ref character varying(256) NOT NULL,
    agent_mode character varying(512) NOT NULL,
    -- rafay meta fields
    display_name character varying(256) NOT NULL,
    created_at timestamp WITH time zone NOT NULL,
    modified_at timestamp WITH time zone,
    deleted_at timestamp with time zone,
    labels jsonb NOT NULL DEFAULT '{}' ::jsonb,
    annotations jsonb NOT NULL DEFAULT '{}' ::jsonb,
    -- bootstrap token spec fields
    token character varying(256) NOT NULL,
    -- bootstrap token status fields
    token_state character varying(256) NOT NULL,
    ip_address character varying(20) NOT NULL,
    last_checked_in timestamp with time zone,
    fingerprint character varying(256) NOT NULL
);

ALTER TABLE sentry_bootstrap_agent OWNER TO admindbuser;

ALTER TABLE ONLY sentry_bootstrap_agent ADD CONSTRAINT sentry_bootstrap_agent_pkey PRIMARY KEY (id);

ALTER TABLE ONLY sentry_bootstrap_agent
    ADD CONSTRAINT sentry_bootstrap_agent_name_templateref_organization_id_partner UNIQUE (name, template_ref, organization_id, partner_id, project_id);

ALTER TABLE ONLY sentry_bootstrap_agent
    ADD CONSTRAINT sentry_bootstrap_agent_token_key UNIQUE (token);

ALTER TABLE ONLY sentry_bootstrap_agent
    ADD CONSTRAINT sentry_bootstrap_agent_template_ref_fkey FOREIGN KEY (template_ref) 
    REFERENCES sentry_bootstrap_agent_template(name);