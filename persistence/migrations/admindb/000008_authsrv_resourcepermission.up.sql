CREATE TABLE IF NOT EXISTS authsrv_resourcepermission (
    id integer NOT NULL,
    name character varying(256) NOT NULL,
    description character varying(512) NOT NULL,
    created_at timestamp with time zone NOT NULL,
    modified_at timestamp with time zone NOT NULL,
    trash boolean NOT NULL,
    resource_urls jsonb NOT NULL,
    resource_action_urls jsonb NOT NULL,
    organization_id integer,
    partner_id integer,
    resource_ref_id character varying(256) NOT NULL
);

ALTER TABLE authsrv_resourcepermission OWNER TO admindbuser;

CREATE SEQUENCE IF NOT EXISTS authsrv_resourcepermission_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER TABLE authsrv_resourcepermission_id_seq OWNER TO admindbuser;

ALTER SEQUENCE authsrv_resourcepermission_id_seq OWNED BY authsrv_resourcepermission.id;

ALTER TABLE ONLY authsrv_resourcepermission ALTER COLUMN id SET DEFAULT nextval('authsrv_resourcepermission_id_seq'::regclass);

ALTER TABLE ONLY authsrv_resourcepermission ADD CONSTRAINT authsrv_resourcepermission_pkey PRIMARY KEY (id);

CREATE INDEX authsrv_resourcepermission_name_97f09d50 ON authsrv_resourcepermission USING btree (name);

CREATE INDEX authsrv_resourcepermission_name_97f09d50_like ON authsrv_resourcepermission USING btree (name varchar_pattern_ops);

CREATE INDEX authsrv_resourcepermission_organization_id_daf7465e ON authsrv_resourcepermission USING btree (organization_id);

CREATE INDEX authsrv_resourcepermission_partner_id_f2ff9ad9 ON authsrv_resourcepermission USING btree (partner_id);

CREATE INDEX authsrv_resourcepermission_resource_ref_id_a47f8b94 ON authsrv_resourcepermission USING btree (resource_ref_id);

CREATE INDEX authsrv_resourcepermission_resource_ref_id_a47f8b94_like ON authsrv_resourcepermission USING btree (resource_ref_id varchar_pattern_ops);

ALTER TABLE ONLY authsrv_resourcepermission
    ADD CONSTRAINT authsrv_resourceperm_organization_id_daf7465e_fk_authsrv_o FOREIGN KEY (organization_id) 
    REFERENCES authsrv_organization(id) DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE ONLY authsrv_resourcepermission
    ADD CONSTRAINT authsrv_resourceperm_partner_id_f2ff9ad9_fk_authsrv_p FOREIGN KEY (partner_id) 
    REFERENCES authsrv_partner(id) DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE ONLY authsrv_resourcepermission
    ADD CONSTRAINT authsrv_resourceperm_resource_ref_id_a47f8b94_fk_authsrv_r FOREIGN KEY (resource_ref_id) 
    REFERENCES authsrv_resource(name) DEFERRABLE INITIALLY DEFERRED;