CREATE TABLE IF NOT EXISTS authsrv_template (
    id uuid default uuid_generate_v4() PRIMARY KEY,
    name character varying(256) NOT NULL,
    description character varying(512) NOT NULL,
    created_at timestamp with time zone NOT NULL,
    modified_at timestamp with time zone NOT NULL,
    trash boolean NOT NULL,
    type character varying(64) NOT NULL,
    source text NOT NULL,
    partner_id uuid NOT NULL REFERENCES authsrv_partner(id) DEFERRABLE INITIALLY DEFERRED
);

CREATE INDEX IF NOT EXISTS authsrv_template_name_274ef2d3 ON authsrv_template USING btree (name);

CREATE INDEX IF NOT EXISTS authsrv_template_name_274ef2d3_like ON authsrv_template USING btree (name varchar_pattern_ops);

CREATE INDEX IF NOT EXISTS authsrv_template_partner_id_2fcb0ded ON authsrv_template USING btree (partner_id);
