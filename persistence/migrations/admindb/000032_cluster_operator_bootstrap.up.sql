CREATE TABLE IF NOT EXISTS cluster_operator_bootstrap (
    cluster_id uuid NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modified_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    trash boolean NOT NULL DEFAULT FALSE,
    yaml_content text
);
