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

--
-- Table describing universes.
--
CREATE TABLE universes (
    id uuid NOT NULL DEFAULT uuid_generate_v4(),
    name text,
    created_at timestamp with time zone default current_timestamp,
    PRIMARY KEY (id)
);

-- Trigger to update the `created_at` field of the table.
CREATE TRIGGER update_universe_creation_time BEFORE INSERT ON universes FOR EACH ROW EXECUTE PROCEDURE update_created_at_column();

--
-- Table describing players.
--
CREATE TABLE players (
    id uuid NOT NULL DEFAULT uuid_generate_v4(),
    name text,
    mail text,
    created_at timestamp with time zone default current_timestamp,
    PRIMARY KEY (id)
);

-- Trigger to update the `created_at` field of the table.
CREATE TRIGGER update_player_creation_time BEFORE INSERT ON players FOR EACH ROW EXECUTE PROCEDURE update_created_at_column();

--
-- Table describing players' accounts in universes.
--
create table accounts (
  uni uuid references universes,
  player uuid references players,
  created_at timestamp with time zone default current_timestamp
);

-- Trigger to update the `created_at` field of the table.
CREATE TRIGGER update_account_creation_time BEFORE INSERT ON accounts FOR EACH ROW EXECUTE PROCEDURE update_created_at_column();
