CREATE TABLE IF NOT EXISTS authsrv_projectaccountresourcerole (
    id integer NOT NULL,
    name character varying(256) NOT NULL,
    description character varying(512) NOT NULL,
    created_at timestamp with time zone NOT NULL,
    modified_at timestamp with time zone NOT NULL,
    trash boolean NOT NULL,
    "default" boolean NOT NULL,
    active boolean NOT NULL,
    account_id integer NOT NULL,
    organization_id integer,
    partner_id integer,
    project_id integer,
    role_id integer NOT NULL
);

ALTER TABLE authsrv_projectaccountresourcerole OWNER TO admindbuser;

CREATE SEQUENCE IF NOT EXISTS authsrv_projectaccountresourcerole_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER TABLE authsrv_projectaccountresourcerole_id_seq OWNER TO admindbuser;

ALTER SEQUENCE authsrv_projectaccountresourcerole_id_seq OWNED BY authsrv_projectaccountresourcerole.id;

ALTER TABLE ONLY authsrv_projectaccountresourcerole ALTER COLUMN id SET DEFAULT nextval('authsrv_projectaccountresourcerole_id_seq'::regclass);

ALTER TABLE ONLY authsrv_projectaccountresourcerole ADD CONSTRAINT authsrv_projectaccountresourcerole_pkey PRIMARY KEY (id);

CREATE INDEX authsrv_projectaccountresourcerole_account_id_532ce8df ON authsrv_projectaccountresourcerole USING btree (account_id);

CREATE INDEX authsrv_projectaccountresourcerole_name_c4c3d60f ON authsrv_projectaccountresourcerole USING btree (name);

CREATE INDEX authsrv_projectaccountresourcerole_name_c4c3d60f_like ON authsrv_projectaccountresourcerole USING btree (name varchar_pattern_ops);

CREATE INDEX authsrv_projectaccountresourcerole_organization_id_91c5602d ON authsrv_projectaccountresourcerole USING btree (organization_id);

CREATE INDEX authsrv_projectaccountresourcerole_partner_id_81bde92c ON authsrv_projectaccountresourcerole USING btree (partner_id);

CREATE INDEX authsrv_projectaccountresourcerole_project_id_f8a43852 ON authsrv_projectaccountresourcerole USING btree (project_id);

CREATE INDEX authsrv_projectaccountresourcerole_role_id_a345b16f ON authsrv_projectaccountresourcerole USING btree (role_id);

ALTER TABLE ONLY authsrv_projectaccountresourcerole
    ADD CONSTRAINT authsrv_projectaccou_account_id_532ce8df_fk_authsrv_a FOREIGN KEY (account_id) 
    REFERENCES authsrv_account(id) DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE ONLY authsrv_projectaccountresourcerole
    ADD CONSTRAINT authsrv_projectaccou_organization_id_91c5602d_fk_authsrv_o FOREIGN KEY (organization_id) 
    REFERENCES authsrv_organization(id) DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE ONLY authsrv_projectaccountresourcerole
    ADD CONSTRAINT authsrv_projectaccou_partner_id_81bde92c_fk_authsrv_p FOREIGN KEY (partner_id) 
    REFERENCES authsrv_partner(id) DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE ONLY authsrv_projectaccountresourcerole
    ADD CONSTRAINT authsrv_projectaccou_project_id_f8a43852_fk_authsrv_p FOREIGN KEY (project_id) 
    REFERENCES authsrv_project(id) DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE ONLY authsrv_projectaccountresourcerole
    ADD CONSTRAINT authsrv_projectaccou_role_id_a345b16f_fk_authsrv_r FOREIGN KEY (role_id) 
    REFERENCES authsrv_resourcerole(id) DEFERRABLE INITIALLY DEFERRED;