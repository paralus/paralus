CREATE TABLE IF NOT EXISTS authsrv_project (
    id uuid default uuid_generate_v4() PRIMARY KEY,
    name character varying(256) NOT NULL,
    description character varying(512) NOT NULL,
    created_at timestamp with time zone NOT NULL,
    modified_at timestamp with time zone NOT NULL,
    trash boolean NOT NULL,
    organization_id uuid REFERENCES authsrv_organization(id) DEFERRABLE INITIALLY DEFERRED,
    partner_id uuid REFERENCES authsrv_partner(id) DEFERRABLE INITIALLY DEFERRED,
    "default" boolean NOT NULL
);

-- update when we have more than one org
CREATE UNIQUE INDEX IF NOT EXISTS authsrv_project_unique_name ON authsrv_project (name) WHERE trash IS false;

CREATE INDEX IF NOT EXISTS authsrv_project_name_1b8dd279 ON authsrv_project USING btree (name);

CREATE INDEX IF NOT EXISTS authsrv_project_name_1b8dd279_like ON authsrv_project USING btree (name varchar_pattern_ops);

CREATE INDEX IF NOT EXISTS authsrv_project_organization_id_77437387 ON authsrv_project USING btree (organization_id);

CREATE INDEX IF NOT EXISTS authsrv_project_partner_id_3d505b76 ON authsrv_project USING btree (partner_id);
