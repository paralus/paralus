CREATE TABLE IF NOT EXISTS authsrv_resourcerole (
    id uuid default uuid_generate_v4() PRIMARY KEY,
    name character varying(256) NOT NULL,
    description character varying(512) NOT NULL,
    created_at timestamp with time zone NOT NULL,
    modified_at timestamp with time zone NOT NULL,
    trash boolean NOT NULL,
    is_global boolean NOT NULL,
    builtin boolean NOT NULL,
    scope character varying(256) NOT NULL,
    organization_id uuid REFERENCES authsrv_organization(id) DEFERRABLE INITIALLY DEFERRED,
    partner_id uuid REFERENCES authsrv_partner(id) DEFERRABLE INITIALLY DEFERRED
);

CREATE INDEX IF NOT EXISTS authsrv_resourcerole_name_a93b875a ON authsrv_resourcerole USING btree (name);

CREATE INDEX IF NOT EXISTS authsrv_resourcerole_name_a93b875a_like ON authsrv_resourcerole USING btree (name varchar_pattern_ops);

CREATE INDEX IF NOT EXISTS authsrv_resourcerole_organization_id_9a0a7e7e ON authsrv_resourcerole USING btree (organization_id);

CREATE INDEX IF NOT EXISTS authsrv_resourcerole_partner_id_de49ca91 ON authsrv_resourcerole USING btree (partner_id);
