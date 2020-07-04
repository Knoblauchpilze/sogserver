
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
    'the fleet has arrived at the assigned coordinates $COORD, found a new planet there and are beginning to develop upon it immediately'
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
    'your recycler(s) ($SHIP_COUNT) have a total cargo capacity of $CARGO. At the target $COORD, $RESOURCES are floating in space. You have harvested $HARVESTED.'
  );

INSERT INTO public.messages_ids ("type", "name", "content")
  VALUES(
    (SELECT id FROM messages_types WHERE type='fleets'),
    'destruction_report_all_destroyed',
    'your deathstar(s) ($SHIP_COUNT) fire at the moon at $COORD. The moon is shaking, and finally collapse under the concentrated graviton influx. However some debris seem to head to your fleet, destroying it.'
  );
INSERT INTO public.messages_ids ("type", "name", "content")
  VALUES(
    (SELECT id FROM messages_types WHERE type='fleets'),
    'destruction_report_moon_destroyed',
    'your deathstar(s) ($SHIP_COUNT) fire at the moon at $COORD. The moon is shaking, and finally collapse under the concentrated graviton influx. Your fleet returns home.'
  );
INSERT INTO public.messages_ids ("type", "name", "content")
  VALUES(
    (SELECT id FROM messages_types WHERE type='fleets'),
    'destruction_report_fleet_destroyed',
    'your deathstar(s) ($SHIP_COUNT) fire at the moon at $COORD. The moon seems to take the hit and not collapse. However a critical failure in the graviton generators occurs, destroying your fleet.'
  );
INSERT INTO public.messages_ids ("type", "name", "content")
  VALUES(
    (SELECT id FROM messages_types WHERE type='fleets'),
    'destruction_report_failed',
    'your deathstar(s) ($SHIP_COUNT) fire at the moon at $COORD. The moon seems to take the hit and not collapse. Your fleet returns home.'
  );

INSERT INTO public.messages_ids ("type", "name", "content")
  VALUES(
    (SELECT id FROM messages_types WHERE type='fleets'),
    'counter_espionage_report',
    'a foreign fleet from planet $PLANET_NAME $COORD ($PLAYER_NAME) has been spotted near your planet $OWN_PLANET_NAME $OWN_COORD. Probability of counter-espionage: $COUNTER_ESPIONAGE%.'
  );
INSERT INTO public.messages_ids ("type", "name", "content")
  VALUES(
    (SELECT id FROM messages_types WHERE type='fleets'),
    'espionage_report',
    '%1'
  );

INSERT INTO public.messages_ids ("type", "name", "content")
  VALUES(
    (SELECT id FROM messages_types WHERE type='fleets'),
    'fight_report',
    '$REPORT_HEADER $REPORT_RESULT $REPORT_FOOTER $REPORT_ATTACKERS $REPORT_DEFENDERS'
  );
INSERT INTO public.messages_ids ("type", "name", "content")
  VALUES(
    (SELECT id FROM messages_types WHERE type='fleets'),
    'fight_report_header',
    'combat report. Battle of $PLANET_NAME $COORD ($DATE)'
  );
INSERT INTO public.messages_ids ("type", "name", "content")
  VALUES(
    (SELECT id FROM messages_types WHERE type='fleets'),
    'fight_report_participant',
    '$ATTACKER $PLAYER_NAME, $PLANET_NAME $COORD. Ships/Defense systems $UNITS_COUNT Unit(s) lost: $UNITS_LOST_COUNT Weapons: $WEAPONS_TECH% Shielding: $SHIELDING_TECH% Armour: $ARMOUR_TECH%'
  );
INSERT INTO public.messages_ids ("type", "name", "content")
  VALUES(
    (SELECT id FROM messages_types WHERE type='fleets'),
    'fight_report_result_attacker_win',
    'Attacker has won the fight !'
  );
INSERT INTO public.messages_ids ("type", "name", "content")
  VALUES(
    (SELECT id FROM messages_types WHERE type='fleets'),
    'fight_report_result_defender_win',
    'Defender has won the fight !'
  );
INSERT INTO public.messages_ids ("type", "name", "content")
  VALUES(
    (SELECT id FROM messages_types WHERE type='fleets'),
    'fight_report_result_draw',
    'Combat ends in a draw !'
  );
INSERT INTO public.messages_ids ("type", "name", "content")
  VALUES(
    (SELECT id FROM messages_types WHERE type='fleets'),
    'fight_report_footer',
    'Plunder: $RESOURCES. Debris: $DEBRIS_FIELD. Unit(s) rebuilt: $UNITS_REBUILT.'
  );
