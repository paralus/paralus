CREATE TABLE IF NOT EXISTS authsrv_projectgrouprole (
    id integer NOT NULL,
    name character varying(256) NOT NULL,
    description character varying(512) NOT NULL,
    created_at timestamp with time zone NOT NULL,
    modified_at timestamp with time zone NOT NULL,
    trash boolean NOT NULL,
    "default" boolean NOT NULL,
    active boolean NOT NULL,
    group_id integer NOT NULL,
    organization_id integer,
    partner_id integer,
    project_id integer,
    role_id integer NOT NULL
);

ALTER TABLE authsrv_projectgrouprole OWNER TO admindbuser;

CREATE SEQUENCE IF NOT EXISTS authsrv_projectgrouprole_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER TABLE authsrv_projectgrouprole_id_seq OWNER TO admindbuser;

ALTER SEQUENCE authsrv_projectgrouprole_id_seq OWNED BY authsrv_projectgrouprole.id;

ALTER TABLE ONLY authsrv_projectgrouprole ALTER COLUMN id SET DEFAULT nextval('authsrv_projectgrouprole_id_seq'::regclass);

ALTER TABLE ONLY authsrv_projectgrouprole ADD CONSTRAINT authsrv_projectgrouprole_pkey PRIMARY KEY (id);

CREATE INDEX authsrv_projectgrouprole_group_id_bda11774 ON authsrv_projectgrouprole USING btree (group_id);

CREATE INDEX authsrv_projectgrouprole_name_34417538 ON authsrv_projectgrouprole USING btree (name);

CREATE INDEX authsrv_projectgrouprole_name_34417538_like ON authsrv_projectgrouprole USING btree (name varchar_pattern_ops);

CREATE INDEX authsrv_projectgrouprole_organization_id_f149c4f0 ON authsrv_projectgrouprole USING btree (organization_id);

CREATE INDEX authsrv_projectgrouprole_partner_id_72198047 ON authsrv_projectgrouprole USING btree (partner_id);

CREATE INDEX authsrv_projectgrouprole_project_id_5c5917b5 ON authsrv_projectgrouprole USING btree (project_id);

CREATE INDEX authsrv_projectgrouprole_role_id_d930456e ON authsrv_projectgrouprole USING btree (role_id);

ALTER TABLE ONLY authsrv_projectgrouprole
    ADD CONSTRAINT authsrv_projectgroup_organization_id_f149c4f0_fk_authsrv_o FOREIGN KEY (organization_id) 
    REFERENCES authsrv_organization(id) DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE ONLY authsrv_projectgrouprole
    ADD CONSTRAINT authsrv_projectgroup_partner_id_72198047_fk_authsrv_p FOREIGN KEY (partner_id) 
    REFERENCES authsrv_partner(id) DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE ONLY authsrv_projectgrouprole
    ADD CONSTRAINT authsrv_projectgroup_project_id_5c5917b5_fk_authsrv_p FOREIGN KEY (project_id) 
    REFERENCES authsrv_project(id) DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE ONLY authsrv_projectgrouprole
    ADD CONSTRAINT authsrv_projectgroup_role_id_d930456e_fk_authsrv_r FOREIGN KEY (role_id) 
    REFERENCES authsrv_resourcerole(id) DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE ONLY authsrv_projectgrouprole
    ADD CONSTRAINT authsrv_projectgrouprole_group_id_bda11774_fk_authsrv_group_id FOREIGN KEY (group_id) 
    REFERENCES authsrv_group(id) DEFERRABLE INITIALLY DEFERRED;