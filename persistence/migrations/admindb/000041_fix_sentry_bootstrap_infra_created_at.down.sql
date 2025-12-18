UPDATE sentry_bootstrap_infra
SET created_at = CURRENT_TIMESTAMP
WHERE created_at IS NULL;

ALTER TABLE sentry_bootstrap_infra
ALTER COLUMN created_at SET DEFAULT CURRENT_TIMESTAMP;

ALTER TABLE sentry_bootstrap_infra
ALTER COLUMN created_at SET NOT NULL;