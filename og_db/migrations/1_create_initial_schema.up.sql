
-- Common properties of the DB
SET client_encoding = 'UTF8';

SET search_path = public, pg_catalog;
SET default_tablespace = '';

-- Convenience function to update the `created_at` column of a table
-- with the current date at the moment of the call.
CREATE OR REPLACE FUNCTION update_created_at_column()
RETURNS TRIGGER AS $$
BEGIN
  NEW.created_at = now();
  return NEW;
END;
$$ language plpgsql;
