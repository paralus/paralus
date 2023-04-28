CREATE TABLE IF NOT EXISTS authsrv_projectaccountresourcerole (
    id uuid default uuid_generate_v4() PRIMARY KEY,
    name character varying(256) NOT NULL,
    description character varying(512) NOT NULL,
    created_at timestamp with time zone NOT NULL,
    modified_at timestamp with time zone NOT NULL,
    trash boolean NOT NULL,
    "default" boolean NOT NULL,
    active boolean NOT NULL,
    account_id uuid NOT NULL,
    organization_id uuid REFERENCES authsrv_organization(id) DEFERRABLE INITIALLY DEFERRED,
    partner_id uuid REFERENCES authsrv_partner(id) DEFERRABLE INITIALLY DEFERRED,
    project_id uuid REFERENCES authsrv_project(id) DEFERRABLE INITIALLY DEFERRED,
    role_id uuid NOT NULL REFERENCES authsrv_resourcerole(id) DEFERRABLE INITIALLY DEFERRED
);

CREATE INDEX IF NOT EXISTS authsrv_projectaccountresourcerole_account_id_532ce8df ON authsrv_projectaccountresourcerole USING btree (account_id);

CREATE INDEX IF NOT EXISTS authsrv_projectaccountresourcerole_name_c4c3d60f ON authsrv_projectaccountresourcerole USING btree (name);

CREATE INDEX IF NOT EXISTS authsrv_projectaccountresourcerole_name_c4c3d60f_like ON authsrv_projectaccountresourcerole USING btree (name varchar_pattern_ops);

CREATE INDEX IF NOT EXISTS authsrv_projectaccountresourcerole_organization_id_91c5602d ON authsrv_projectaccountresourcerole USING btree (organization_id);

CREATE INDEX IF NOT EXISTS authsrv_projectaccountresourcerole_partner_id_81bde92c ON authsrv_projectaccountresourcerole USING btree (partner_id);

CREATE INDEX IF NOT EXISTS authsrv_projectaccountresourcerole_project_id_f8a43852 ON authsrv_projectaccountresourcerole USING btree (project_id);

CREATE INDEX IF NOT EXISTS authsrv_projectaccountresourcerole_role_id_a345b16f ON authsrv_projectaccountresourcerole USING btree (role_id);
