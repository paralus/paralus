CREATE TABLE IF NOT EXISTS authsrv_project (
    id uuid NOT NULL default uuid_generate_v4(),
    name character varying(256) NOT NULL,
    description character varying(512) NOT NULL,
    created_at timestamp with time zone NOT NULL,
    modified_at timestamp with time zone NOT NULL,
    trash boolean NOT NULL,
    organization_id uuid,
    partner_id uuid,
    "default" boolean NOT NULL
);

ALTER TABLE ONLY authsrv_project ADD CONSTRAINT authsrv_project_pkey PRIMARY KEY (id);

-- update when we have more than one org
CREATE UNIQUE index authsrv_project_unique_name ON authsrv_project (name) WHERE trash IS false;

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
