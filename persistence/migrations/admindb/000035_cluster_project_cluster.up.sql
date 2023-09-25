CREATE TABLE IF NOT EXISTS cluster_project_cluster (
    project_id uuid NOT NULL,
    cluster_id uuid NOT NULL REFERENCES cluster_clusters(id) DEFERRABLE INITIALLY DEFERRED,
    trash boolean NOT NULL default false
);

CREATE INDEX IF NOT EXISTS cluster_project_cluster_project_id_cluster_id_key ON cluster_project_cluster USING btree (project_id, cluster_id);
