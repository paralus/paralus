CREATE TABLE IF NOT EXISTS authsrv_template (
    id integer NOT NULL,
    name character varying(256) NOT NULL,
    description character varying(512) NOT NULL,
    created_at timestamp with time zone NOT NULL,
    modified_at timestamp with time zone NOT NULL,
    trash boolean NOT NULL,
    type character varying(64) NOT NULL,
    source text NOT NULL,
    partner_id integer NOT NULL
);

ALTER TABLE authsrv_template OWNER TO admindbuser;

CREATE SEQUENCE IF NOT EXISTS authsrv_template_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER TABLE authsrv_template_id_seq OWNER TO admindbuser;

ALTER SEQUENCE authsrv_template_id_seq OWNED BY authsrv_template.id;

ALTER TABLE ONLY authsrv_template ALTER COLUMN id SET DEFAULT nextval('authsrv_template_id_seq'::regclass);

ALTER TABLE ONLY authsrv_template ADD CONSTRAINT authsrv_template_pkey PRIMARY KEY (id);

CREATE INDEX authsrv_template_name_274ef2d3 ON authsrv_template USING btree (name);

CREATE INDEX authsrv_template_name_274ef2d3_like ON authsrv_template USING btree (name varchar_pattern_ops);

CREATE INDEX authsrv_template_partner_id_2fcb0ded ON authsrv_template USING btree (partner_id);

ALTER TABLE ONLY authsrv_template
    ADD CONSTRAINT authsrv_template_partner_id_2fcb0ded_fk_authsrv_partner_id FOREIGN KEY (partner_id) 
    REFERENCES authsrv_partner(id) DEFERRABLE INITIALLY DEFERRED;