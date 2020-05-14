
-- Create a function allowing to register a message with
-- the specified type for a given player.
CREATE OR REPLACE FUNCTION create_message_for(player_id uuid, message_name text, VARIADIC args text[]) RETURNS VOID AS $$
DECLARE
  msg_id uuid := uuid_generate_v4();
  pos integer := 0;
  arg text;
BEGIN
  -- Insert the message itself.
  INSERT INTO messages_players
    SELECT
      msg_id AS id,
      player_id AS player,
      mi.id AS message
    FROM
      messages_ids AS mi
    WHERE
      mi.name = message_name;

  -- And then all its arguments. We need a counter to
  -- determine the position of the arg and preserve
  -- the input order.
  FOREACH arg IN ARRAY args
  LOOP
    INSERT INTO messages_arguments("message", "position", "argument")
      VALUES(msg_id, pos, arg);

    pos := pos + 1;
  END LOOP;
END
$$ LANGUAGE plpgsql;
