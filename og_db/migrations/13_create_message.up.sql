
-- Create the messages type.
CREATE TABLE messages_types (
  id uuid NOT NULL DEFAULT uuid_generate_v4(),
  type text NOT NULL,
  PRIMARY KEY (id),
  UNIQUE (type)
);

-- Create the possible messages identifiers.
CREATE TABLE messages_ids (
  id uuid NOT NULL DEFAULT uuid_generate_v4(),
  type uuid NOT NULL,
  content text NOT NULL,
  PRIMARY KEY (id),
  FOREIGN KEY (type) REFERENCES messages_types(id)
);

-- Create the table referencing messages for players.
CREATE TABLE messages_players (
  id uuid NOT NULL DEFAULT uuid_generate_v4(),
  player uuid NOT NULL,
  message uuid NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  FOREIGN KEY(player) REFERENCES players(id),
  FOREIGN KEY(message) REFERENCES messages_ids
);

-- Create the trigger on the table to update the `created_at` field.
CREATE TRIGGER update_messages_creation BEFORE INSERT ON messages_players FOR EACH ROW EXECUTE PROCEDURE update_created_at();

-- Create the table referencing the arguments for messages.
CREATE TABLE messages_arguments (
  message uuid NOT NULL,
  position integer NOT NULL,
  argument text NOT NULL,
  FOREIGN KEY (message) REFERENCES messages_players(id),
  UNIQUE (message, position)
);
