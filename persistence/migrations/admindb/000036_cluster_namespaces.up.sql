CREATE TABLE IF NOT EXISTS cluster_namespaces (
    cluster_id uuid NOT NULL,
    name varchar NOT NULL,
    hash varchar NOT NULL,
    deleted_at timestamp WITH time zone,
    type varchar not null,
    namespace jsonb not null,
    conditions jsonb not null default '[]'::jsonb,
    status jsonb not null default '{}'::jsonb
);

ALTER TABLE cluster_namespaces OWNER TO admindbuser;

ALTER TABLE ONLY cluster_namespaces ADD CONSTRAINT cluster_namespaces_pkey PRIMARY KEY (cluster_id, name);

ALTER TABLE ONLY cluster_namespaces
    ADD CONSTRAINT cluster_project_cluster_cluster_id_fkey FOREIGN KEY (cluster_id) 
    REFERENCES cluster_clusters(id) DEFERRABLE INITIALLY DEFERRED;