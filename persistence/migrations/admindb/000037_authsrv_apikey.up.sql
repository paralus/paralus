CREATE TABLE IF NOT EXISTS authsrv_apikey (
    id uuid NOT NULL default uuid_generate_v4(),
    name varchar NOT NULL,
    description varchar NOT NULL,
    created_at timestamp WITH time zone NOT NULL,
    modified_at timestamp WITH time zone,
    trash boolean NOT NULL default false,
    key varchar NOT NULL,
    account_id uuid,
    organization_id uuid,
    partner_id uuid not null,
    secret_migration varchar NOT NULL,
    secret text not null
);

ALTER TABLE authsrv_apikey OWNER TO admindbuser;

ALTER TABLE ONLY authsrv_apikey ADD CONSTRAINT authsrv_apikey_pkey PRIMARY KEY (id);