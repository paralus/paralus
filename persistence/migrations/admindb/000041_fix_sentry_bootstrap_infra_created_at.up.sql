ALTER TABLE sentry_bootstrap_infra
ALTER COLUMN created_at DROP NOT NULL;

ALTER TABLE sentry_bootstrap_infra
ALTER COLUMN created_at DROP DEFAULT;