
-- Generate the fight report header. The report will
-- be posted for the player specified in input.
CREATE OR REPLACE FUNCTION fleet_fight_report_header(fleet_id uuid, player_id uuid) RETURNS VOID AS $$
DECLARE
  target_kind text;

  target_name text;
  target_coordinates text;
  moment text;
BEGIN
  -- Select the name of the planet where the fight took
  -- place. It can either be a planet or a moon. We will
  -- also fetch the coordinates of the planet along with
  -- the date at which the fight occurred.
  SELECT target_type INTO target_kind FROM fleets WHERE id = fleet_id;
  IF NOT FOUND THEN
    RAISE EXCEPTION 'Invalid target type for fleet % in fleet fight report header operation', fleet_id;
  END IF;

  IF target_kind = 'planet' THEN
    SELECT
      p.name,
      concat_ws(':', p.galaxy,  p.solar_system,  p.position),
      to_char(f.arrival_time, 'MM-DD-YYYY HH24:MI:SS')
    INTO
      target_name,
      target_coordinates,
      moment
    FROM
      fleets AS f
      INNER JOIN planets AS p ON f.target = p.id
    WHERE
      id = fleet_id;
  END IF;

  IF target_kind = 'moon' THEN
    SELECT
      m.name,
      concat_ws(':', p.galaxy,  p.solar_system,  p.position),
      to_char(f.arrival_time, 'MM-DD-YYYY HH24:MI:SS')
    INTO
      target_name,
      target_coordinates,
      moment
    FROM
      fleets AS f
      INNER JOIN moons AS m ON f.target = m.id
      INNER JOIN planets AS p ON m.planet = p.id
    WHERE
      id = fleet_id;
  END IF;

  -- Create the message for the specified player.
  PERFORM create_message_for(player_id, 'fight_report_header', target_name, target_coordinates, moment);
END
$$ LANGUAGE plpgsql;

-- Generate the list of forces in a fleet fight and
-- post the created report to the specified player.
CREATE OR REPLACE FUNCTION fleet_fight_report_attacker_participant(fleet_id uuid, player_id uuid) RETURNS VOID AS $$
DECLARE
  source_kind text;

  source_name text;
  source_coordinates text;
  source_player_name text;

  
BEGIN
  -- Fetch information about the source planet of the
  -- participant.
  SELECT source_type INTO source_kind FROM fleets WHERE id = fleet_id;
  IF NOT FOUND THEN
    RAISE EXCEPTION 'Invalid source type for fleet % in fleet fight report attacker participant operation', fleet_id;
  END IF;

  IF target_kind = 'planet' THEN
    SELECT
      p.name,
      concat_ws(':', p.galaxy,  p.solar_system,  p.position),
      pl.name
    INTO
      source_name,
      source_coordinates,
      source_player_name
    FROM
      fleets AS f
      INNER JOIN planets AS p ON f.source = p.id
      INNER JOIN players AS pl ON p.player = pl.id
    WHERE
      id = fleet_id;
  END IF;

  IF target_kind = 'moon' THEN
    SELECT
      m.name,
      concat_ws(':', p.galaxy,  p.solar_system,  p.position),
      pl.name
    INTO
      source_name,
      source_coordinates,
      source_player_name
    FROM
      fleets AS f
      INNER JOIN moons AS m ON f.source = m.id
      INNER JOIN planets AS p ON m.planet = p.id
      INNER JOIN players AS pl ON p.player = pl.id
    WHERE
      id = fleet_id;
  END IF;

  -- Create the message for the specified player.
  PERFORM create_message_for(player_id, source_player_name, source_name, source_coordinates, 'TODO: S/D count', 'TODO: Units lost', 'TODO: Techs');
END
$$ LANGUAGE plpgsql;

INSERT INTO public.messages_ids ("type", "name", "content")
  VALUES(
    (SELECT id FROM messages_types WHERE type='fleets'),
    'fight_report_participant',
    '$PLAYER_NAME, $PLANET_NAME $COORD. Ships/Defense systems $UNITS_COUNT Unit(s) lost: $UNITS_LOST_COUNT Weapons: $WEAPONS_TECH% Shielding: $SHIELDING_TECH% Armour: $ARMOUR_TECH%'
  );

-- Generate the status of the fight. Depending on the
-- provided outcome the correct message will be posted
-- to the specified player.
CREATE OR REPLACE FUNCTION fleet_fight_report_status(outcome text, player_id uuid) RETURNS VOID AS $$
BEGIN
  -- The perspective of the fight is seen from the
  -- defender point of view.
  IF outcome = 'victory' THEN
    PERFORM create_message_for(player_id, 'fight_report_result_defender_win');
  END IF;

  IF outcome = 'draw' THEN
    PERFORM create_message_for(player_id, 'fight_report_result_draw');
  END IF;

  IF outcome = 'loss' THEN
    PERFORM create_message_for(player_id, 'fight_report_result_attacker_win');
  END IF;
END
$$ LANGUAGE plpgsql;

-- Generate a fight report for the specified fleet. The
-- report is generated for the owner of the fleet.
CREATE OR REPLACE FUNCTION fleet_fight_report(fleet_id uuid, outcome text) RETURNS VOID AS $$
DECLARE
  target_kind text;
  target_name text;
  target_coordinates text;
  target_player_name text;
  moment text;
  player_id uuid;
  player_name text;
  source_kind text;
  source_name text;
  source_coordinates text;
BEGIN
  SELECT target_type INTO target_kind FROM fleets WHERE id = fleet_id;
  IF NOT FOUND THEN
    RAISE EXCEPTION 'Invalid target type for fleet % in fleet fight report operation', fleet_id;
  END IF;

  SELECT source_type INTO source_kind FROM fleets WHERE id = fleet_id;
  IF NOT FOUND THEN
    RAISE EXCEPTION 'Invalid source type for fleet % in fleet fight report operation', fleet_id;
  END IF;

  -- Generate the report's header.
  IF target_kind = 'planet' THEN
    SELECT
      p.name,
      concat_ws(':', p.galaxy,  p.solar_system,  p.position),
      to_char(f.arrival_time, 'MM-DD-YYYY HH24:MI:SS'),
      f.player,
      pl.name
    INTO
      target_name,
      target_coordinates,
      moment,
      player_id,
      target_player_name
    FROM
      fleets AS f
      INNER JOIN planets AS p ON f.target = p.id
      INNER JOIN players AS pl ON p.player = pl.id
    WHERE
      id = fleet_id;
  END IF;

  IF target_kind = 'moon' THEN
    SELECT
      m.name,
      concat_ws(':', p.galaxy,  p.solar_system,  p.position),
      to_char(f.arrival_time, 'MM-DD-YYYY HH24:MI:SS'),
      f.player,
      pl.name
    INTO
      target_name,
      target_coordinates,
      moment,
      player_id,
      target_player_name
    FROM
      fleets AS f
      INNER JOIN moons AS m ON f.target = m.id
      INNER JOIN planets AS p ON m.planet = p.id
      INNER JOIN players AS pl ON p.player = pl.id
    WHERE
      id = fleet_id;
  END IF;

  PERFORM create_message_for(player_id, 'fight_report_header', target_name, target_coordinates, moment);

  -- Generate the forces that were involved in this fight
  -- for the attacker.
  IF source_kind = 'planet' THEN
    SELECT
      p.name,
      concat_ws(':', p.galaxy,  p.solar_system,  p.position),
      pl.name
    INTO
      source_name,
      source_coordinates,
      player_name
    FROM
      fleets AS f
      INNER JOIN planets AS p ON p.source = p.id
      INNER JOIN players AS pl ON f.player = pl.id
    WHERE
      id = fleet_id;
  END IF;

  IF source_kind = 'moon' THEN
    SELECT
      m.name,
      concat_ws(':', p.galaxy,  p.solar_system,  p.position),
      pl.player
    INTO
      source_name,
      source_coordinates,
      player_name
    FROM
      fleets AS f
      INNER JOIN moons AS m ON f.source = m.id
      INNER JOIN planets AS p ON m.planet = p.id
      INNER JOIN players AS pl ON f.player = pl.id
    WHERE
      id = fleet_id;
  END IF;

  PERFORM create_message_for(player_id, 'fleet_report_participant', player_name, source_name, source_coordinates, 'haha');
  PERFORM create_message_for(player_id, 'fleet_report_participant', target_player_name, target_name, target_coordinates, 'hoho');
  -- '$PLAYER_NAME, $PLANET_NAME $COORD. Ships/Defense systems $UNITS_COUNT Unit(s) lost: $UNITS_LOST_COUNT Weapons: $WEAPONS_TECH% Shielding: $SHIELDING_TECH% Armour: $ARMOUR_TECH%'
  -- TODO: Generate forces participating to the fight for
  -- each participant.

  -- Generate the result for the fight. Note that
  -- the outcome is expressed using the defender's
  -- perspective.
  IF outcome = 'victory' THEN
    PERFORM create_message_for(player_id, 'fight_report_result_defender_win');
  END IF;

  IF outcome = 'draw' THEN
    PERFORM create_message_for(player_id, 'fight_report_result_draw');
  END IF;

  IF outcome = 'loss' THEN
    PERFORM create_message_for(player_id, 'fight_report_result_attacker_win');
  END IF;


  -- Generate the footer of the fight report.
  -- PERFORM create_message_for(player_id, 'fight_report_footer', );
  -- 'Attacker has won the fight ! Plunder: $RESOURCES. Debris: $DEBRIS_FIELD. Unit(s) rebuilt: $UNITS_REBUILT.'

  -- TODO: Should handle this:
  -- create_message_for(player_id uuid, message_name text, VARIADIC args text[]) RETURNS VOID AS $$
  -- fight_report_participant
  -- fight_report_footer
END
$$ LANGUAGE plpgsql;
