CREATE TABLE IF NOT EXISTS authsrv_project (
    id integer NOT NULL,
    name character varying(256) NOT NULL,
    description character varying(512) NOT NULL,
    created_at timestamp with time zone NOT NULL,
    modified_at timestamp with time zone NOT NULL,
    trash boolean NOT NULL,
    organization_id integer,
    partner_id integer,
    "default" boolean NOT NULL
);

ALTER TABLE authsrv_project OWNER TO admindbuser;

CREATE SEQUENCE IF NOT EXISTS authsrv_project_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER TABLE authsrv_project_id_seq OWNER TO admindbuser;

ALTER SEQUENCE authsrv_project_id_seq OWNED BY authsrv_project.id;

ALTER TABLE ONLY authsrv_project ALTER COLUMN id SET DEFAULT nextval('authsrv_project_id_seq'::regclass);

ALTER TABLE ONLY authsrv_project ADD CONSTRAINT authsrv_project_pkey PRIMARY KEY (id);

CREATE INDEX authsrv_project_name_1b8dd279 ON authsrv_project USING btree (name);

CREATE INDEX authsrv_project_name_1b8dd279_like ON authsrv_project USING btree (name varchar_pattern_ops);

CREATE INDEX authsrv_project_organization_id_77437387 ON authsrv_project USING btree (organization_id);

CREATE INDEX authsrv_project_partner_id_3d505b76 ON authsrv_project USING btree (partner_id);

ALTER TABLE ONLY authsrv_project
    ADD CONSTRAINT authsrv_project_organization_id_77437387_fk_authsrv_o FOREIGN KEY (organization_id) 
    REFERENCES authsrv_organization(id) DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE ONLY authsrv_project
    ADD CONSTRAINT authsrv_project_partner_id_3d505b76_fk_authsrv_partner_id FOREIGN KEY (partner_id) 
    REFERENCES authsrv_partner(id) DEFERRABLE INITIALLY DEFERRED;