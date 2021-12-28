CREATE TABLE IF NOT EXISTS cluster_nodes (
    cluster_id integer NOT NULL,
    organization_id integer not null default 0,
    partner_id integer not null default 0,
    project_id integer not null default 0,
    name varchar NOT NULL,
    display_name varchar NOT NULL,
    created_at timestamp WITH time zone NOT NULL,
    modified_at timestamp WITH time zone,
    deleted_at timestamp with time zone,
    labels jsonb NOT NULL DEFAULT '{}'::jsonb,
    annotations jsonb NOT NULL DEFAULT '{}'::jsonb,
    unschedulable bool NOT NULL DEFAULT FALSE,
    taints jsonb NOT NULL DEFAULT '[]'::jsonb,
    conditions jsonb NOT NULL DEFAULT '[]'::jsonb,
    node_info jsonb NOT NULL DEFAULT '{}'::jsonb,
    state varchar NOT NULL,
    capacity jsonb NOT NULL DEFAULT '{}'::jsonb,
    allocatable jsonb NOT NULL DEFAULT '{}'::jsonb,
    allocated jsonb NOT NULL DEFAULT '{}'::jsonb,
    ips jsonb NOT NULL DEFAULT '[]'::jsonb,
    id bigserial
);

ALTER TABLE cluster_nodes OWNER TO clusterdbuser;

CREATE SEQUENCE IF NOT EXISTS cluster_nodes_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER TABLE cluster_nodes_id_seq OWNER TO clusterdbuser;

ALTER SEQUENCE cluster_nodes_id_seq OWNED BY cluster_nodes.id;

ALTER TABLE ONLY cluster_nodes ADD CONSTRAINT cluster_nodes_name_cluster_id_key PRIMARY KEY (name, cluster_id);

CREATE INDEX idx_nodes_lables ON cluster_nodes USING GIN (labels jsonb_path_ops);

ALTER TABLE ONLY cluster_nodes
    ADD CONSTRAINT cluster_nodes_cluster_id_fkey FOREIGN KEY (cluster_id) 
    REFERENCES cluster_clusters(id) DEFERRABLE INITIALLY DEFERRED;