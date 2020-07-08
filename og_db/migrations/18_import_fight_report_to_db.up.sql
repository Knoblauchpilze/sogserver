
-- Generate the fight report header. The report will
-- be posted for the player specified in input.
CREATE OR REPLACE FUNCTION fleet_fight_report_header(player_id uuid, planet_id uuid, planet_kind text, moment timestamp with time zone, report_id uuid) RETURNS VOID AS $$
DECLARE
  planet_name text;
  planet_coordinates text;
  moment_text text;

  header_id uuid;
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
  SELECT * INTO header_id FROM create_message_for(player_id, 'fight_report_header', planet_name, planet_coordinates, moment_text);

  -- Register the header as an argument of the
  -- parent report. Note that the header is
  -- always the first argument of the fight
  -- report.
  INSERT INTO messages_arguments("message", "position", "argument") VALUES(report_id, 0, header_id);
END
$$ LANGUAGE plpgsql;

-- Generate the list of forces in a fleet fight and
-- post the created report to the specified player.
CREATE OR REPLACE FUNCTION fleet_fight_report_outsider_participant(player_id uuid, fleet_id uuid, remains json, report_id uuid, pOffset integer) RETURNS integer AS $$
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

  msg_id uuid;
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
  SELECT * INTO msg_id FROM create_message_for(player_id, 'fight_report_participant', hostile_status::text, source_player_name, source_name, source_coordinates, units_count::text, units_lost::text, weapon_tech, shielding_tech, armour_tech);

  -- Register this participant as a child of the
  -- parent fight report.
  INSERT INTO messages_arguments("message", "position", "argument") VALUES(report_id, pOffset, msg_id);

  -- Return the index of the next argument of the
  -- fight report.
  RETURN pOffset + 1;
END
$$ LANGUAGE plpgsql;

-- Generate the list of forces in a fleet fight in
-- the case of the defender. Note that this should
-- only be used for the owner of the planet where
-- the fleet is fighting as it includes the sum
-- of the defense systems existing on the planet.
CREATE OR REPLACE FUNCTION fleet_fight_report_indigenous_participant(player_id uuid, planet_id uuid, kind text, ships_remains json, def_remains json, report_id uuid, pOffset integer) RETURNS integer AS $$
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

  msg_id uuid;
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
  SELECT * INTO msg_id FROM create_message_for(player_id, 'fight_report_participant', 'false', target_player_name, target_name, target_coordinates, (ships_count + defenses_count)::text, (ships_lost + defenses_lost)::text, weapon_tech, shielding_tech, armour_tech);

  -- Register this participant as a child of the
  -- parent fight report.
  INSERT INTO messages_arguments("message", "position", "argument") VALUES(report_id, pOffset, msg_id);

  -- Return the index of the next argument of the
  -- fight report.
  RETURN pOffset + 1;
END
$$ LANGUAGE plpgsql;

-- Generate the status of the fight. Depending on the
-- provided outcome the correct message will be posted
-- to the specified player.
CREATE OR REPLACE FUNCTION fleet_fight_report_status(player_id uuid, outcome text, report_id uuid, pOffset integer) RETURNS integer AS $$
DECLARE
  status_id uuid;
BEGIN
  -- The perspective of the fight is seen from the
  -- defender point of view. Note that as this type
  -- of message do not take any argument we provide
  -- a dummy array so that the variadic condition
  -- is satisfied.
  IF outcome = 'victory' THEN
    SELECT * INTO status_id FROM create_message_for(player_id, 'fight_report_result_defender_win', VARIADIC '{}'::text[]);
  END IF;

  IF outcome = 'draw' THEN
    SELECT * INTO status_id FROM create_message_for(player_id, 'fight_report_result_draw', VARIADIC '{}'::text[]);
  END IF;

  IF outcome = 'loss' THEN
    SELECT * INTO status_id FROM create_message_for(player_id, 'fight_report_result_attacker_win', VARIADIC '{}'::text[]);
  END IF;

  -- Register the status as an argument of the
  -- parent report.
  INSERT INTO messages_arguments("message", "position", "argument") VALUES(report_id, pOffset, status_id);

  -- Return the position of the next argument
  -- for the parent report.
  RETURN pOffset + 1;
END
$$ LANGUAGE plpgsql;

-- Generate the report indicating the final result
-- of the fight including the generated debris field
-- and the plundered resources if any.
CREATE OR REPLACE FUNCTION fleet_fight_report_footer(player_id uuid, pillage json, debris json, rebuilt integer, report_id uuid, pOffset integer) RETURNS VOID AS $$
DECLARE
  resources_pillaged text;
  resources_dispersed text;

  footer_id uuid;
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
  SELECT * INTO footer_id FROM create_message_for(player_id, 'fight_report_footer', resources_pillaged, resources_dispersed, rebuilt::text);

  -- Register the footer as an argument of the
  -- parent report.
  INSERT INTO messages_arguments("message", "position", "argument") VALUES(report_id, pOffset, footer_id);
END
$$ LANGUAGE plpgsql;

-- Creation of a fight report for the specified player
-- player in input. It handles the generation of all
-- parts of the report and post each one of them as a
-- message to the `player_id` in input.
-- Allows to mutualize a bit the code used in the
-- `fight_report` function where a lot of reports
-- need to be created for all the involved players.
CREATE OR REPLACE FUNCTION fight_report_for_player(player_id uuid, attacking_fleets json, defending_fleets json, indigenous uuid, planet_id uuid, planet_kind text, moment timestamp with time zone, outcome text, fleet_remains json, ships_remains json, def_remains json, pillage json, debris json, rebuilt integer) RETURNS VOID AS $$
DECLARE
  report_id uuid := uuid_generate_v4();
  pOffset integer := 0;

  fleet_data json;
  fleet_id uuid;
BEGIN
  -- Start to assemble the fight report for this
  -- player.
  INSERT INTO messages_players(id, player, message)
    SELECT
      report_id,
      player_id,
      mi.id
    FROM
      messages_ids AS mi
    WHERE
      mi.name = 'fight_report';

  -- Generate the header.
  PERFORM fleet_fight_report_header(player_id, planet_id, planet_kind, moment, report_id);

  -- Generate attackers.
  pOffset := 1;

  FOR fleet_data IN SELECT * FROM json_array_elements(attacking_fleets)
  LOOP
    fleet_id := (fleet_data->>'fleet')::uuid;
    SELECT * INTO pOffset FROM fleet_fight_report_outsider_participant(player_id, fleet_id, fleet_remains, report_id, pOffset);
  END LOOP;

  -- Generate defenders.
  FOR fleet_data IN SELECT * FROM json_array_elements(defending_fleets)
  LOOP
    fleet_id := (fleet_data->>'fleet')::uuid;
    SELECT * INTO pOffset FROM fleet_fight_report_outsider_participant(player_id, fleet_id, fleet_remains, report_id, pOffset);
  END LOOP;

  SELECT * INTO pOffset FROM fleet_fight_report_indigenous_participant(player_id, planet_id, planet_kind, ships_remains, def_remains, report_id, pOffset);

  -- Generate status.
  SELECT * INTO pOffset FROM fleet_fight_report_status(player_id, outcome, report_id, pOffset);

  -- Generate footer.
  PERFORM fleet_fight_report_footer(player_id, pillage, debris, rebuilt, report_id, pOffset);
END
$$ LANGUAGE plpgsql;

-- General orchestration function allowing to perform
-- the genration of the fight reports for the fleet
-- in input. Reports for both participants will be
-- generated assuming that the fleet is not part of
-- an ACS operation
CREATE OR REPLACE FUNCTION fight_report(players json, attacking_fleets json, defending_fleets json, indigenous uuid, planet_id uuid, planet_kind text, moment timestamp with time zone, outcome text, fleet_remains json, ships_remains json, def_remains json, pillage json, debris json, rebuilt integer) RETURNS VOID AS $$
DECLARE
  player_data json;
  player_id uuid;
BEGIN
  -- Generate a report for each of the players
  -- involved in the fight.
  FOR player_data IN SELECT * FROM json_array_elements(players)
  LOOP
     player_id := (player_data->>'player')::uuid;
     PERFORM fight_report_for_player(player_id, attacking_fleets, defending_fleets, indigenous, planet_id, planet_kind, moment, outcome, fleet_remains, ships_remains, def_remains, pillage, debris, rebuilt);
  END LOOP;

  -- The indigenous guy does not exist in the
  -- input `players` array so we need to
  -- handle this case afterwards.
  PERFORM fight_report_for_player(indigenous, attacking_fleets, defending_fleets, indigenous, planet_id, planet_kind, moment, outcome, fleet_remains, ships_remains, def_remains, pillage, debris, rebuilt);
END
$$ LANGUAGE plpgsql;
