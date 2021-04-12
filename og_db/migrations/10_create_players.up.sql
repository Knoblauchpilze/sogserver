
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
  name text NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  FOREIGN KEY (account) REFERENCES accounts(id),
  FOREIGN KEY (universe) REFERENCES universes(id),
  UNIQUE (universe, account),
  UNIQUE (universe, name)
);

-- Create the trigger on the table to update the `created_at` field.
CREATE TRIGGER update_players_creation BEFORE INSERT ON players FOR EACH ROW EXECUTE PROCEDURE update_created_at();

-- Create a table representing the technologies for a given player.
CREATE TABLE players_technologies (
  player uuid NOT NULL,
  technology uuid NOT NULL,
  level integer NOT NULL DEFAULT 0,
  FOREIGN KEY (player) REFERENCES players(id),
  FOREIGN KEY (technology) REFERENCES technologies(id)
);

-- Create a table representing the points accumulated for a player.
CREATE TABLE players_points (
  player uuid NOT NULL,
  economy_points numeric(15, 5) NOT NULL DEFAULT 0,
  research_points numeric(15, 5) NOT NULL DEFAULT 0,
  military_points numeric(15, 5) NOT NULL DEFAULT 0,
  military_points_built numeric(15, 5) NOT NULL DEFAULT 0,
  military_points_lost numeric(15, 5) NOT NULL DEFAULT 0,
  military_points_destroyed numeric(15, 5) NOT NULL DEFAULT 0,
  FOREIGN KEY (player) REFERENCES players(id)
);

-- Create a table to define title for player names.
CREATE TABLE players_titles (
  title text NOT NULL,
  UNIQUE (title)
);

-- Create a table to define names for players.
CREATE TABLE players_names (
  name text NOT NULL,
  UNIQUE (name)
);

-- Seed the players' titles.
INSERT INTO public.players_titles ("title") VALUES('Emperor');
INSERT INTO public.players_titles ("title") VALUES('Constable');
INSERT INTO public.players_titles ("title") VALUES('Warlord');
INSERT INTO public.players_titles ("title") VALUES('Engineer');
INSERT INTO public.players_titles ("title") VALUES('Seneschal');
INSERT INTO public.players_titles ("title") VALUES('Hotshot');

-- Seed the player's names.
INSERT INTO public.players_names ("name") VALUES('Jonzy');
INSERT INTO public.players_names ("name") VALUES('Bighanta');
INSERT INTO public.players_names ("name") VALUES('Choupeau');
INSERT INTO public.players_names ("name") VALUES('Knoppgrunt');
INSERT INTO public.players_names ("name") VALUES('Lesfruits');
