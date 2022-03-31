CREATE TABLE IF NOT EXISTS authsrv_accountresourcerole (
    id uuid NOT NULL default uuid_generate_v4(),
    name character varying(256) NOT NULL,
    description character varying(512) NOT NULL,
    created_at timestamp with time zone NOT NULL,
    modified_at timestamp with time zone NOT NULL,
    trash boolean NOT NULL,
    "default" boolean NOT NULL,
    active boolean NOT NULL,
    account_id uuid NOT NULL,
    organization_id uuid,
    partner_id uuid,
    role_id uuid NOT NULL
);

ALTER TABLE authsrv_accountresourcerole OWNER TO admindbuser;

ALTER TABLE ONLY authsrv_accountresourcerole ADD CONSTRAINT authsrv_accountresourcerole_pkey PRIMARY KEY (id);

CREATE INDEX authsrv_accountresourcerole_account_id_229069ae ON authsrv_accountresourcerole USING btree (account_id);

CREATE INDEX authsrv_accountresourcerole_name_e1f3d8a0 ON authsrv_accountresourcerole USING btree (name);

CREATE INDEX authsrv_accountresourcerole_name_e1f3d8a0_like ON authsrv_accountresourcerole USING btree (name varchar_pattern_ops);

CREATE INDEX authsrv_accountresourcerole_organization_id_22bb772c ON authsrv_accountresourcerole USING btree (organization_id);

CREATE INDEX authsrv_accountresourcerole_partner_id_8e96aff4 ON authsrv_accountresourcerole USING btree (partner_id);

CREATE INDEX authsrv_accountresourcerole_role_id_769ec143 ON authsrv_accountresourcerole USING btree (role_id);

ALTER TABLE ONLY authsrv_accountresourcerole
    ADD CONSTRAINT authsrv_accountresou_organization_id_22bb772c_fk_authsrv_o FOREIGN KEY (organization_id) 
    REFERENCES authsrv_organization(id) DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE ONLY authsrv_accountresourcerole
    ADD CONSTRAINT authsrv_accountresou_partner_id_8e96aff4_fk_authsrv_p FOREIGN KEY (partner_id) 
    REFERENCES authsrv_partner(id) DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE ONLY authsrv_accountresourcerole
    ADD CONSTRAINT authsrv_accountresou_role_id_769ec143_fk_authsrv_r FOREIGN KEY (role_id) 
    REFERENCES authsrv_resourcerole(id) DEFERRABLE INITIALLY DEFERRED;
