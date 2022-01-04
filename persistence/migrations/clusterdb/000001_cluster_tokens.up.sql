CREATE TABLE IF NOT EXISTS cluster_tokens (
    -- database id fields
    id bigserial NOT NULL,
    name varchar NOT NULL,
    organization_id integer not null default 0,
    partner_id integer not null default 0,
    project_id integer not null default 0,
    -- rafay meta fields
    display_name varchar NOT NULL,
    created_at timestamp WITH time zone NOT NULL,
    modified_at timestamp WITH time zone,
    deleted_at timestamp with time zone,
    labels jsonb NOT NULL DEFAULT '{}'::jsonb,
    annotations jsonb NOT NULL DEFAULT '{}'::jsonb,
    -- cluster token spec fields
    token_type varchar NOT NULL,
    -- cluster token status fields
    state varchar NOT NULL
);

ALTER TABLE cluster_tokens OWNER TO clusterdbuser;

CREATE SEQUENCE IF NOT EXISTS cluster_tokens_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER TABLE cluster_tokens_id_seq OWNER TO clusterdbuser;

ALTER SEQUENCE cluster_tokens_id_seq OWNED BY cluster_tokens.id;

ALTER TABLE ONLY cluster_tokens ALTER COLUMN id SET DEFAULT nextval('cluster_tokens_id_seq'::regclass);

ALTER TABLE ONLY cluster_tokens ADD CONSTRAINT cluster_tokens_pkey PRIMARY KEY (name);