
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
  name text NOT NULL,
  content text NOT NULL,
  PRIMARY KEY (id),
  FOREIGN KEY (type) REFERENCES messages_types(id),
  UNIQUE (name)
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

-- Seed the messages types.
INSERT INTO public.messages_types ("type") VALUES('fleets');
INSERT INTO public.messages_types ("type") VALUES('communication');
INSERT INTO public.messages_types ("type") VALUES('economy');
INSERT INTO public.messages_types ("type") VALUES('universe');
INSERT INTO public.messages_types ("type") VALUES('system');

-- Seed the messages identifiers.
INSERT INTO public.messages_ids ("type", "name", "content")
  VALUES(
    (SELECT id FROM messages_types WHERE type='fleets'),
    'colonization_suceeded',
    'the fleet has arrived at the assigned coordinates %1, found a new planet there and are beginning to develop upon it immediately'
  );
INSERT INTO public.messages_ids ("type", "name", "content")
  VALUES(
    (SELECT id FROM messages_types WHERE type='fleets'),
    'colonization_failed',
    'the fleet has arrived at the assigned coordinates %1 and could not perform the colonization process so they are returning home'
  );

INSERT INTO public.messages_ids ("type", "name", "content")
  VALUES(
    (SELECT id FROM messages_types WHERE type='fleets'),
    'harvesting_report',
    'your recycler(s) (%1) have a total cargo capacity of %2. At the target %3, %4 are floating in space. You have harvested %5.'
  );
