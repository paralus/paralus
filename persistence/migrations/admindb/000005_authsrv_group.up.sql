CREATE TABLE IF NOT EXISTS authsrv_group (
    id integer NOT NULL,
    name character varying(256) NOT NULL,
    description character varying(512) NOT NULL,
    created_at timestamp with time zone NOT NULL,
    modified_at timestamp with time zone NOT NULL,
    trash boolean NOT NULL,
    organization_id integer NOT NULL,
    partner_id integer NOT NULL,
    type character varying(64) NOT NULL
);

ALTER TABLE authsrv_group OWNER TO admindbuser;

CREATE SEQUENCE IF NOT EXISTS authsrv_group_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER TABLE authsrv_group_id_seq OWNER TO admindbuser;

ALTER SEQUENCE authsrv_group_id_seq OWNED BY authsrv_group.id;

ALTER TABLE ONLY authsrv_group ALTER COLUMN id SET DEFAULT nextval('authsrv_group_id_seq'::regclass);

ALTER TABLE ONLY authsrv_group ADD CONSTRAINT authsrv_group_pkey PRIMARY KEY (id);

CREATE INDEX authsrv_group_name_d90b4524 ON authsrv_group USING btree (name);

CREATE INDEX authsrv_group_name_d90b4524_like ON authsrv_group USING btree (name varchar_pattern_ops);

CREATE INDEX authsrv_group_organization_id_e070e826 ON authsrv_group USING btree (organization_id);

CREATE INDEX authsrv_group_partner_id_1de9ab46 ON authsrv_group USING btree (partner_id);

ALTER TABLE ONLY authsrv_group
    ADD CONSTRAINT authsrv_group_organization_id_e070e826_fk_authsrv_o FOREIGN KEY (organization_id) 
    REFERENCES authsrv_organization(id) DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE ONLY authsrv_group
    ADD CONSTRAINT authsrv_group_partner_id_1de9ab46_fk_authsrv_partner_id FOREIGN KEY (partner_id) 
    REFERENCES authsrv_partner(id) DEFERRABLE INITIALLY DEFERRED;