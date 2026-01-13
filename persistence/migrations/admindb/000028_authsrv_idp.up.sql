CREATE TABLE IF NOT EXISTS authsrv_idp (
    id uuid default uuid_generate_v4() PRIMARY KEY,
    name character varying(256) NOT NULL,
    description character varying(512) NOT NULL,
    created_at timestamp with time zone NOT NULL,
    modified_at timestamp with time zone NOT NULL,
    idp_name character varying(256) NOT NULL,
    domain character varying(64) NOT NULL,
    trash boolean NOT NULL,
    organization_id uuid REFERENCES authsrv_organization(id) DEFERRABLE INITIALLY DEFERRED,
    partner_id uuid NOT NULL REFERENCES authsrv_partner(id) DEFERRABLE INITIALLY DEFERRED,
    sso_url character varying(256),
    idp_cert text,
    sp_cert text,
    sp_key text,
    metadata_url character varying(256),
    metadata_filename character varying(64),
    metadata bytea,
    group_attribute_name text,
    is_sae_enabled boolean default false,
    CONSTRAINT authsrv_idp_partner_id_domain_5669b152_uniq UNIQUE (partner_id, domain)
);

CREATE INDEX IF NOT EXISTS authsrv_idp_organization_id_4219d6ee ON authsrv_idp USING btree (organization_id);

CREATE INDEX IF NOT EXISTS authsrv_idp_partner_id_beb7c8df ON authsrv_idp USING btree (partner_id);

CREATE INDEX IF NOT EXISTS authsrv_idp_partner_id_domain_5669b152_idx ON authsrv_idp USING btree (partner_id, domain);
