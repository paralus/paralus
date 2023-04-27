CREATE TABLE IF NOT EXISTS authsrv_accountresourcerole (
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
    role_id uuid NOT NULL REFERENCES authsrv_resourcerole(id) DEFERRABLE INITIALLY DEFERRED
);

CREATE INDEX IF NOT EXISTS authsrv_accountresourcerole_account_id_229069ae ON authsrv_accountresourcerole USING btree (account_id);

CREATE INDEX IF NOT EXISTS authsrv_accountresourcerole_name_e1f3d8a0 ON authsrv_accountresourcerole USING btree (name);

CREATE INDEX IF NOT EXISTS authsrv_accountresourcerole_name_e1f3d8a0_like ON authsrv_accountresourcerole USING btree (name varchar_pattern_ops);

CREATE INDEX IF NOT EXISTS authsrv_accountresourcerole_organization_id_22bb772c ON authsrv_accountresourcerole USING btree (organization_id);

CREATE INDEX IF NOT EXISTS authsrv_accountresourcerole_partner_id_8e96aff4 ON authsrv_accountresourcerole USING btree (partner_id);

CREATE INDEX IF NOT EXISTS authsrv_accountresourcerole_role_id_769ec143 ON authsrv_accountresourcerole USING btree (role_id);
