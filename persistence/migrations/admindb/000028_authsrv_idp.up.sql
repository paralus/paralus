CREATE TABLE IF NOT EXISTS authsrv_idp (
    id uuid NOT NULL default uuid_generate_v4(),
    name character varying(256) NOT NULL,
    description character varying(512) NOT NULL,
    created_at timestamp with time zone NOT NULL,
    modified_at timestamp with time zone NOT NULL,
    idp_name character varying(256) NOT NULL,
    domain character varying(64) NOT NULL,
    trash boolean NOT NULL,
    organization_id uuid,
    partner_id uuid NOT NULL,
    sso_url character varying(256),
    idp_cert text,
    sp_cert text,
    sp_key text,
    metadata_url character varying(256),
    metadata_filename character varying(64),
    metadata bytea,
    group_attribute_name text,
    is_sae_enabled boolean default false
);

ALTER TABLE authsrv_idp OWNER TO admindbuser;

ALTER TABLE ONLY authsrv_idp ADD CONSTRAINT authsrv_idp_pkey PRIMARY KEY (id);

ALTER TABLE ONLY authsrv_idp ADD CONSTRAINT authsrv_idp_partner_id_domain_5669b152_uniq UNIQUE (partner_id, domain);

CREATE INDEX authsrv_idp_organization_id_4219d6ee ON authsrv_idp USING btree (organization_id);

CREATE INDEX authsrv_idp_partner_id_beb7c8df ON authsrv_idp USING btree (partner_id);

CREATE INDEX authsrv_idp_partner_id_domain_5669b152_idx ON authsrv_idp USING btree (partner_id, domain);

ALTER TABLE ONLY authsrv_idp
    ADD CONSTRAINT authsrv_idp_organization_id_4219d6ee_fk_authsrv_organization_id FOREIGN KEY (organization_id) 
    REFERENCES authsrv_organization(id) DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE ONLY authsrv_idp
    ADD CONSTRAINT authsrv_idp_partner_id_beb7c8df_fk_authsrv_partner_id FOREIGN KEY (partner_id) 
    REFERENCES authsrv_partner(id) DEFERRABLE INITIALLY DEFERRED;