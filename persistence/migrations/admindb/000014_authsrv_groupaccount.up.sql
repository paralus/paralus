CREATE TABLE IF NOT EXISTS authsrv_groupaccount (
    id uuid default uuid_generate_v4() PRIMARY KEY,
    name character varying(256) NOT NULL,
    description character varying(512) NOT NULL,
    created_at timestamp with time zone NOT NULL,
    modified_at timestamp with time zone NOT NULL,
    trash boolean NOT NULL,
    account_id uuid NOT NULL,
    group_id uuid NOT NULL REFERENCES authsrv_group(id) DEFERRABLE INITIALLY DEFERRED,
    active boolean not null default true
);

CREATE INDEX IF NOT EXISTS authsrv_groupaccount_account_id_041e4e98 ON authsrv_groupaccount USING btree (account_id);

CREATE INDEX IF NOT EXISTS authsrv_groupaccount_group_id_c67750ef ON authsrv_groupaccount USING btree (group_id);

CREATE INDEX IF NOT EXISTS authsrv_groupaccount_name_d17de056 ON authsrv_groupaccount USING btree (name);

CREATE INDEX IF NOT EXISTS authsrv_groupaccount_name_d17de056_like ON authsrv_groupaccount USING btree (name varchar_pattern_ops);
