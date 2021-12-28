CREATE TABLE IF NOT EXISTS authsrv_groupaccount (
    id integer NOT NULL,
    name character varying(256) NOT NULL,
    description character varying(512) NOT NULL,
    created_at timestamp with time zone NOT NULL,
    modified_at timestamp with time zone NOT NULL,
    trash boolean NOT NULL,
    account_id integer NOT NULL,
    group_id integer NOT NULL,
    active boolean not null default true
);

ALTER TABLE authsrv_groupaccount OWNER TO admindbuser;

CREATE SEQUENCE IF NOT EXISTS authsrv_groupaccount_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER TABLE authsrv_groupaccount_id_seq OWNER TO admindbuser;

ALTER SEQUENCE authsrv_groupaccount_id_seq OWNED BY authsrv_groupaccount.id;

ALTER TABLE ONLY authsrv_groupaccount ALTER COLUMN id SET DEFAULT nextval('authsrv_groupaccount_id_seq'::regclass);

ALTER TABLE ONLY authsrv_groupaccount ADD CONSTRAINT authsrv_groupaccount_pkey PRIMARY KEY (id);

CREATE INDEX authsrv_groupaccount_account_id_041e4e98 ON authsrv_groupaccount USING btree (account_id);

CREATE INDEX authsrv_groupaccount_group_id_c67750ef ON authsrv_groupaccount USING btree (group_id);

CREATE INDEX authsrv_groupaccount_name_d17de056 ON authsrv_groupaccount USING btree (name);

CREATE INDEX authsrv_groupaccount_name_d17de056_like ON authsrv_groupaccount USING btree (name varchar_pattern_ops);

ALTER TABLE ONLY authsrv_groupaccount
    ADD CONSTRAINT authsrv_groupaccount_account_id_041e4e98_fk_authsrv_account_id FOREIGN KEY (account_id) 
    REFERENCES authsrv_account(id) DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE ONLY authsrv_groupaccount
    ADD CONSTRAINT authsrv_groupaccount_group_id_c67750ef_fk_authsrv_group_id FOREIGN KEY (group_id) 
    REFERENCES authsrv_group(id) DEFERRABLE INITIALLY DEFERRED;