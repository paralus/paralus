CREATE TABLE IF NOT EXISTS authsrv_projectaccountnamespacerole (
    id integer NOT NULL,
    name character varying(256) NOT NULL,
    description character varying(512) NOT NULL,
    created_at timestamp with time zone NOT NULL,
    modified_at timestamp with time zone NOT NULL,
    trash boolean NOT NULL,
    namespace_id integer NOT NULL,
    active boolean NOT NULL,
    account_id integer NOT NULL,
    organization_id integer,
    partner_id integer,
    project_id integer,
    role_id integer NOT NULL
);

ALTER TABLE authsrv_projectaccountnamespacerole OWNER TO admindbuser;

CREATE SEQUENCE IF NOT EXISTS authsrv_projectaccountnamespacerole_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER TABLE authsrv_projectaccountnamespacerole_id_seq OWNER TO admindbuser;

ALTER SEQUENCE authsrv_projectaccountnamespacerole_id_seq OWNED BY authsrv_projectaccountnamespacerole.id;

ALTER TABLE ONLY authsrv_projectaccountnamespacerole ALTER COLUMN id SET DEFAULT nextval('authsrv_projectaccountnamespacerole_id_seq'::regclass);

ALTER TABLE ONLY authsrv_projectaccountnamespacerole ADD CONSTRAINT authsrv_projectaccountnamespacerole_pkey PRIMARY KEY (id);

CREATE INDEX authsrv_projectaccountnamespacerole_account_id_4fac0ac2 ON authsrv_projectaccountnamespacerole USING btree (account_id);

CREATE INDEX authsrv_projectaccountnamespacerole_name_216353a4 ON authsrv_projectaccountnamespacerole USING btree (name);

CREATE INDEX authsrv_projectaccountnamespacerole_name_216353a4_like ON authsrv_projectaccountnamespacerole USING btree (name varchar_pattern_ops);

CREATE INDEX authsrv_projectaccountnamespacerole_organization_id_96c921c9 ON authsrv_projectaccountnamespacerole USING btree (organization_id);

CREATE INDEX authsrv_projectaccountnamespacerole_partner_id_9bec6899 ON authsrv_projectaccountnamespacerole USING btree (partner_id);

CREATE INDEX authsrv_projectaccountnamespacerole_project_id_66e567ed ON authsrv_projectaccountnamespacerole USING btree (project_id);

CREATE INDEX authsrv_projectaccountnamespacerole_role_id_8a5411cc ON authsrv_projectaccountnamespacerole USING btree (role_id);

ALTER TABLE ONLY authsrv_projectaccountnamespacerole
    ADD CONSTRAINT authsrv_projectaccou_account_id_4fac0ac2_fk_authsrv_a FOREIGN KEY (account_id) 
    REFERENCES authsrv_account(id) DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE ONLY authsrv_projectaccountnamespacerole
    ADD CONSTRAINT authsrv_projectaccou_organization_id_96c921c9_fk_authsrv_o FOREIGN KEY (organization_id) 
    REFERENCES authsrv_organization(id) DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE ONLY authsrv_projectaccountnamespacerole
    ADD CONSTRAINT authsrv_projectaccou_partner_id_9bec6899_fk_authsrv_p FOREIGN KEY (partner_id) 
    REFERENCES authsrv_partner(id) DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE ONLY authsrv_projectaccountnamespacerole
    ADD CONSTRAINT authsrv_projectaccou_project_id_66e567ed_fk_authsrv_p FOREIGN KEY (project_id) 
    REFERENCES authsrv_project(id) DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE ONLY authsrv_projectaccountnamespacerole
    ADD CONSTRAINT authsrv_projectaccou_role_id_8a5411cc_fk_authsrv_r FOREIGN KEY (role_id) 
    REFERENCES authsrv_resourcerole(id) DEFERRABLE INITIALLY DEFERRED;