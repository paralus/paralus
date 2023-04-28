CREATE TABLE IF NOT EXISTS authsrv_organization (
    id uuid default uuid_generate_v4() PRIMARY KEY,
    name character varying(256) NOT NULL,
    description character varying(512) NOT NULL,
    created_at timestamp with time zone NOT NULL,
    modified_at timestamp with time zone NOT NULL,
    trash boolean NOT NULL,
    settings jsonb NOT NULL,
    billing_address text NOT NULL,
    partner_id uuid NOT NULL REFERENCES authsrv_partner(id) DEFERRABLE INITIALLY DEFERRED,
    active boolean NOT NULL,
    approved boolean NOT NULL,
    type character varying(64) NOT NULL,
    address_line1 text NOT NULL,
    address_line2 text NOT NULL,
    city text NOT NULL,
    country text NOT NULL,
    phone text NOT NULL,
    state text NOT NULL,
    zipcode text NOT NULL,
    deleted_name character varying(256),
    is_private boolean,
    is_totp_enabled boolean NOT NULL,
    are_clusters_shared boolean NOT NULL,
    psps_enabled boolean default TRUE,
    custom_psps_enabled boolean,
    default_blueprints_enabled boolean default TRUE,
    referer character varying(30),
    CONSTRAINT authsrv_organization_name_partner_id_7d1113b9_uniq UNIQUE (name, partner_id)
);

CREATE INDEX IF NOT EXISTS authsrv_organization_name_23376e56 ON authsrv_organization USING btree (name);

CREATE INDEX IF NOT EXISTS authsrv_organization_name_23376e56_like ON authsrv_organization USING btree (name varchar_pattern_ops);

CREATE INDEX IF NOT EXISTS authsrv_organization_partner_id_7b55b579 ON authsrv_organization USING btree (partner_id);
