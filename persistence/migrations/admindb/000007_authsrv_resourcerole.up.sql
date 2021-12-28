CREATE TABLE IF NOT EXISTS authsrv_resourcerole (
    id integer NOT NULL,
    name character varying(256) NOT NULL,
    description character varying(512) NOT NULL,
    created_at timestamp with time zone NOT NULL,
    modified_at timestamp with time zone NOT NULL,
    trash boolean NOT NULL,
    is_global boolean NOT NULL,
    scope character varying(256) NOT NULL,
    organization_id integer,
    partner_id integer
);

ALTER TABLE authsrv_resourcerole OWNER TO admindbuser;

CREATE SEQUENCE IF NOT EXISTS authsrv_resourcerole_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER TABLE authsrv_resourcerole_id_seq OWNER TO admindbuser;

ALTER SEQUENCE authsrv_resourcerole_id_seq OWNED BY authsrv_resourcerole.id;

ALTER TABLE ONLY authsrv_resourcerole ALTER COLUMN id SET DEFAULT nextval('authsrv_resourcerole_id_seq'::regclass);

ALTER TABLE ONLY authsrv_resourcerole ADD CONSTRAINT authsrv_resourcerole_pkey PRIMARY KEY (id);

CREATE INDEX authsrv_resourcerole_name_a93b875a ON authsrv_resourcerole USING btree (name);

CREATE INDEX authsrv_resourcerole_name_a93b875a_like ON authsrv_resourcerole USING btree (name varchar_pattern_ops);

CREATE INDEX authsrv_resourcerole_organization_id_9a0a7e7e ON authsrv_resourcerole USING btree (organization_id);

CREATE INDEX authsrv_resourcerole_partner_id_de49ca91 ON authsrv_resourcerole USING btree (partner_id);

ALTER TABLE ONLY authsrv_resourcerole
    ADD CONSTRAINT authsrv_resourcerole_organization_id_9a0a7e7e_fk_authsrv_o FOREIGN KEY (organization_id) 
    REFERENCES authsrv_organization(id) DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE ONLY authsrv_resourcerole
    ADD CONSTRAINT authsrv_resourcerole_partner_id_de49ca91_fk_authsrv_partner_id FOREIGN KEY (partner_id) 
    REFERENCES authsrv_partner(id) DEFERRABLE INITIALLY DEFERRED;