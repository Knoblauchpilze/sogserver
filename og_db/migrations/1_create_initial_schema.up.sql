
-- Common properties of the DB
SET client_encoding = 'UTF8';

SET search_path = public, pg_catalog;
SET default_tablespace = '';

-- Register a function to automatically update the `created_at` field
-- of a new object to insert in the DB. We also create a similar one
-- to update the `joined_at` column and a `updated_at` column.
CREATE OR REPLACE FUNCTION update_created_at() RETURNS TRIGGER AS $$
  BEGIN
    NEW.created_at = now();
    RETURN NEW;
  END;
$$ language 'plpgsql';

CREATE OR REPLACE FUNCTION update_joined_at() RETURNS TRIGGER AS $$
  BEGIN
    NEW.joined_at = now();
    RETURN NEW;
  END;
$$ language 'plpgsql';

CREATE OR REPLACE FUNCTION update_updated_at() RETURNS TRIGGER AS $$
  BEGIN
    NEW.updated_at = now();
    RETURN NEW;
  END;
$$ language 'plpgsql';
