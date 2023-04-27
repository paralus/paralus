CREATE TABLE IF NOT EXISTS sentry_kubectl_cluster_settings (
    name varchar PRIMARY KEY,
    organization_id uuid NOT NULL,
    partner_id uuid NOT NULL,
    disable_web_kubectl boolean NOT NULL DEFAULT FALSE,
    disable_cli_kubectl boolean NOT NULL DEFAULT FALSE,
    modified_at timestamp WITH time zone,
    created_at timestamp WITH time zone NOT NULL,
    deleted_at timestamp WITH time zone,
    CONSTRAINT sentry_kubectl_cluster_settin_name_partner_id_organization__key UNIQUE (name, partner_id, organization_id)
);
