
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

  moment timestamp with time zone;
BEGIN
  -- Fetch information on the spy.
  SELECT
    source_type,
    arrival_time
  INTO
    spy_planet_kind,
    moment
  FROM
    fleets
  WHERE
    id = fleet_id;
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
  PERFORM create_message_for(spied_id, 'counter_espionage_report', moment, spy_planet_name, spy_coordinates, spy_name, spied_planet_name, spied_coordinates, prob::text);
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
  spied_name text;

  moment timestamp with time zone;
  moment_text text;

  header_id uuid;
BEGIN
  -- Fetch information on the spied guy.
  SELECT
    target_type,
    arrival_time
  INTO
    spied_planet_kind,
    moment
  FROM
    fleets
  WHERE
    id = fleet_id;
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
  SELECT * INTO header_id FROM create_message_for(player_id, 'espionage_report_header', moment, spied_planet_name, spied_coordinates, spied_name, moment_text);

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
  spied_planet_kind text;
  spied_planet_id uuid;

  moment timestamp with time zone;

  temprow record;
  resource_msg_id uuid;
BEGIN
  -- To generate the resources, we need to create
  -- a message for each resource existing on the
  -- planet that is the destination of the fleet.
  -- To do so, we will first select all the res,
  -- and then iterate over each one of them.
  -- For each one, we will have to create a new
  -- message of type `espionage_report_resources`
  -- with arguments the name of the resource and
  -- its amount.
  -- We will also need to update the resources
  -- on the planet to be sure that it's up to
  -- date. We consider that fleets and other
  -- actions have already be handled for the
  -- planet at the moment of calling this script
  -- though.

  -- Gather information about the spied planet.
  SELECT
    target_type,
    target,
    arrival_time
  INTO
    spied_planet_kind,
    spied_planet_id,
    moment
  FROM
    fleets
  WHERE
    id = fleet_id;

  IF NOT FOUND THEN
    RAISE EXCEPTION 'Invalid spied planet kind for fleet % in report ships for espionage operation', fleet_id;
  END IF;

  -- Update the resources on the planet.
  PERFORM update_resources_for_planet_to_time(spied_planet_id, moment);

  -- Traverse all the resources existing on
  -- the target planet.
  IF spied_planet_kind = 'planet' THEN
    FOR temprow IN
      SELECT
        pr.res
      FROM
        planets_resources AS pr
        INNER JOIN resources AS r ON pr.res = r.id
      WHERE
        pr.planet = spied_planet_id
        AND r.movable = 'true'
    LOOP
      -- Create the message representing this resource.
      resource_msg_id := uuid_generate_v4();

      INSERT INTO messages_players("id", "player", "message", "created_at")
        SELECT
          resource_msg_id,
          player_id,
          mi.id,
          moment
        FROM
          messages_ids AS mi
        WHERE
          mi.name = 'espionage_report_resources';

      -- Register this message as an argument of
      -- the main espionage report.
      INSERT INTO messages_arguments("message", "position", "argument") VALUES(report_id, pOffset, resource_msg_id);
      pOffset := pOffset + 1;

      -- Generate the argument for this message.
      INSERT INTO messages_arguments("message", "position", "argument") VALUES(resource_msg_id, 0, temprow.res);

      INSERT INTO messages_arguments("message", "position", "argument")
        SELECT
          resource_msg_id,
          1,
          amount
        FROM
          planets_resources
        WHERE
          planet = spied_planet_id
          AND res = temprow.res;
    END LOOP;
  END IF;

  IF spied_planet_kind = 'moon' THEN
    FOR temprow IN
      SELECT
        mr.res
      FROM
        moons_resources AS mr
        INNER JOIN resources AS r ON mr.res = r.id
      WHERE
        mr.moon = spied_planet_id
        AND r.movable = 'true'
    LOOP
      -- Create the message representing this resource.
      resource_msg_id := uuid_generate_v4();

      INSERT INTO messages_players("id", "player", "message", "created_at")
        SELECT
          resource_msg_id,
          player_id,
          mi.id,
          moment
        FROM
          messages_ids AS mi
        WHERE
          mi.name = 'espionage_report_resources';

      -- Register this message as an argument of
      -- the main espionage report.
      INSERT INTO messages_arguments("message", "position", "argument") VALUES(report_id, pOffset, resource_msg_id);
      pOffset := pOffset + 1;

      -- Generate the argument for this message.
      INSERT INTO messages_arguments("message", "position", "argument") VALUES(resource_msg_id, 0, temprow.res);

      INSERT INTO messages_arguments("message", "position", "argument")
        SELECT
          resource_msg_id,
          1,
          amount
        FROM
          moons_resources
        WHERE
          moon = spied_planet_id
          AND res = temprow.res;
    END LOOP;
  END IF;

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
  SELECT target_type, arrival_time INTO spied_planet_kind, moment FROM fleets WHERE id = fleet_id;
  IF NOT FOUND THEN
    RAISE EXCEPTION 'Invalid spied planet kind for fleet % in report activity for espionage operation', fleet_id;
  END IF;

  IF spied_planet_kind = 'planet' THEN
    SELECT
      p.last_activity,
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
      m.last_activity,
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
  limit_for_activity = last_activity + interval '1 hour';

  IF limit_for_activity < moment THEN
    SELECT * INTO activity_id FROM create_message_for(player_id, 'espionage_report_no_activity', moment, VARIADIC '{}'::text[]);
  END IF;

  IF limit_for_activity >= moment THEN
    SELECT EXTRACT(MINUTE FROM moment - last_activity) INTO minutes_elapsed;

    SELECT * INTO activity_id FROM create_message_for(player_id, 'espionage_report_some_activity', moment, minutes_elapsed::text);
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
  spied_planet_kind text;
  spied_planet_id uuid;

  moment timestamp with time zone;

  temprow record;
  ship_msg_id uuid;
BEGIN
  -- To generate the ships, we need to create a
  -- message for each ship existing on the planet
  -- that is the destination of the fleet.
  -- To do so, we will first select all the ships,
  -- and then iterate over each one of them.
  -- For each one, we will have to create a new
  -- message of type `espionage_report_ships`
  -- with arguments the name of the ship and
  -- its count.

  -- Gather information about the spied planet.
  SELECT
    target_type,
    target,
    arrival_time
  INTO
    spied_planet_kind,
    spied_planet_id,
    moment
  FROM
    fleets
  WHERE
    id = fleet_id;

  IF NOT FOUND THEN
    RAISE EXCEPTION 'Invalid spied planet kind for fleet % in report ships for espionage operation', fleet_id;
  END IF;

  -- Traverse all the ships existing on
  -- the target planet.
  IF spied_planet_kind = 'planet' THEN
    FOR temprow IN
      SELECT ship FROM planets_ships WHERE planet = spied_planet_id AND count > 0
    LOOP
      -- Create the message representing this ship.
      ship_msg_id := uuid_generate_v4();

      INSERT INTO messages_players("id", "player", "message", "created_at")
        SELECT
          ship_msg_id,
          player_id,
          mi.id,
          moment
        FROM
          messages_ids AS mi
        WHERE
          mi.name = 'espionage_report_ships';

      -- Register this message as an argument of
      -- the main espionage report.
      INSERT INTO messages_arguments("message", "position", "argument") VALUES(report_id, pOffset, ship_msg_id);
      pOffset := pOffset + 1;

      -- Generate the argument for this message.
      INSERT INTO messages_arguments("message", "position", "argument") VALUES(ship_msg_id, 0, temprow.ship);

      INSERT INTO messages_arguments("message", "position", "argument")
        SELECT
          ship_msg_id,
          1,
          count
        FROM
          planets_ships
        WHERE
          planet = spied_planet_id
          AND ship = temprow.ship;
    END LOOP;
  END IF;

  IF spied_planet_kind = 'moon' THEN
    FOR temprow IN
      SELECT ship FROM moons_ships WHERE moon = spied_planet_id AND count > 0
    LOOP
      -- Create the message representing this ship.
      ship_msg_id := uuid_generate_v4();

      INSERT INTO messages_players("id", "player", "message", "created_at")
        SELECT
          resourceship_msg_id_msg_id,
          player_id,
          mi.id,
          moment
        FROM
          messages_ids AS mi
        WHERE
          mi.name = 'espionage_report_ships';

      -- Register this message as an argument of
      -- the main espionage report.
      INSERT INTO messages_arguments("message", "position", "argument") VALUES(report_id, pOffset, ship_msg_id);
      pOffset := pOffset + 1;

      -- Generate the argument for this message.
      INSERT INTO messages_arguments("message", "position", "argument") VALUES(ship_msg_id, 0, temprow.ship);

      INSERT INTO messages_arguments("message", "position", "argument")
        SELECT
          ship_msg_id,
          1,
          amount
        FROM
          moons_ships
        WHERE
          moon = spied_planet_id
          AND ship = temprow.ship;
    END LOOP;
  END IF;

  RETURN pOffset;
END
$$ LANGUAGE plpgsql;

-- Similar to the `generate_resources_report` but
-- for the defenses part of the espionage report.
CREATE OR REPLACE FUNCTION generate_defenses_report(player_id uuid, fleet_id uuid, pOffset integer, report_id uuid) RETURNS integer AS $$
DECLARE
  spied_planet_kind text;
  spied_planet_id uuid;

  moment timestamp with time zone;

  temprow record;
  defense_msg_id uuid;
BEGIN
  -- To generate the defenses, we need to create
  -- a message for each defense existing on the
  -- planet that is the destination of the fleet.
  -- To do so, we will first select all defense
  -- systems, and then iterate over each one of
  -- them.
  -- For each one, we will have to create a new
  -- message of type `espionage_report_defenses`
  -- with arguments the name of the defense and
  -- its count.

  -- Gather information about the spied planet.
  SELECT
    target_type,
    target,
    arrival_time
  INTO
    spied_planet_kind,
    spied_planet_id,
    moment
  FROM
    fleets
  WHERE
    id = fleet_id;

  IF NOT FOUND THEN
    RAISE EXCEPTION 'Invalid spied planet kind for fleet % in report defenses for espionage operation', fleet_id;
  END IF;

  -- Traverse all the defenses existing on
  -- the target planet.
  IF spied_planet_kind = 'planet' THEN
    FOR temprow IN
      SELECT defense FROM planets_defenses WHERE planet = spied_planet_id AND count > 0
    LOOP
      -- Create the message representing this defense.
      defense_msg_id := uuid_generate_v4();

      INSERT INTO messages_players("id", "player", "message", "created_at")
        SELECT
          defense_msg_id,
          player_id,
          mi.id,
          moment
        FROM
          messages_ids AS mi
        WHERE
          mi.name = 'espionage_report_defenses';

      -- Register this message as an argument of
      -- the main espionage report.
      INSERT INTO messages_arguments("message", "position", "argument") VALUES(report_id, pOffset, defense_msg_id);
      pOffset := pOffset + 1;

      -- Generate the argument for this message.
      INSERT INTO messages_arguments("message", "position", "argument") VALUES(defense_msg_id, 0, temprow.defense);

      INSERT INTO messages_arguments("message", "position", "argument")
        SELECT
          defense_msg_id,
          1,
          count
        FROM
          planets_defenses
        WHERE
          planet = spied_planet_id
          AND defense = temprow.defense;
    END LOOP;
  END IF;

  IF spied_planet_kind = 'moon' THEN
    FOR temprow IN
      SELECT defense FROM moons_defenses WHERE moon = spied_planet_id AND count > 0
    LOOP
      -- Create the message representing this defense.
      defense_msg_id := uuid_generate_v4();

      INSERT INTO messages_players("id", "player", "message", "created_at")
        SELECT
          defense_msg_id,
          player_id,
          mi.id,
          moment
        FROM
          messages_ids AS mi
        WHERE
          mi.name = 'espionage_report_defenses';

      -- Register this message as an argument of
      -- the main espionage report.
      INSERT INTO messages_arguments("message", "position", "argument") VALUES(report_id, pOffset, defense_msg_id);
      pOffset := pOffset + 1;

      -- Generate the argument for this message.
      INSERT INTO messages_arguments("message", "position", "argument") VALUES(defense_msg_id, 0, temprow.defense);

      INSERT INTO messages_arguments("message", "position", "argument")
        SELECT
          defense_msg_id,
          1,
          amount
        FROM
          moons_defenses
        WHERE
          moon = spied_planet_id
          AND defense = temprow.defense;
    END LOOP;
  END IF;

  RETURN pOffset;
END
$$ LANGUAGE plpgsql;

-- Similar to the `generate_resources_report` but
-- for the buildings part of the espionage report.
CREATE OR REPLACE FUNCTION generate_buildings_report(player_id uuid, fleet_id uuid, pOffset integer, report_id uuid) RETURNS integer AS $$
DECLARE
  spied_planet_kind text;
  spied_planet_id uuid;

  moment timestamp with time zone;

  temprow record;
  building_msg_id uuid;
BEGIN
  -- The creation of the buildings in the report
  -- is similar to what happens for the defenses
  -- and ships and resources.

  -- Gather information about the spied planet.
  SELECT
    target_type,
    target,
    arrival_time
  INTO
    spied_planet_kind,
    spied_planet_id,
    moment
  FROM
    fleets
  WHERE
    id = fleet_id;

  IF NOT FOUND THEN
    RAISE EXCEPTION 'Invalid spied planet kind for fleet % in report buildings for espionage operation', fleet_id;
  END IF;

  -- Traverse all the buildings existing on
  -- the target planet.
  IF spied_planet_kind = 'planet' THEN
    FOR temprow IN
      SELECT building FROM planets_buildings WHERE planet = spied_planet_id AND level > 0
    LOOP
      -- Create the message representing this building.
      building_msg_id := uuid_generate_v4();

      INSERT INTO messages_players("id", "player", "message", "created_at")
        SELECT
          building_msg_id,
          player_id,
          mi.id,
          moment
        FROM
          messages_ids AS mi
        WHERE
          mi.name = 'espionage_report_buildings';

      -- Register this message as an argument of
      -- the main espionage report.
      INSERT INTO messages_arguments("message", "position", "argument") VALUES(report_id, pOffset, building_msg_id);
      pOffset := pOffset + 1;

      -- Generate the argument for this message.
      INSERT INTO messages_arguments("message", "position", "argument") VALUES(building_msg_id, 0, temprow.building);

      INSERT INTO messages_arguments("message", "position", "argument")
        SELECT
          building_msg_id,
          1,
          level
        FROM
          planets_buildings
        WHERE
          planet = spied_planet_id
          AND building = temprow.building;
    END LOOP;
  END IF;

  IF spied_planet_kind = 'moon' THEN
    FOR temprow IN
      SELECT building FROM moons_buildings WHERE moon = spied_planet_id AND level > 0
    LOOP
      -- Create the message representing this building.
      building_msg_id := uuid_generate_v4();

      INSERT INTO messages_players("id", "player", "message", "created_at")
        SELECT
          building_msg_id,
          player_id,
          mi.id,
          moment
        FROM
          messages_ids AS mi
        WHERE
          mi.name = 'espionage_report_buildings';

      -- Register this message as an argument of
      -- the main espionage report.
      INSERT INTO messages_arguments("message", "position", "argument") VALUES(report_id, pOffset, building_msg_id);
      pOffset := pOffset + 1;

      -- Generate the argument for this message.
      INSERT INTO messages_arguments("message", "position", "argument") VALUES(building_msg_id, 0, temprow.building);

      INSERT INTO messages_arguments("message", "position", "argument")
        SELECT
          building_msg_id,
          1,
          amount
        FROM
          moons_buildings
        WHERE
          moon = spied_planet_id
          AND building = temprow.building;
    END LOOP;
  END IF;

  RETURN pOffset;
END
$$ LANGUAGE plpgsql;

-- Similar to the `generate_resources_report` but
-- for the technologies part of the espionage report.
CREATE OR REPLACE FUNCTION generate_technologies_report(player_id uuid, fleet_id uuid, pOffset integer, report_id uuid) RETURNS VOID AS $$
DECLARE
  spied_planet_kind text;
  spied_planet_id uuid;
  spied_player_id uuid;

  moment timestamp with time zone;

  temprow record;
  tech_msg_id uuid;
BEGIN
  -- The creation of the technologies in the
  -- report is similar to what happens for
  -- ships, defenses or buildings.
  -- The difference though is that the techs
  -- are identical whether the fleet spies a
  -- planet or a moon so we have to fetch the
  -- player owning the celestial body instead.

  -- Gather information about the spied planet
  -- and player.
  SELECT
    target_type,
    target,
    arrival_time
  INTO
    spied_planet_kind,
    spied_planet_id,
    moment
  FROM
    fleets
  WHERE
    id = fleet_id;

  IF NOT FOUND THEN
    RAISE EXCEPTION 'Invalid spied planet kind for fleet % in report technologies for espionage operation', fleet_id;
  END IF;

  -- Traverse all the technologies existing on
  -- the target planet.
  IF spied_planet_kind = 'planet' THEN
    SELECT
      pl.id
    INTO
      spied_player_id
    FROM
      fleets AS f
      INNER JOIN planets AS p ON p.id = f.target
      INNER JOIN players AS pl ON pl.id = p.player
    WHERE
      f.id = fleet_id;
  END IF;

  IF spied_planet_kind = 'moon' THEN
    SELECT
      pl.id
    INTO
      spied_player_id
    FROM
      fleets AS f
      INNER JOIN moons AS m ON m.id = f.target
      INNER JOIN planets AS p ON p.id = m.planet
      INNER JOIN players AS pl ON pl.id = p.player
    WHERE
      f.id = fleet_id;
  END IF;

  -- Traverse all technologies developed by this player.
  -- Note that we will only consider techs that have a
  -- level greater than `0`.
  FOR temprow IN
    SELECT technology FROM players_technologies WHERE technology = spied_player_id AND level > 0
  LOOP
    -- Create the message representing this building.
    tech_msg_id := uuid_generate_v4();

    INSERT INTO messages_players("id", "player", "message", "created_at")
      SELECT
        tech_msg_id,
        player_id,
        mi.id,
        moment
      FROM
        messages_ids AS mi
      WHERE
        mi.name = 'espionage_report_buildings';

    -- Register this message as an argument of
    -- the main espionage report.
    INSERT INTO messages_arguments("message", "position", "argument") VALUES(report_id, pOffset, tech_msg_id);
    pOffset := pOffset + 1;

    -- Generate the argument for this message.
    INSERT INTO messages_arguments("message", "position", "argument") VALUES(tech_msg_id, 0, temprow.technology);

    INSERT INTO messages_arguments("message", "position", "argument")
      SELECT
        tech_msg_id,
        1,
        level
      FROM
        players_technologies
      WHERE
        player = spied_player_id
        AND technology = temprow.technology;
  END LOOP;
END
$$ LANGUAGE plpgsql;

-- Script allowing to perform the registration of an
-- espionage report for the player owning the input
-- fleet with the level of information.
CREATE OR REPLACE FUNCTION espionage_report(fleet_id uuid, counter_espionage integer, info_level integer) RETURNS VOID AS $$
DECLARE
  report_id uuid := uuid_generate_v4();

  arg_count integer := 1;

  spy_planet_kind text;
  spy_planet_id uuid;
  spy_id uuid;

  moment timestamp with time zone;
BEGIN
  -- We need to generate the counter espionage report for
  -- the player that was targeted by the fleet.
  PERFORM generate_counter_espionage_report(fleet_id, counter_espionage);

  -- Retrieve the player's identifier from the fleet.
  SELECT source_type, target, arrival_time INTO spy_planet_kind, spy_planet_id, moment FROM fleets WHERE id = fleet_id;
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
  INSERT INTO messages_players("id", "player", "message", "created_at")
    SELECT
      report_id,
      spy_id,
      mi.id,
      moment
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

  -- Update the last activity on the target planet. This
  -- needs to be done after the generation of the report
  -- so as not to get a false indication in the activity
  -- status.
  IF spy_planet_kind = 'target' THEN
    UPDATE planets SET last_activity = moment WHERE id = spy_planet_id;
  END IF;

  IF spy_planet_kind = 'moon' THEN
    UPDATE moons SET last_activity = moment WHERE id = spy_planet_id;
  END IF;
END
$$ LANGUAGE plpgsql;