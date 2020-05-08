
-- Create the table defining accounts.
CREATE TABLE accounts (
  id uuid NOT NULL DEFAULT uuid_generate_v4(),
  name text NOT NULL,
  mail text NOT NULL,
  password text NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE (name),
  UNIQUE (mail)
);

-- Create the trigger on the table to update the `created_at` field.
CREATE TRIGGER update_accounts_creation BEFORE INSERT ON accounts FOR EACH ROW EXECUTE PROCEDURE update_created_at();

-- Create the table defining players' accounts in various universes.
CREATE TABLE players (
  id uuid NOT NULL DEFAULT uuid_generate_v4(),
  universe uuid NOT NULL,
  account uuid NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  name text NOT NULL,
  PRIMARY KEY (id),
  FOREIGN KEY (account) REFERENCES accounts(id),
  FOREIGN KEY (universe) REFERENCES universes(id),
  UNIQUE (universe, account),
  UNIQUE (universe, name)
);

-- Create the trigger on the table to update the `created_at` field.
CREATE TRIGGER update_players_creation BEFORE INSERT ON players FOR EACH ROW EXECUTE PROCEDURE update_created_at();

-- Create a table representing the technologies for a given player.
CREATE TABLE player_technologies (
  player uuid NOT NULL,
  technology uuid NOT NULL,
  level integer NOT NULL DEFAULT 0,
  FOREIGN KEY (player) REFERENCES players(id),
  FOREIGN KEY (technology) REFERENCES technologies(id)
);
