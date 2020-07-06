
-- Perform the generation of the counter espionage
-- report: we need to gather information about the
-- spy and the spied planet/moon to do so.
-- Note that the report will always be generated
-- for the target of the fleet.
CREATE OR REPLACE FUNCTION generate_counter_espionage_report(fleet_id uuid, prob integer) RETURNS VOID AS $$
DECLARE
  spied_planet_kind text;
  spied_planet_name text;
  spied_coordinates text;
  spied_id uuid;

  spy_planet_kind text;
  spy_planet_name text;
  spy_coordinates text;
  spy_name text;
BEGIN
  -- Fetch information on the spy.
  SELECT source_type INTO spy_planet_kind FROM fleets WHERE id = fleet_id;
  IF NOT FOUND THEN
    RAISE EXCEPTION 'Invalid spy planet kind for fleet % in counter espionage operation', fleet_id;
  END IF;

  IF spy_planet_kind = 'planet' THEN
    SELECT
      p.name,
      concat_ws(':', p.galaxy,  p.solar_system,  p.position),
      pl.name
    INTO
      spy_planet_name,
      spy_coordinates,
      spy_name
    FROM
      fleets AS f
      INNER JOIN planets AS p ON f.source = p.id
      INNER JOIN players AS pl ON p.player = pl.id
    WHERE
      f.id = fleet_id;
  END IF;

  IF spy_planet_kind = 'moon' THEN
    SELECT
      m.name,
      concat_ws(':', p.galaxy,  p.solar_system,  p.position),
      pl.name
    INTO
      spy_planet_name,
      spy_coordinates,
      spy_name
    FROM
      fleets AS f
      INNER JOIN moons AS m ON f.source = m.id
      INNER JOIN planets AS p ON m.planet = p.id
      INNER JOIN players AS pl ON p.player = pl.id
    WHERE
      f.id = fleet_id;
  END IF;

  -- Fetch information on the spied player.
  SELECT target_type INTO spied_planet_kind FROM fleets WHERE id = fleet_id;
  IF NOT FOUND THEN
    RAISE EXCEPTION 'Invalid spied planet kind for fleet % in counter espionage operation', fleet_id;
  END IF;

  IF spied_planet_kind = 'planet' THEN
    SELECT
      p.name,
      concat_ws(':', p.galaxy,  p.solar_system,  p.position),
      pl.id
    INTO
      spied_planet_name,
      spied_coordinates,
      spied_id
    FROM
      fleets AS f
      INNER JOIN planets AS p ON f.target = p.id
      INNER JOIN players AS pl ON p.player = pl.id
    WHERE
      f.id = fleet_id;
  END IF;

  IF spied_planet_kind = 'moon' THEN
    SELECT
      m.name,
      concat_ws(':', p.galaxy,  p.solar_system,  p.position),
      pl.id
    INTO
      spied_planet_name,
      spied_coordinates,
      spied_id
    FROM
      fleets AS f
      INNER JOIN moons AS m ON f.target = m.id
      INNER JOIN planets AS p ON m.planet = p.id
      INNER JOIN players AS pl ON p.player = pl.id
    WHERE
      f.id = fleet_id;
  END IF;

  -- Perform the generation of the counter espionage
  -- report for the target of the fleet.
  PERFORM create_message_for(spied_id, 'counter_espionage_report', spy_planet_name, spy_coordinates, spy_name, spied_planet_name, spied_coordinates, prob);
END
$$ LANGUAGE plpgsql;

-- Generates the header of the espionage report
-- for the spy. It includes the time at which
-- the whole report was generated and info on
-- the planet that was spied.
CREATE OR REPLACE FUNCTION generate_header_report(player_id uuid, fleet_id uuid, report_id uuid) RETURNS VOID AS $$
DECLARE
  spied_planet_kind text;
  spied_planet_name text;
  spied_coordinates text;
  spied_name uuid;

  moment_text text;

  header_id uuid;
BEGIN
  -- Fetch information on the spied guy.
  SELECT target_type INTO spied_planet_kind FROM fleets WHERE id = fleet_id;
  IF NOT FOUND THEN
    RAISE EXCEPTION 'Invalid spied planet kind for fleet % in report header for espionage operation', fleet_id;
  END IF;

  IF spied_planet_kind = 'planet' THEN
    SELECT
      p.name,
      concat_ws(':', p.galaxy,  p.solar_system,  p.position),
      pl.name,
      to_char(f.arrival_time, 'MM-DD-YYYY HH24:MI:SS')
    INTO
      spied_planet_name,
      spied_coordinates,
      spied_name,
      moment_text
    FROM
      fleets AS f
      INNER JOIN planets AS p ON f.target = p.id
      INNER JOIN players AS pl ON p.player = pl.id
    WHERE
      f.id = fleet_id;
  END IF;

  IF spied_planet_kind = 'moon' THEN
    SELECT
      p.name,
      concat_ws(':', p.galaxy,  p.solar_system,  p.position),
      pl.name,
      to_char(f.arrival_time, 'MM-DD-YYYY HH24:MI:SS')
    INTO
      spied_planet_name,
      spied_coordinates,
      spied_name,
      moment_text
    FROM
      fleets AS f
      INNER JOIN moons AS m ON f.target = m.id
      INNER JOIN planets AS p ON m.planet = p.id
      INNER JOIN players AS pl ON p.player = pl.id
    WHERE
      f.id = fleet_id;
  END IF;

  -- Perform the creation of the message.
  SELECT * INTO header_id FROM create_message_for(player_id, 'espionage_report_header', spied_planet_name, spied_coordinates, spied_name, moment_text);

  -- Register this header as the first argument
  -- of the parent espionage report.
  INSERT INTO messages_arguments("message", "position", "argument") VALUES(report_id, 0, header_id);
END
$$ LANGUAGE plpgsql;

-- This function generates the part of the report
-- that indicates the resources on the target of
-- the espionage fleet.
CREATE OR REPLACE FUNCTION generate_resources_report(player_id uuid, fleet_id uuid, pOffset integer, report_id uuid) RETURNS integer AS $$
DECLARE
BEGIN
  -- TODO: Handle this.
  RETURN pOffset;
END
$$ LANGUAGE plpgsql;

-- This function generates the part of the report
-- that indicates the activity on a spied planet.
CREATE OR REPLACE FUNCTION generate_activity_report(player_id uuid, fleet_id uuid, pOffset integer, report_id uuid) RETURNS integer AS $$
DECLARE
  spied_planet_kind text;
  last_activity timestamp with time zone;

  moment timestamp with time zone;
  limit_for_activity timestamp with time zone;

  minutes_elapsed integer;

  activity_id uuid;
BEGIN
  -- Fetch information on the spied guy.
  SELECT target_type INTO spied_planet_kind FROM fleets WHERE id = fleet_id;
  IF NOT FOUND THEN
    RAISE EXCEPTION 'Invalid spied planet kind for fleet % in report activity for espionage operation', fleet_id;
  END IF;

  IF spied_planet_kind = 'planet' THEN
    SELECT
      -- TODO: We should add some mechanism to account for the activity of a planet.
      p.created_at,
      f.arrival_time
    INTO
      last_activity,
      moment
    FROM
      fleets AS f
      INNER JOIN planets AS p ON f.target = p.id
    WHERE
      f.id = fleet_id;
  END IF;

  IF spied_planet_kind = 'moon' THEN
    SELECT
      -- TODO: We should add some mechanism to account for the activity of a planet.
      p.created_at,
      f.arrival_time
    INTO
      last_activity,
      moment
    FROM
      fleets AS f
      INNER JOIN moons AS m ON f.target = m.id
      INNER JOIN planets AS p ON m.planet = p.id
    WHERE
      f.id = fleet_id;
  END IF;

  -- Compute whether the planet was active in the
  -- last hour.
  limit_for_activity = last_activity - interval '1 hour';

  IF limit_for_activity < moment THEN
    SELECT * INTO activity_id FROM create_message_for(player_id, 'espionage_report_no_activity', VARIADIC '{}'::text[]);
  END IF;

  IF limit_for_activity >= moment THEN
    SELECT EXTRACT(MINUTE FROM moment - last_activity) INTO minutes_elapsed;

    SELECT * INTO activity_id FROM create_message_for(player_id, 'espionage_report_some_activity', minutes_elapsed);
  END IF;

  -- Register this header as an argument of the
  -- parent espionage report.
  INSERT INTO messages_arguments("message", "position", "argument") VALUES(report_id, pOffset, activity_id);

  -- Return the next offset for a parameter of the
  -- parent espionage message.
  RETURN pOffset + 1;
END
$$ LANGUAGE plpgsql;

-- Similar to the `generate_resources_report` but
-- for the ships part of the espionage report.
CREATE OR REPLACE FUNCTION generate_ships_report(player_id uuid, fleet_id uuid, pOffset integer, report_id uuid) RETURNS integer AS $$
DECLARE
BEGIN
  -- TODO: Handle this.
  RETURN pOffset;
END
$$ LANGUAGE plpgsql;

-- Similar to the `generate_resources_report` but
-- for the defenses part of the espionage report.
CREATE OR REPLACE FUNCTION generate_defenses_report(player_id uuid, fleet_id uuid, pOffset integer, report_id uuid) RETURNS integer AS $$
DECLARE
BEGIN
  -- TODO: Handle this.
  RETURN pOffset;
END
$$ LANGUAGE plpgsql;

-- Similar to the `generate_resources_report` but
-- for the buildings part of the espionage report.
CREATE OR REPLACE FUNCTION generate_buildings_report(player_id uuid, fleet_id uuid, pOffset integer, report_id uuid) RETURNS integer AS $$
DECLARE
BEGIN
  -- TODO: Handle this.
  RETURN pOffset;
END
$$ LANGUAGE plpgsql;

-- Similar to the `generate_resources_report` but
-- for the technologies part of the espionage report.
CREATE OR REPLACE FUNCTION generate_technologies_report(player_id uuid, fleet_id uuid, pOffset integer, report_id uuid) RETURNS VOID AS $$
DECLARE
BEGIN
  -- TODO: Handle this.
END
$$ LANGUAGE plpgsql;

-- Script allowing to perform the registration of an
-- espionage report for the player owning the input
-- fleet with the level of information.
CREATE OR REPLACE FUNCTION espionage_report(fleet_id uuid, counter_espionage integer, info_level integer) RETURNS VOID AS $$
DECLARE
  report_id uuid := uuid_generate_v4();

  arg_count integer := 0;

  spy_planet_kind text;
  spy_id uuid;
BEGIN
  -- We need to generate the counter espionage report for
  -- the player that was targeted by the fleet.
  PERFORM generate_counter_espionage_report(fleet_id, counter_espionage);

  -- Retrieve the player's identifier from the fleet.
  SELECT source_type INTO spy_planet_kind FROM fleets WHERE id = fleet_id;
  IF NOT FOUND THEN
    RAISE EXCEPTION 'Invalid spy planet kind for fleet % in espionage operation', fleet_id;
  END IF;

  IF spy_planet_kind = 'planet' THEN
    SELECT
      pl.id
    INTO
      spy_id
    FROM
      fleets AS f
      INNER JOIN planets AS p ON f.source = p.id
      INNER JOIN players AS pl ON p.player = pl.id
    WHERE
      f.id = fleet_id;
  END IF;

  IF spy_planet_kind = 'moon' THEN
    SELECT
      pl.id
    INTO
      spy_id
    FROM
      fleets AS f
      INNER JOIN moons AS m ON f.source = m.id
      INNER JOIN planets AS p ON m.planet = p.id
      INNER JOIN players AS pl ON p.player = pl.id
    WHERE
      f.id = fleet_id;
  END IF;

  -- Based on the info level we will generate all the
  -- needed parts of the report. The parts were derived
  -- from this report:
  -- http://wiki.ogame.org/index.php/Tutorial:Espionage
  -- First we need to create the wrapper around the
  -- espionage message: this will be used by the parts
  -- of the report to link to the parent report.
  INSERT INTO messages_players(id, player, message)
    SELECT
      report_id,
      spy_id,
      mi.id
    FROM
      messages_ids AS mi
    WHERE
      mi.name = 'espionage_report';

  PERFORM generate_header_report(spy_id, fleet_id, report_id);

  -- Always display the resources.
  SELECT * INTO arg_count FROM generate_resources_report(spy_id, fleet_id, arg_count, report_id);

  -- Always display the activity.
  SELECT * INTO arg_count FROM generate_activity_report(spy_id, fleet_id, arg_count, report_id);

  IF info_level > 0 THEN
    SELECT * INTO arg_count FROM generate_ships_report(spy_id, fleet_id, arg_count, report_id);
  END IF;

  IF info_level > 1 THEN
    SELECT * INTO arg_count FROM generate_defenses_report(spy_id, fleet_id, arg_count, report_id);
  END IF;

  IF info_level > 2 THEN
    SELECT * INTO arg_count FROM generate_buildings_report(spy_id, fleet_id, arg_count, report_id);
  END IF;

  IF info_level > 3 THEN
    PERFORM generate_technologies_report(spy_id, fleet_id, arg_count, report_id);
  END IF;
END
$$ LANGUAGE plpgsql;