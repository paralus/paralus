CREATE TABLE IF NOT EXISTS authsrv_template (
    id uuid NOT NULL default uuid_generate_v4(),
    name character varying(256) NOT NULL,
    description character varying(512) NOT NULL,
    created_at timestamp with time zone NOT NULL,
    modified_at timestamp with time zone NOT NULL,
    trash boolean NOT NULL,
    type character varying(64) NOT NULL,
    source text NOT NULL,
    partner_id uuid NOT NULL
);

ALTER TABLE authsrv_template OWNER TO admindbuser;

ALTER TABLE ONLY authsrv_template ADD CONSTRAINT authsrv_template_pkey PRIMARY KEY (id);

CREATE INDEX authsrv_template_name_274ef2d3 ON authsrv_template USING btree (name);

CREATE INDEX authsrv_template_name_274ef2d3_like ON authsrv_template USING btree (name varchar_pattern_ops);

CREATE INDEX authsrv_template_partner_id_2fcb0ded ON authsrv_template USING btree (partner_id);

ALTER TABLE ONLY authsrv_template
    ADD CONSTRAINT authsrv_template_partner_id_2fcb0ded_fk_authsrv_partner_id FOREIGN KEY (partner_id) 
    REFERENCES authsrv_partner(id) DEFERRABLE INITIALLY DEFERRED;