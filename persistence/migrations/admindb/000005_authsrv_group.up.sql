CREATE TABLE IF NOT EXISTS authsrv_group (
    id uuid default uuid_generate_v4() PRIMARY KEY,
    name character varying(256) NOT NULL,
    description character varying(512) NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modified_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    trash boolean NOT NULL DEFAULT FALSE,
    organization_id uuid NOT NULL REFERENCES authsrv_organization(id) DEFERRABLE INITIALLY DEFERRED,
    partner_id uuid NOT NULL REFERENCES authsrv_partner(id) DEFERRABLE INITIALLY DEFERRED,
    type character varying(64) NOT NULL
);

CREATE INDEX IF NOT EXISTS authsrv_group_name_d90b4524 ON authsrv_group USING btree (name);

CREATE INDEX IF NOT EXISTS authsrv_group_name_d90b4524_like ON authsrv_group USING btree (name varchar_pattern_ops);

CREATE INDEX IF NOT EXISTS authsrv_group_organization_id_e070e826 ON authsrv_group USING btree (organization_id);

CREATE INDEX IF NOT EXISTS authsrv_group_partner_id_1de9ab46 ON authsrv_group USING btree (partner_id);
