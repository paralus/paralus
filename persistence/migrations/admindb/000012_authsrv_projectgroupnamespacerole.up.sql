CREATE TABLE IF NOT EXISTS authsrv_projectgroupnamespacerole (
    id uuid NOT NULL default uuid_generate_v4(),
    name character varying(256) NOT NULL,
    description character varying(512) NOT NULL,
    created_at timestamp with time zone NOT NULL,
    modified_at timestamp with time zone NOT NULL,
    trash boolean NOT NULL,
    namespace_id character varying(64) NOT NULL,
    active boolean NOT NULL,
    group_id uuid NOT NULL,
    organization_id uuid,
    partner_id uuid,
    project_id uuid,
    role_id uuid NOT NULL
);

ALTER TABLE authsrv_projectgroupnamespacerole OWNER TO admindbuser;

ALTER TABLE ONLY authsrv_projectgroupnamespacerole ADD CONSTRAINT authsrv_projectgroupnamespacerole_pkey PRIMARY KEY (id);

CREATE INDEX authsrv_projectgroupnamespacerole_group_id_15ba5c48 ON authsrv_projectgroupnamespacerole USING btree (group_id);

CREATE INDEX authsrv_projectgroupnamespacerole_name_0d0cc737 ON authsrv_projectgroupnamespacerole USING btree (name);

CREATE INDEX authsrv_projectgroupnamespacerole_name_0d0cc737_like ON authsrv_projectgroupnamespacerole USING btree (name varchar_pattern_ops);

CREATE INDEX authsrv_projectgroupnamespacerole_organization_id_0b4626e3 ON authsrv_projectgroupnamespacerole USING btree (organization_id);

CREATE INDEX authsrv_projectgroupnamespacerole_partner_id_698d3c06 ON authsrv_projectgroupnamespacerole USING btree (partner_id);

CREATE INDEX authsrv_projectgroupnamespacerole_project_id_70668f98 ON authsrv_projectgroupnamespacerole USING btree (project_id);

CREATE INDEX authsrv_projectgroupnamespacerole_role_id_75ee38a5 ON authsrv_projectgroupnamespacerole USING btree (role_id);

ALTER TABLE ONLY authsrv_projectgroupnamespacerole
    ADD CONSTRAINT authsrv_projectgroup_group_id_15ba5c48_fk_authsrv_g FOREIGN KEY (group_id) 
    REFERENCES authsrv_group(id) DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE ONLY authsrv_projectgroupnamespacerole
    ADD CONSTRAINT authsrv_projectgroup_organization_id_0b4626e3_fk_authsrv_o FOREIGN KEY (organization_id) 
    REFERENCES authsrv_organization(id) DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE ONLY authsrv_projectgroupnamespacerole
    ADD CONSTRAINT authsrv_projectgroup_partner_id_698d3c06_fk_authsrv_p FOREIGN KEY (partner_id) 
    REFERENCES authsrv_partner(id) DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE ONLY authsrv_projectgroupnamespacerole
    ADD CONSTRAINT authsrv_projectgroup_project_id_70668f98_fk_authsrv_p FOREIGN KEY (project_id) 
    REFERENCES authsrv_project(id) DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE ONLY authsrv_projectgroupnamespacerole
    ADD CONSTRAINT authsrv_projectgroup_role_id_75ee38a5_fk_authsrv_r FOREIGN KEY (role_id) 
    REFERENCES authsrv_resourcerole(id) DEFERRABLE INITIALLY DEFERRED;