DROP FUNCTION IF EXISTS providers_after_change_trigger() CASCADE;
DROP TRIGGER IF EXISTS providers_updated ON authsrv_oidc_provider;
DROP TRIGGER IF EXISTS providers_inserted ON authsrv_oidc_provider;
DROP TRIGGER IF EXISTS providers_deleted ON authsrv_oidc_provider;
