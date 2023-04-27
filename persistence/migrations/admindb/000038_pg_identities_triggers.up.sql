CREATE OR REPLACE FUNCTION identities_after_change() RETURNS TRIGGER AS $$
  DECLARE
  row RECORD;
  output TEXT;
  
  BEGIN
  IF (TG_OP = 'DELETE') THEN
    row = OLD;
  ELSE
    row = NEW;
  END IF;
  
  output = TG_OP || ',' || row.id || ',' || row.traits;
  PERFORM pg_notify('identities:changed',output);
  RETURN NULL;
  END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE TRIGGER trigger_identities_update
  AFTER INSERT OR UPDATE OR DELETE
  ON identities
  FOR EACH ROW
  EXECUTE PROCEDURE identities_after_change();
