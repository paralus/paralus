CREATE TABLE IF NOT EXISTS authsrv_grouprole (
    id uuid NOT NULL default uuid_generate_v4(),
    name character varying(256) NOT NULL,
    description character varying(512) NOT NULL,
    created_at timestamp with time zone NOT NULL,
    modified_at timestamp with time zone NOT NULL,
    trash boolean NOT NULL,
    "default" boolean NOT NULL,
    active boolean NOT NULL,
    group_id uuid NOT NULL,
    organization_id uuid,
    partner_id uuid,
    role_id uuid NOT NULL
);

ALTER TABLE ONLY authsrv_grouprole ADD CONSTRAINT authsrv_grouprole_pkey PRIMARY KEY (id);

CREATE INDEX authsrv_grouprole_group_id_2f1402a5 ON authsrv_grouprole USING btree (group_id);

CREATE INDEX authsrv_grouprole_name_3810bc7c ON authsrv_grouprole USING btree (name);

CREATE INDEX authsrv_grouprole_name_3810bc7c_like ON authsrv_grouprole USING btree (name varchar_pattern_ops);

CREATE INDEX authsrv_grouprole_organization_id_9e77495d ON authsrv_grouprole USING btree (organization_id);

CREATE INDEX authsrv_grouprole_partner_id_f27b027a ON authsrv_grouprole USING btree (partner_id);

CREATE INDEX authsrv_grouprole_role_id_786f31f9 ON authsrv_grouprole USING btree (role_id);

ALTER TABLE ONLY authsrv_grouprole
    ADD CONSTRAINT authsrv_grouprole_group_id_2f1402a5_fk_authsrv_group_id FOREIGN KEY (group_id) 
    REFERENCES authsrv_group(id) DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE ONLY authsrv_grouprole
    ADD CONSTRAINT authsrv_grouprole_organization_id_9e77495d_fk_authsrv_o FOREIGN KEY (organization_id) 
    REFERENCES authsrv_organization(id) DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE ONLY authsrv_grouprole
    ADD CONSTRAINT authsrv_grouprole_partner_id_f27b027a_fk_authsrv_partner_id FOREIGN KEY (partner_id) 
    REFERENCES authsrv_partner(id) DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE ONLY authsrv_grouprole
    ADD CONSTRAINT authsrv_grouprole_role_id_786f31f9_fk_authsrv_resourcerole_id FOREIGN KEY (role_id) 
    REFERENCES authsrv_resourcerole(id) DEFERRABLE INITIALLY DEFERRED;