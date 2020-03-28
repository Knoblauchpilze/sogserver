
-- Create the table defining accounts.
CREATE TABLE accounts (
    id uuid NOT NULL DEFAULT uuid_generate_v4(),
    name text NOT NULL,
    mail text NOT NULL UNIQUE,
    created_at timestamp with time zone default current_timestamp,
    PRIMARY KEY (id)
);

-- Trigger to update the `created_at` field of the table.
CREATE TRIGGER update_account_creation_time BEFORE INSERT ON accounts FOR EACH ROW EXECUTE PROCEDURE update_created_at_column();

-- Create the table defining players' accounts in various universes.
CREATE TABLE players (
  id uuid NOT NULL DEFAULT uuid_generate_v4(),
  uni uuid NOT NULL references universes,
  player uuid NOT NULL references accounts,
  created_at timestamp with time zone default current_timestamp,
  name text NOT NULL,
  PRIMARY KEY (id)
);

-- Trigger to update the `created_at` field of the table.
CREATE TRIGGER update_player_creation_time BEFORE INSERT ON players FOR EACH ROW EXECUTE PROCEDURE update_created_at_column();

-- Create a table representing the technologies for a given player.
CREATE TABLE player_technologies (
  player uuid NOT NULL references players,
  technology uuid NOT NULL references technologies,
  level integer NOT NULL default 0
);
