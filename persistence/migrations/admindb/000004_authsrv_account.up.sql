CREATE TABLE IF NOT EXISTS authsrv_ssoaccount (
    password character varying(128) NOT NULL,
    last_login timestamp with time zone,
    id uuid NOT NULL default uuid_generate_v4(),
    name character varying(256) NOT NULL,
    organization_id uuid,
    description character varying(512) NOT NULL,
    created_at timestamp with time zone NOT NULL,
    modified_at timestamp with time zone NOT NULL,
    trash boolean NOT NULL,
    username character varying(256) NOT NULL,
    phone character varying(36) NOT NULL,
    first_name character varying(64) NOT NULL,
    last_name character varying(64) NOT NULL,
    groups jsonb,
    last_logout timestamp with time zone
);

ALTER TABLE authsrv_ssoaccount OWNER TO admindbuser;

ALTER TABLE ONLY authsrv_ssoaccount ADD CONSTRAINT authsrv_ssoaccount_pkey PRIMARY KEY (id);

ALTER TABLE ONLY authsrv_ssoaccount ADD CONSTRAINT authsrv_ssoaccount_username_key UNIQUE (username);

CREATE INDEX authsrv_ssoaccount_name_4def83cc ON authsrv_ssoaccount USING btree (name);

CREATE INDEX authsrv_ssoaccount_name_4def83cc_like ON authsrv_ssoaccount USING btree (name varchar_pattern_ops);

CREATE INDEX authsrv_ssoaccount_organization_id_d2a979a5 ON authsrv_ssoaccount USING btree (organization_id);

CREATE INDEX authsrv_ssoaccount_username_029374ce_like ON authsrv_ssoaccount USING btree (username varchar_pattern_ops);

ALTER TABLE ONLY authsrv_ssoaccount
    ADD CONSTRAINT authsrv_ssoaccount_organization_id_d2a979a5_fk_authsrv_o FOREIGN KEY (organization_id) 
    REFERENCES authsrv_organization(id) DEFERRABLE INITIALLY DEFERRED;