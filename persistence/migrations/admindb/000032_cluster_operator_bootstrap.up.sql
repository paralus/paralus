CREATE TABLE IF NOT EXISTS cluster_operator_bootstrap (
    cluster_id uuid NOT NULL,
    created_at timestamp with time zone NOT NULL,
    modified_at timestamp with time zone NOT NULL,
    trash boolean DEFAULT false NOT NULL,
    yaml_content text
);
