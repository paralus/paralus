CREATE TABLE IF NOT EXISTS authsrv_resourcepermission (
    id uuid default uuid_generate_v4() PRIMARY KEY,
    name character varying(256) NOT NULL,
    scope character varying(256) NOT NULL,
    base_url character varying(256) NOT NULL,
    description character varying(512) NOT NULL,
    created_at timestamp with time zone NOT NULL,
    modified_at timestamp with time zone NOT NULL,
    trash boolean NOT NULL,
    resource_urls jsonb NOT NULL,
    resource_action_urls jsonb NOT NULL
);

CREATE INDEX IF NOT EXISTS authsrv_resourcepermission_name_97f09d50 ON authsrv_resourcepermission USING btree (name);

CREATE INDEX IF NOT EXISTS authsrv_resourcepermission_name_97f09d50_like ON authsrv_resourcepermission USING btree (name varchar_pattern_ops);
