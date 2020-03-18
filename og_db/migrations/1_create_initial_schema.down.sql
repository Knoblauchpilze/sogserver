-- Drop tables created in the `up` part of the migration.
drop trigger update_account_creation_time on accounts;
drop table accounts;

drop trigger update_player_creation_time on players;
drop table players;

drop trigger update_universe_creation_time on universes;
drop table universes;

-- Drop convenience update to `created_at` column.
DROP FUNCTION update_created_at_column();
