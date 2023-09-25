CREATE TABLE IF NOT EXISTS cluster_tokens (
    id uuid NOT NULL default uuid_generate_v4(),
    name varchar PRIMARY KEY,
    organization_id uuid not null,
    partner_id uuid not null,
    project_id uuid not null,
    display_name varchar NOT NULL,
    created_at timestamp WITH time zone NOT NULL,
    modified_at timestamp WITH time zone,
    deleted_at timestamp with time zone,
    trash boolean NOT NULL default false,
    labels jsonb NOT NULL DEFAULT '{}'::jsonb,
    annotations jsonb NOT NULL DEFAULT '{}'::jsonb,
    token_type varchar NOT NULL,
    state varchar NOT NULL
);
