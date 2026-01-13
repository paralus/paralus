CREATE TABLE IF NOT EXISTS authsrv_resourcerolepermission (
    id uuid default uuid_generate_v4() PRIMARY KEY,
    name character varying(256) NOT NULL,
    description character varying(512) NOT NULL,
    created_at timestamp with time zone NOT NULL,
    modified_at timestamp with time zone NOT NULL,
    trash boolean NOT NULL,
    resource_permission_id uuid NOT NULL REFERENCES authsrv_resourcepermission(id) DEFERRABLE INITIALLY DEFERRED,
    resource_role_id uuid NOT NULL
);

CREATE INDEX IF NOT EXISTS authsrv_resourcerolepermission_name_a65794e7 ON authsrv_resourcerolepermission USING btree (name);

CREATE INDEX IF NOT EXISTS authsrv_resourcerolepermission_name_a65794e7_like ON authsrv_resourcerolepermission USING btree (name varchar_pattern_ops);

CREATE INDEX IF NOT EXISTS authsrv_resourcerolepermission_resource_permission_id_c076e909 ON authsrv_resourcerolepermission USING btree (resource_permission_id);

CREATE INDEX IF NOT EXISTS authsrv_resourcerolepermission_resource_role_id_054f52d8 ON authsrv_resourcerolepermission USING btree (resource_role_id);
