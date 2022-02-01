CREATE FUNCTION providers_after_change_trigger()
RETURNS TRIGGER AS $$
BEGIN
  PERFORM pg_notify('provider:changed', '');
  RETURN NULL;
END;
$$
LANGUAGE plpgsql;

CREATE TRIGGER providers_updated
AFTER UPDATE ON authsrv_oidc_provider
FOR EACH ROW EXECUTE PROCEDURE providers_after_change_trigger();

CREATE TRIGGER providers_inserted
AFTER INSERT ON authsrv_oidc_provider
FOR EACH ROW EXECUTE PROCEDURE providers_after_change_trigger();

CREATE TRIGGER providers_deleted
AFTER DELETE ON authsrv_oidc_provider
FOR EACH ROW EXECUTE PROCEDURE providers_after_change_trigger();
