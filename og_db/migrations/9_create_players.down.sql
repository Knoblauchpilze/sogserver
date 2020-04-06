
-- Drop the table referencing the technologies per player.
DROP TABLE player_technologies;

-- Drop both the accounts and players tables and their associated triggers.
DROP TRIGGER update_players_creation ON players;
DROP TABLE players;

DROP TRIGGER update_accounts_creation ON accounts;
DROP TABLE accounts;
