
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
    'transport_arrival_owner',
    'your fleet from $PLANET_NAME $COORD arrives at $PLANET_NAME $COORD ($PLAYER_NAME). The fleet deposits $RESOURCES'
  );
INSERT INTO public.messages_ids ("type", "name", "content")
  VALUES(
    (SELECT id FROM messages_types WHERE type='fleets'),
    'transport_arrival_receiver',
    'a fleet from $PLANET_NAME $COORD ($PLAYER_NAME) has reached the planet $PLANET_NAME $COORD. The fleet deposits $RESOURCES'
  );

INSERT INTO public.messages_ids ("type", "name", "content")
  VALUES(
    (SELECT id FROM messages_types WHERE type='fleets'),
    'acs_defend_arrival_owner',
    'your fleet from $PLANET_NAME $COORD arrives at $PLANET_NAME $COORD ($PLAYER_NAME) and parks in orbit for a watchful defense'
  );
INSERT INTO public.messages_ids ("type", "name", "content")
  VALUES(
    (SELECT id FROM messages_types WHERE type='fleets'),
    'acs_defend_arrival_receiver',
    'a fleet from $PLANET_NAME $COORD ($PLAYER_NAME) arrives at $PLANET_NAME $COORD. The fleet parks in orbit to defend us against any threats'
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
    'fleet_return_owner',
    'your fleet returns from $PLANET_NAME $COORD. The fleet deposits $RESOURCES'
  );
INSERT INTO public.messages_ids ("type", "name", "content")
  VALUES(
    (SELECT id FROM messages_types WHERE type='fleets'),
    'fleet_return_owner_harvest',
    'your fleet returns from $PLANET_NAME $COORD. The fleet deposits $RESOURCES'
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
    '$REPORT_HEADER $REPORT_RESOURCES $REPORT_ACTIVITY $REPORT_SHIPS $REPORT_DEFENSES $REPORT_BUILDINGS $REPORT_TECHNOLOGIES'
  );
INSERT INTO public.messages_ids ("type", "name", "content")
  VALUES(
    (SELECT id FROM messages_types WHERE type='fleets'),
    'espionage_report_header',
    'Resources on $PLANET_NAME $COORD ($PLAYER_NAME) at $DATE'
  );
INSERT INTO public.messages_ids ("type", "name", "content")
  VALUES(
    (SELECT id FROM messages_types WHERE type='fleets'),
    'espionage_report_resources',
    '$RESOURCE_NAME: $AMOUNT'
  );
INSERT INTO public.messages_ids ("type", "name", "content")
  VALUES(
    (SELECT id FROM messages_types WHERE type='fleets'),
    'espionage_report_some_activity',
    'your probe scan found abnormalities in the planet''s atmosphere indicating an activity in the last $INTERVAL minute(s).'
  );
INSERT INTO public.messages_ids ("type", "name", "content")
  VALUES(
    (SELECT id FROM messages_types WHERE type='fleets'),
    'espionage_report_no_activity',
    'your probe scan found no abnormalities in the planet''s atmosphere. An activity on this planet within the last hour can therefore almost be ruled out'
  );
INSERT INTO public.messages_ids ("type", "name", "content")
  VALUES(
    (SELECT id FROM messages_types WHERE type='fleets'),
    'espionage_report_ships',
    '$SHIP_NAME: $COUNT'
  );
INSERT INTO public.messages_ids ("type", "name", "content")
  VALUES(
    (SELECT id FROM messages_types WHERE type='fleets'),
    'espionage_report_defenses',
    '$DEFENSE_SYSTEM_NAME: $COUNT'
  );
INSERT INTO public.messages_ids ("type", "name", "content")
  VALUES(
    (SELECT id FROM messages_types WHERE type='fleets'),
    'espionage_report_buildings',
    '$BUILDING_NAME: $LEVEL'
  );
INSERT INTO public.messages_ids ("type", "name", "content")
  VALUES(
    (SELECT id FROM messages_types WHERE type='fleets'),
    'espionage_report_technologies',
    '$TECHNOLOGY_NAME: $LEVEL'
  );

INSERT INTO public.messages_ids ("type", "name", "content")
  VALUES(
    (SELECT id FROM messages_types WHERE type='fleets'),
    'fight_report',
    '$REPORT_HEADER $REPORT_ATTACKERS $REPORT_DEFENDERS $REPORT_RESULT $REPORT_FOOTER'
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
