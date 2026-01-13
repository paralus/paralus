CREATE TABLE IF NOT EXISTS authsrv_projectgroupnamespacerole (
    id uuid default uuid_generate_v4() PRIMARY KEY,
    name character varying(256) NOT NULL,
    description character varying(512) NOT NULL,
    created_at timestamp with time zone NOT NULL,
    modified_at timestamp with time zone NOT NULL,
    trash boolean NOT NULL,
    namespace character varying(64) NOT NULL,
    active boolean NOT NULL,
    group_id uuid NOT NULL REFERENCES authsrv_group(id) DEFERRABLE INITIALLY DEFERRED,
    organization_id uuid REFERENCES authsrv_organization(id) DEFERRABLE INITIALLY DEFERRED,
    partner_id uuid REFERENCES authsrv_partner(id) DEFERRABLE INITIALLY DEFERRED,
    project_id uuid REFERENCES authsrv_project(id) DEFERRABLE INITIALLY DEFERRED,
    role_id uuid NOT NULL REFERENCES authsrv_resourcerole(id) DEFERRABLE INITIALLY DEFERRED
);

CREATE INDEX IF NOT EXISTS authsrv_projectgroupnamespacerole_group_id_15ba5c48 ON authsrv_projectgroupnamespacerole USING btree (group_id);

CREATE INDEX IF NOT EXISTS authsrv_projectgroupnamespacerole_name_0d0cc737 ON authsrv_projectgroupnamespacerole USING btree (name);

CREATE INDEX IF NOT EXISTS authsrv_projectgroupnamespacerole_name_0d0cc737_like ON authsrv_projectgroupnamespacerole USING btree (name varchar_pattern_ops);

CREATE INDEX IF NOT EXISTS authsrv_projectgroupnamespacerole_organization_id_0b4626e3 ON authsrv_projectgroupnamespacerole USING btree (organization_id);

CREATE INDEX IF NOT EXISTS authsrv_projectgroupnamespacerole_partner_id_698d3c06 ON authsrv_projectgroupnamespacerole USING btree (partner_id);

CREATE INDEX IF NOT EXISTS authsrv_projectgroupnamespacerole_project_id_70668f98 ON authsrv_projectgroupnamespacerole USING btree (project_id);

CREATE INDEX IF NOT EXISTS authsrv_projectgroupnamespacerole_role_id_75ee38a5 ON authsrv_projectgroupnamespacerole USING btree (role_id);
