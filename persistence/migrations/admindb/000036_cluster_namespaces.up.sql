CREATE TABLE IF NOT EXISTS cluster_namespaces (
    cluster_id uuid NOT NULL REFERENCES cluster_clusters(id) DEFERRABLE INITIALLY DEFERRED,
    name varchar NOT NULL,
    hash varchar NOT NULL,
    deleted_at timestamp WITH time zone,
    type varchar not null,
    namespace jsonb not null,
    conditions jsonb not null default '[]'::jsonb,
    status jsonb not null default '{}'::jsonb,
    PRIMARY KEY (cluster_id, name)
);
