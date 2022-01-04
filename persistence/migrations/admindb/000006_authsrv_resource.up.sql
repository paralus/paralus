CREATE TABLE IF NOT EXISTS authsrv_resource (
    name character varying(256) NOT NULL,
    base_url character varying(256) NOT NULL,
    resource_urls jsonb NOT NULL,
    resource_action_urls jsonb NOT NULL,
    trash boolean NOT NULL
);

ALTER TABLE authsrv_resource OWNER TO admindbuser;

ALTER TABLE ONLY authsrv_resource ADD CONSTRAINT authsrv_resource_pkey PRIMARY KEY (name);

CREATE INDEX authsrv_resource_name_5a29761f_like ON authsrv_resource USING btree (name varchar_pattern_ops);