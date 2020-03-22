
-- Drop both the accounts and players tables and the associated triggers.
DROP TRIGGER update_player_creation_time ON players;
DROP TABLE players;

DROP TRIGGER update_account_creation_time ON accounts;
DROP TABLE accounts;
