
-- Create the table defining players.
CREATE TABLE players (
    id uuid NOT NULL DEFAULT uuid_generate_v4(),
    name text,
    mail text,
    created_at timestamp with time zone default current_timestamp,
    PRIMARY KEY (id)
);

-- Trigger to update the `created_at` field of the table.
CREATE TRIGGER update_player_creation_time BEFORE INSERT ON players FOR EACH ROW EXECUTE PROCEDURE update_created_at_column();

-- Create the table defining players' accounts in various universes.
create table accounts (
  uni uuid NOT NULL references universes,
  player uuid NOT NULL references players,
  created_at timestamp with time zone default current_timestamp
);

-- Trigger to update the `created_at` field of the table.
CREATE TRIGGER update_account_creation_time BEFORE INSERT ON accounts FOR EACH ROW EXECUTE PROCEDURE update_created_at_column();
