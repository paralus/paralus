CREATE TABLE IF NOT EXISTS authsrv_oidc_provider (
    id uuid NOT NULL default uuid_generate_v4(),
    name character varying(256) NOT NULL,
    description character varying(512) NOT NULL,
    organization_id uuid,
    partner_id uuid NOT NULL,
    created_at timestamp with time zone NOT NULL,
    modified_at timestamp with time zone NOT NULL,

    provider_name character varying(256) NOT NULL,
    mapper_url character varying(256),
    mapper_filename character varying(256),
    client_id character varying(256) NOT NULL,
    client_secret character varying(256) NOT NULL,
    scopes text[] NOT NULL,
    issuer_url character varying(256) NOT NULL,
    auth_url character varying(256),
    token_url character varying(256),
    requested_claims jsonb NOT NULL DEFAULT '{}' ::jsonb,
    predefined boolean default false,
    trash boolean NOT NULL
);

ALTER TABLE authsrv_oidc_provider OWNER TO admindbuser;

ALTER TABLE ONLY authsrv_oidc_provider ADD CONSTRAINT authsrv_oidc_provider_pkey PRIMARY KEY (id);

CREATE UNIQUE index authsrv_oidc_provider_issuer_url ON authsrv_oidc_provider (issuer_url) WHERE trash IS false;

CREATE UNIQUE index authsrv_oidc_provider_name ON authsrv_oidc_provider (name) WHERE trash IS false;

CREATE INDEX authsrv_oidc_provider_organization_id_4219d6ee ON authsrv_oidc_provider USING btree (organization_id);

CREATE INDEX authsrv_oidc_provider_partner_id_beb7c8df ON authsrv_oidc_provider USING btree (partner_id);

ALTER TABLE ONLY authsrv_oidc_provider
    ADD CONSTRAINT authsrv_oidc_provider_organization_id_4219d6ee_fk_authsrv_organization_id FOREIGN KEY (organization_id) 
    REFERENCES authsrv_organization(id) DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE ONLY authsrv_oidc_provider
    ADD CONSTRAINT authsrv_oidc_provider_partner_id_beb7c8df_fk_authsrv_partner_id FOREIGN KEY (partner_id) 
    REFERENCES authsrv_partner(id) DEFERRABLE INITIALLY DEFERRED;
