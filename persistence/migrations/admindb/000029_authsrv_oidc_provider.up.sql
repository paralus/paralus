CREATE TABLE IF NOT EXISTS authsrv_oidc_provider (
    id uuid default uuid_generate_v4() PRIMARY KEY,
    name character varying(256) NOT NULL,
    description character varying(512) NOT NULL,
    organization_id uuid REFERENCES authsrv_organization(id) DEFERRABLE INITIALLY DEFERRED,
    partner_id uuid NOT NULL REFERENCES authsrv_partner(id) DEFERRABLE INITIALLY DEFERRED,
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

CREATE UNIQUE INDEX IF NOT EXISTS authsrv_oidc_provider_issuer_url ON authsrv_oidc_provider (issuer_url) WHERE trash IS false;

CREATE UNIQUE INDEX IF NOT EXISTS authsrv_oidc_provider_name ON authsrv_oidc_provider (name) WHERE trash IS false;

CREATE INDEX IF NOT EXISTS authsrv_oidc_provider_organization_id_4219d6ee ON authsrv_oidc_provider USING btree (organization_id);

CREATE INDEX IF NOT EXISTS authsrv_oidc_provider_partner_id_beb7c8df ON authsrv_oidc_provider USING btree (partner_id);
