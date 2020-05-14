
-- Drop the table defining messages arguments.
DROP TABLE messages_arguments;

-- Drop the table defining messages for players and its associated trigger.
DROP TRIGGER update_messages_creation ON messages_players;
DROP TABLE messages_players;

-- Drop the table defining messages templates.
DROP TABLE messages_ids;

-- Drop the table defining messages types.
DROP TABLE messages_types;
