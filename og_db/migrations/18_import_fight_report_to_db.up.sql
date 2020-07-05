
-- Generate the fight report header. The report will
-- be posted for the player specified in input.
CREATE OR REPLACE FUNCTION fleet_fight_report_header(player_id uuid, planet_id uuid, planet_kind text, moment timestamp with time zone) RETURNS uuid AS $$
DECLARE
  planet_name text;
  planet_coordinates text;
  moment_text text;
BEGIN
  IF planet_kind = 'planet' THEN
    SELECT
      name,
      concat_ws(':', galaxy,  solar_system,  position),
      to_char(moment, 'MM-DD-YYYY HH24:MI:SS')
    INTO
      planet_name,
      planet_coordinates,
      moment_text
    FROM
      planets
    WHERE
      id = planet_id;
  END IF;

  IF planet_kind = 'moon' THEN
    SELECT
      m.name,
      concat_ws(':', p.galaxy,  p.solar_system,  p.position),
      to_char(moment, 'MM-DD-YYYY HH24:MI:SS')
    INTO
      planet_name,
      planet_coordinates,
      moment_text
    FROM
      moons AS m
      INNER JOIN planets AS p ON m.planet = p.id
    WHERE
      m.id = planet_id;
  END IF;

  -- Create the message for the specified player.
  RETURN create_message_for(player_id, 'fight_report_header', planet_name, planet_coordinates, moment_text);
END
$$ LANGUAGE plpgsql;

-- Generate the list of forces in a fleet fight and
-- post the created report to the specified player.
CREATE OR REPLACE FUNCTION fleet_fight_report_outsider_participant(player_id uuid, fleet_id uuid, remains json) RETURNS uuid AS $$
DECLARE
  source_kind text;

  source_name text;
  source_coordinates text;
  source_player_id uuid;
  source_player_name text;
  hostile_status text;

  weapon_tech text;
  shielding_tech text;
  armour_tech text;

  units_count integer;
  units_lost integer;
BEGIN
  -- Fetch information about the source planet of the
  -- participant.
  SELECT source_type INTO source_kind FROM fleets WHERE id = fleet_id;
  IF NOT FOUND THEN
    RAISE EXCEPTION 'Invalid source type for fleet % in fleet fight report attacker participant operation', fleet_id;
  END IF;

  IF source_kind = 'planet' THEN
    SELECT
      p.name,
      concat_ws(':', p.galaxy,  p.solar_system,  p.position),
      f.player,
      pl.name,
      fo.hostile
    INTO
      source_name,
      source_coordinates,
      source_player_id,
      source_player_name,
      hostile_status
    FROM
      fleets AS f
      INNER JOIN fleets_objectives AS fo ON f.objective = fo.id
      INNER JOIN planets AS p ON f.source = p.id
      INNER JOIN players AS pl ON p.player = pl.id
    WHERE
      f.id = fleet_id;
  END IF;

  IF source_kind = 'moon' THEN
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

  -- Fetch the level of the technologies for the player.
  SELECT
    concat_ws('%', 10 * level)
  INTO
    weapon_tech
  FROM
    players_technologies AS pt
    INNER JOIN technologies AS t ON pt.technology = t.id
  WHERE
    player = source_player_id
    AND t.name = 'weapons';

  SELECT
    concat_ws('%', 10 * level)
  INTO
    shielding_tech
  FROM
    players_technologies AS pt
    INNER JOIN technologies AS t ON pt.technology = t.id
  WHERE
    player = source_player_id
    AND t.name = 'shielding';

  SELECT
    concat_ws('%', 10 * level)
  INTO
    armour_tech
  FROM
    players_technologies AS pt
    INNER JOIN technologies AS t ON pt.technology = t.id
  WHERE
    player = source_player_id
    AND t.name = 'armour';

  -- Compute the total amount of ships brought by
  -- this participant into the fight. It is just
  -- the number of ships sent to attack.
  SELECT COALESCE(SUM(count), 0) INTO units_count FROM fleets_ships WHERE fleet = fleet_id;

  -- Compute the total amount of resources lost
  -- by this attacker: it is computed as the
  -- difference between the initial fleet and
  -- the remaining fleet, only considering the
  -- resources that can be dispersed.
  WITH rs AS (
    SELECT
      t.fleet,
      t.ship,
      t.count
    FROM
      json_to_recordset(remains) AS t(fleet uuid, ship uuid, count integer)
    WHERE
      t.fleet = fleet_id
    )
  SELECT
    COALESCE(SUM((COALESCE(fs.count, rs.count) - rs.count) * sc.cost), 0)
  INTO
    units_lost
  FROM
    rs
    LEFT JOIN fleets_ships AS fs ON rs.ship = fs.ship
    INNER JOIN ships_costs AS sc ON rs.ship = sc.element
    INNER JOIN resources AS r ON sc.res = r.id
  WHERE
    rs.fleet = fleet_id
    AND r.dispersable = 'true';

  -- Create the message for the specified player.
  RETURN create_message_for(player_id, 'fight_report_participant', hostile_status::text, source_player_name, source_name, source_coordinates, units_count::text, units_lost::text, weapon_tech, shielding_tech, armour_tech);
END
$$ LANGUAGE plpgsql;

-- Generate the list of forces in a fleet fight in
-- the case of the defender. Note that this should
-- only be used for the owner of the planet where
-- the fleet is fighting as it includes the sum
-- of the defense systems existing on the planet.
CREATE OR REPLACE FUNCTION fleet_fight_report_indigenous_participant(player_id uuid, planet_id uuid, kind text, ships_remains json, def_remains json) RETURNS uuid AS $$
DECLARE
  target_player_id uuid;
  target_name text;
  target_coordinates text;
  target_player_name text;

  weapon_tech text;
  shielding_tech text;
  armour_tech text;

  ships_count integer;
  defenses_count integer;
  ships_lost integer := 0;
  defenses_lost integer := 0;
BEGIN
  -- Fetch information about the target planet of the
  -- participant: this is what is really meant by the
  -- 'indigenous' term. Compared to a regular attacker
  -- or defender it also includes the defenses that
  -- exist on the planet where the fight is happening.
  IF kind != 'planet' AND kind != 'moon' THEN
    RAISE EXCEPTION 'Invalid kind % specified for % in indigenous participant fight report', kind, planet_id;
  END IF;

  IF kind = 'planet' THEN
    SELECT
      pl.id,
      p.name,
      concat_ws(':', p.galaxy,  p.solar_system,  p.position),
      pl.name
    INTO
      target_player_id,
      target_name,
      target_coordinates,
      target_player_name
    FROM
      planets AS p
      INNER JOIN players AS pl ON p.player = pl.id
    WHERE
      p.id = planet_id;
  END IF;

  IF kind = 'moon' THEN
    SELECT
      pl.id,
      m.name,
      concat_ws(':', p.galaxy,  p.solar_system,  p.position),
      pl.name
    INTO
      target_player_id,
      target_name,
      target_coordinates,
      target_player_name
    FROM
      moons AS m
      INNER JOIN planets AS p ON m.planet = p.id
      INNER JOIN players AS pl ON p.player = pl.id
    WHERE
      m.id = planet_id;
  END IF;

  -- Fetch the level of the technologies for the player.
  SELECT
    concat_ws('%', 10 * level)
  INTO
    weapon_tech
  FROM
    players_technologies AS pt
    INNER JOIN technologies AS t ON pt.technology = t.id
  WHERE
    player = target_player_id
    AND t.name = 'weapons';

  SELECT
    concat_ws('%', 10 * level)
  INTO
    shielding_tech
  FROM
    players_technologies AS pt
    INNER JOIN technologies AS t ON pt.technology = t.id
  WHERE
    player = target_player_id
    AND t.name = 'shielding';

  SELECT
    concat_ws('%', 10 * level)
  INTO
    armour_tech
  FROM
    players_technologies AS pt
    INNER JOIN technologies AS t ON pt.technology = t.id
  WHERE
    player = target_player_id
    AND t.name = 'armour';

  -- The total amount of ships used by this player
  -- is the amount of ships deployed on the planet
  -- (or moon).
  IF kind = 'planet' THEN
    SELECT COALESCE(SUM(count), 0) INTO ships_count FROM planets_ships WHERE planet = planet_id;

    SELECT COALESCE(SUM(count), 0) INTO defenses_count FROM planets_defenses WHERE planet = planet_id;
  END IF;

  IF kind = 'moon' THEN
    SELECT COALESCE(SUM(count), 0) INTO ships_count FROM moons_ships WHERE moon = planet_id;

    SELECT COALESCE(SUM(count), 0) INTO defenses_count FROM moons_defenses WHERE moon = planet_id;
  END IF;

  -- Compute the total amount of resources lost
  -- by this player: it is computed as the diff
  -- between the initial fleet and the remaining
  -- fleet, only considering the resources that
  -- can be dispersed. Note that we also need to
  -- include the defenses in the total.
  IF kind = 'planet' THEN
    WITH rs AS (
      SELECT
        t.ship,
        t.count
      FROM
        json_to_recordset(ships_remains) AS t(ship uuid, count integer)
      )
    SELECT
      COALESCE(SUM((ps.count - COALESCE(rs.count, ps.count)) * sc.cost), 0)
    INTO
      ships_lost
    FROM
      planets_ships AS ps
      LEFT JOIN rs ON ps.ship = rs.ship
      INNER JOIN ships_costs AS sc ON ps.ship = sc.element
      INNER JOIN resources AS r ON sc.res = r.id
    WHERE
      ps.planet = planet_id
      AND r.dispersable = 'true';

    WITH rd AS (
      SELECT
        t.defense,
        t.count
      FROM
        json_to_recordset(def_remains) AS t(defense uuid, count integer)
      )
    SELECT
      COALESCE(SUM((pd.count - COALESCE(rd.count, pd.count)) * dc.cost), 0)
    INTO
      defenses_lost
    FROM
      planets_defenses AS pd
      LEFT JOIN rd ON pd.defense = rd.defense
      INNER JOIN defenses_costs AS dc ON pd.defense = dc.element
      INNER JOIN resources AS r ON dc.res = r.id
    WHERE
      pd.planet = planet_id
      AND r.dispersable = 'true';
  END IF;

  IF kind = 'moon' THEN
    WITH rs AS (
      SELECT
        t.ship,
        t.count
      FROM
        json_to_recordset(ships_remains) AS t(ship uuid, count integer)
      )
    SELECT
      COALESCE(SUM((ms.count - COALESCE(rs.count, ms.count)) * sc.cost), 0)
    INTO
      ships_lost
    FROM
      moons_ships AS ms
      LEFT JOIN rs ON ms.ship = rs.ship
      INNER JOIN ships_costs AS sc ON ms.ship = sc.element
      INNER JOIN resources AS r ON sc.res = r.id
    WHERE
      ms.moon = planet_id
      AND r.dispersable = 'true';

    WITH rd AS (
      SELECT
        t.defense,
        t.count
      FROM
        json_to_recordset(def_remains) AS t(defense uuid, count integer)
      )
    SELECT
      COALESCE(SUM((md.count - COALESCE(rd.count, md.count)) * dc.cost), 0)
    INTO
      defenses_lost
    FROM
      moons_defenses AS md
      LEFT JOIN rd ON md.defense = rd.defense
      INNER JOIN defenses_costs AS dc ON md.defense = dc.element
      INNER JOIN resources AS r ON dc.res = r.id
    WHERE
      md.moon = planet_id
      AND r.dispersable = 'true';
  END IF;

  -- Create the message for the specified player.
  -- Note that the indigenous participant to a
  -- fight is obviously not hostile.
  RETURN create_message_for(player_id, 'fight_report_participant', 'false', target_player_name, target_name, target_coordinates, (ships_count + defenses_count)::text, (ships_lost + defenses_lost)::text, weapon_tech, shielding_tech, armour_tech);
END
$$ LANGUAGE plpgsql;

-- Generate the status of the fight. Depending on the
-- provided outcome the correct message will be posted
-- to the specified player.
CREATE OR REPLACE FUNCTION fleet_fight_report_status(player_id uuid, outcome text) RETURNS uuid AS $$
BEGIN
  -- The perspective of the fight is seen from the
  -- defender point of view. Note that as this type
  -- of message do not take any argument we provide
  -- a dummy array so that the variadic condition
  -- is satisfied.
  IF outcome = 'victory' THEN
    RETURN create_message_for(player_id, 'fight_report_result_defender_win', VARIADIC '{}'::text[]);
  END IF;

  IF outcome = 'draw' THEN
    RETURN create_message_for(player_id, 'fight_report_result_draw', VARIADIC '{}'::text[]);
  END IF;

  IF outcome = 'loss' THEN
    RETURN create_message_for(player_id, 'fight_report_result_attacker_win', VARIADIC '{}'::text[]);
  END IF;
END
$$ LANGUAGE plpgsql;

-- Generate the report indicating the final result
-- of the fight including the generated debris field
-- and the plundered resources if any.
CREATE OR REPLACE FUNCTION fleet_fight_report_footer(player_id uuid, pillage json, debris json, rebuilt integer) RETURNS uuid AS $$
DECLARE
  resources_pillaged text;
  resources_dispersed text;
BEGIN
  -- Generate the plundered resources string.
  WITH rp AS (
    SELECT
      t.resource,
      t.amount
    FROM
      json_to_recordset(pillage) AS t(resource uuid, amount numeric(15, 5))
    )
  SELECT
    string_agg(concat_ws(' unit(s) of ', floor(COALESCE(rp.amount, 0))::integer::text, r.name), ', ')
  INTO
    resources_pillaged
  FROM
    resources AS r
    LEFT JOIN rp on r.id = rp.resource
  WHERE
    r.movable='true';

  -- Generate resources that end up in the debris field
  -- in a similar way to the resources pillaged.
  WITH rp AS (
    SELECT
      t.resource,
      t.amount
    FROM
      json_to_recordset(debris) AS t(resource uuid, amount numeric(15, 5))
    )
  SELECT
    string_agg(concat_ws(' unit(s) of ', floor(rp.amount)::integer::text, r.name), ', ')
  INTO
    resources_dispersed
  FROM
    rp
    INNER JOIN resources AS r ON rp.resource = r.id;

  -- Create the message entry.
  RETURN create_message_for(player_id, 'fight_report_footer', resources_pillaged, resources_dispersed, rebuilt::text);
END
$$ LANGUAGE plpgsql;

-- General orchestration function allowing to perform
-- the genration of the fight reports for the fleet
-- in input. Reports for both participants will be
-- generated assuming that the fleet is not part of
-- an ACS operation
CREATE OR REPLACE FUNCTION fight_report(players json, fleets json, indigenous uuid, planet_id uuid, planet_kind text, moment timestamp with time zone, outcome text, fleet_remains json, ships_remains json, def_remains json, pillage json, debris json, rebuilt integer) RETURNS VOID AS $$
DECLARE
  player_data json;
  fleet_data json;

  player_id uuid;
  fleet_id uuid;

  report_uuid uuid;

  report_header uuid;
  report_status uuid;
  report_footer uuid;
  report_participant uuid;
  report_indigenous uuid;

  report_arg_count integer;
BEGIN
  -- Post each part of the report. We need to post
  -- a full report to each of the players that are
  -- defined in the `players` input argument and
  -- also to the indigenous guy (identifier on its
  -- own with the `indigenous` variable).
  
  -- Iterate over the list of players and post the
  -- fight report for each one of them.
  FOR player_data IN SELECT * FROM json_array_elements(players)
  LOOP
     player_id := (player_data->>'player')::uuid;

    -- Header, status and footer.
    SELECT * INTO report_header FROM fleet_fight_report_header(player_id, planet_id, planet_kind, moment);
    SELECT * INTO report_status FROM fleet_fight_report_status(player_id, outcome);
    SELECT * INTO report_footer FROM fleet_fight_report_footer(player_id, pillage, debris, rebuilt);

    -- Start to assemble the fight report for this
    -- player.
    report_uuid := uuid_generate_v4();

    INSERT INTO messages_players(id, player, message)
      SELECT
        report_uuid,
        player_id,
        mi.id
      FROM
        messages_ids AS mi
      WHERE
        mi.name = 'fight_report';

    INSERT INTO messages_arguments("message", "position", "argument") VALUES(report_uuid, 0, report_header);
    INSERT INTO messages_arguments("message", "position", "argument") VALUES(report_uuid, 1, report_status);
    INSERT INTO messages_arguments("message", "position", "argument") VALUES(report_uuid, 2, report_footer);

    -- Don't forget the indigenous guy.
    SELECT * INTO report_indigenous FROM fleet_fight_report_indigenous_participant(player_id, planet_id, planet_kind, ships_remains, def_remains);

    INSERT INTO messages_arguments("message", "position", "argument") VALUES(report_uuid, 3, report_indigenous);

    -- Post a participant for each fleet that is
    -- participating to the fight.
    report_arg_count := 4;

    FOR fleet_data IN SELECT * FROM json_array_elements(fleets)
    LOOP
      fleet_id := (fleet_data->>'fleet')::uuid;

      -- Generate report for the current player.
      SELECT * INTO report_participant FROM fleet_fight_report_outsider_participant(player_id, fleet_id, fleet_remains);

      INSERT INTO messages_arguments("message", "position", "argument") VALUES(report_uuid, report_arg_count, report_participant);

      report_arg_count := report_arg_count + 1;
    END LOOP;
  END LOOP;

  -- Generate the fight report for the indigenous
  -- guy as well.
  SELECT * INTO report_header FROM fleet_fight_report_header(indigenous, planet_id, planet_kind, moment);
  SELECT * INTO report_status FROM fleet_fight_report_status(indigenous, outcome);
  SELECT * INTO report_footer FROM fleet_fight_report_footer(indigenous, pillage, debris, rebuilt);

  -- Start to assemble the fight report for this
  -- player.
  report_uuid := uuid_generate_v4();

  INSERT INTO messages_players(id, player, message)
    SELECT
      report_uuid,
      indigenous,
      mi.id
    FROM
      messages_ids AS mi
    WHERE
      mi.name = 'fight_report';

  INSERT INTO messages_arguments("message", "position", "argument") VALUES(report_uuid, 0, report_header);
  INSERT INTO messages_arguments("message", "position", "argument") VALUES(report_uuid, 1, report_status);
  INSERT INTO messages_arguments("message", "position", "argument") VALUES(report_uuid, 2, report_footer);

  -- Register the main defender units in the
  -- report as a participant.
  SELECT * INTO report_indigenous FROM fleet_fight_report_indigenous_participant(indigenous, planet_id, planet_kind, ships_remains, def_remains);

  INSERT INTO messages_arguments("message", "position", "argument") VALUES(report_uuid, 3, report_indigenous);

  report_arg_count := 4;

  FOR fleet_data IN SELECT * FROM json_array_elements(fleets)
  LOOP
    fleet_id := (fleet_data->>'fleet')::uuid;

    SELECT * INTO report_participant FROM fleet_fight_report_outsider_participant(indigenous, fleet_id, fleet_remains);

    INSERT INTO messages_arguments("message", "position", "argument") VALUES(report_uuid, report_arg_count, report_participant);

    report_arg_count := report_arg_count + 1;
  END LOOP;
END
$$ LANGUAGE plpgsql;
