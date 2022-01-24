CREATE TABLE IF NOT EXISTS cluster_project_cluster (
    project_id uuid NOT NULL,
    cluster_id uuid NOT NULL
);

ALTER TABLE cluster_project_cluster OWNER TO clusterdbuser;

CREATE INDEX cluster_project_cluster_project_id_cluster_id_key ON cluster_project_cluster USING btree (project_id, cluster_id);

ALTER TABLE ONLY cluster_project_cluster
    ADD CONSTRAINT cluster_project_cluster_cluster_id_fkey FOREIGN KEY (cluster_id) 
    REFERENCES cluster_clusters(id) DEFERRABLE INITIALLY DEFERRED;