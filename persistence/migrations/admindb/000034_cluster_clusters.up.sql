CREATE TABLE IF NOT EXISTS cluster_clusters (
    id uuid NOT NULL default uuid_generate_v4(),
    organization_id uuid not null,
    partner_id uuid not null,
    project_id uuid not null,
    metro_id uuid,
    name varchar NOT NULL,
    display_name varchar NOT NULL,
    created_at timestamp WITH time zone NOT NULL,
    modified_at timestamp WITH time zone,
    trash boolean NOT NULL default false,
    deleted_at timestamp with time zone,
    labels jsonb NOT NULL DEFAULT '{}'::jsonb,
    annotations jsonb NOT NULL DEFAULT '{}'::jsonb,
    blueprint_ref varchar NOT NULL default '',
    cluster_type text,
    override_selector varchar NOT NULL default '',
    token varchar not null,
    conditions jsonb NOT NULL default '[]'::jsonb,
    published_blueprint varchar NOT NULL default '',
    system_task_count integer NOT NULL DEFAULT 0,
    custom_task_count integer NOT NULL DEFAULT 0,
    auxiliary_task_count integer NOT NULL DEFAULT 0,
    extra jsonb NOT NULL DEFAULT '{}'::jsonb,
    share_mode VARCHAR DEFAULT 'CUSTOM',
    proxy_config jsonb
);

ALTER TABLE ONLY cluster_clusters ADD CONSTRAINT cluster_clusters_pkey PRIMARY KEY (id);

CREATE INDEX cluster_clusters_name_organization_id_partner_id_key ON cluster_clusters USING btree (name, organization_id, partner_id);

CREATE INDEX idx_cluster_blueprint ON cluster_clusters USING btree (blueprint_ref, published_blueprint);

CREATE INDEX idx_clusters_labels ON cluster_clusters USING GIN (labels jsonb_path_ops);

ALTER TABLE ONLY cluster_clusters
    ADD CONSTRAINT cluster_clusters_token_fkey FOREIGN KEY (token) 
    REFERENCES cluster_tokens(name) DEFERRABLE INITIALLY DEFERRED;
