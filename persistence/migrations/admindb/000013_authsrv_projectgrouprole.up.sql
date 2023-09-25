CREATE TABLE IF NOT EXISTS authsrv_projectgrouprole (
    id uuid default uuid_generate_v4() PRIMARY KEY,
    name character varying(256) NOT NULL,
    description character varying(512) NOT NULL,
    created_at timestamp with time zone NOT NULL,
    modified_at timestamp with time zone NOT NULL,
    trash boolean NOT NULL,
    "default" boolean NOT NULL,
    active boolean NOT NULL,
    group_id uuid NOT NULL REFERENCES authsrv_group(id) DEFERRABLE INITIALLY DEFERRED,
    organization_id uuid REFERENCES authsrv_organization(id) DEFERRABLE INITIALLY DEFERRED,
    partner_id uuid REFERENCES authsrv_partner(id) DEFERRABLE INITIALLY DEFERRED,
    project_id uuid REFERENCES authsrv_project(id) DEFERRABLE INITIALLY DEFERRED,
    role_id uuid NOT NULL REFERENCES authsrv_resourcerole(id) DEFERRABLE INITIALLY DEFERRED
);

CREATE INDEX IF NOT EXISTS authsrv_projectgrouprole_group_id_bda11774 ON authsrv_projectgrouprole USING btree (group_id);

CREATE INDEX IF NOT EXISTS authsrv_projectgrouprole_name_34417538 ON authsrv_projectgrouprole USING btree (name);

CREATE INDEX IF NOT EXISTS authsrv_projectgrouprole_name_34417538_like ON authsrv_projectgrouprole USING btree (name varchar_pattern_ops);

CREATE INDEX IF NOT EXISTS authsrv_projectgrouprole_organization_id_f149c4f0 ON authsrv_projectgrouprole USING btree (organization_id);

CREATE INDEX IF NOT EXISTS authsrv_projectgrouprole_partner_id_72198047 ON authsrv_projectgrouprole USING btree (partner_id);

CREATE INDEX IF NOT EXISTS authsrv_projectgrouprole_project_id_5c5917b5 ON authsrv_projectgrouprole USING btree (project_id);

CREATE INDEX IF NOT EXISTS authsrv_projectgrouprole_role_id_d930456e ON authsrv_projectgrouprole USING btree (role_id);
