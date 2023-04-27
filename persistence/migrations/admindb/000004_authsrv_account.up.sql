CREATE TABLE IF NOT EXISTS authsrv_ssoaccount (
    password character varying(128) NOT NULL,
    last_login timestamp with time zone,
    id uuid default uuid_generate_v4() PRIMARY KEY,
    name character varying(256) NOT NULL,
    organization_id uuid REFERENCES authsrv_organization(id) DEFERRABLE INITIALLY DEFERRED,
    description character varying(512) NOT NULL,
    created_at timestamp with time zone NOT NULL,
    modified_at timestamp with time zone NOT NULL,
    trash boolean NOT NULL,
    username character varying(256) NOT NULL UNIQUE,
    phone character varying(36) NOT NULL,
    first_name character varying(64) NOT NULL,
    last_name character varying(64) NOT NULL,
    groups jsonb,
    last_logout timestamp with time zone
);

CREATE INDEX IF NOT EXISTS authsrv_ssoaccount_name_4def83cc ON authsrv_ssoaccount USING btree (name);

CREATE INDEX IF NOT EXISTS authsrv_ssoaccount_name_4def83cc_like ON authsrv_ssoaccount USING btree (name varchar_pattern_ops);

CREATE INDEX IF NOT EXISTS authsrv_ssoaccount_organization_id_d2a979a5 ON authsrv_ssoaccount USING btree (organization_id);

CREATE INDEX IF NOT EXISTS authsrv_ssoaccount_username_029374ce_like ON authsrv_ssoaccount USING btree (username varchar_pattern_ops);
