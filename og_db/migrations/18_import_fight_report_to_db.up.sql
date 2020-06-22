
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
      f.id = fleet_id;
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
      f.id = fleet_id;
  END IF;

  -- Create the message for the specified player.
  PERFORM create_message_for(player_id, 'fight_report_header', target_name, target_coordinates, moment);
END
$$ LANGUAGE plpgsql;

-- Generate the list of forces in a fleet fight and
-- post the created report to the specified player.
CREATE OR REPLACE FUNCTION fleet_fight_report_outsiders_participant(fleet_id uuid, player_id uuid, remains json) RETURNS VOID AS $$
DECLARE
  source_kind text;

  source_name text;
  source_coordinates text;
  source_player_id uuid;
  source_player_name text;

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

  IF target_kind = 'planet' THEN
    SELECT
      p.name,
      concat_ws(':', p.galaxy,  p.solar_system,  p.position),
      f.player,
      pl.name
    INTO
      source_name,
      source_coordinates,
      source_player_id,
      source_player_name
    FROM
      fleets AS f
      INNER JOIN planets AS p ON f.source = p.id
      INNER JOIN players AS pl ON p.player = pl.id
    WHERE
      f.id = fleet_id;
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
      t.ship,
      t.count
    FROM
      json_to_recordset(remains) AS t(ship uuid, count integer)
    )
  SELECT
    COALESCE(SUM(sc.cost), 0)
  INTO
    units_lost
  FROM
    fleets_ships AS fs
    LEFT JOIN rs ON fs.ship = rs.ship
    INNER JOIN ships_costs AS sc ON fs.ship = sc.element
    INNER JOIN resources AS r ON sc.res = r.id
  WHERE
    fs.fleet = fleet_id
    AND r.dispersable = 'true';

  -- Create the message for the specified player.
  PERFORM create_message_for(player_id, 'fight_report_participant', source_player_name, source_name, source_coordinates, units_count::text, units_lost::text, weapons_tech, shielding_tech, armour_tech);
END
$$ LANGUAGE plpgsql;

-- Generate the list of forces in a fleet fight in
-- the case of the defender. Note that this should
-- only be used for the owner of the planet where
-- the fleet is fighting as it includes the sum
-- of the defense systems existing on the planet.
CREATE OR REPLACE FUNCTION fleet_fight_report_indigenous_participant(planet_id uuid, kind text, player_id uuid, ships_remains json, def_remains json) RETURNS VOID AS $$
DECLARE
  target_player_id uuid;
  target_name text;
  target_coordinates text;
  target_player_name text;

  weapon_tech text;
  shielding_tech text;
  armour_tech text;

  ships_count text;
  defenses_count text;
  ships_lost integer;
  defenses_lost integer;
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
      COALESCE(SUM(sc.cost), 0)
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

    WITH rs AS (
      SELECT
        t.defense,
        t.count
      FROM
        json_to_recordset(def_remains) AS t(defense uuid, count integer)
      )
    SELECT
      COALESCE(SUM(dc.cost), 0)
    INTO
      defenses_lost
    FROM
      planets_defenses AS pd
      LEFT JOIN rs ON pd.defense = rs.defense
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
      COALESCE(SUM(sc.cost), 0)
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

    WITH rs AS (
      SELECT
        t.defense,
        t.count
      FROM
        json_to_recordset(def_remains) AS t(defense uuid, count integer)
      )
    SELECT
      COALESCE(SUM(dc.cost), 0)
    INTO
      defenses_lost
    FROM
      moons_defenses AS md
      LEFT JOIN rs ON md.defense = rs.defense
      INNER JOIN defenses_costs AS dc ON md.defense = dc.element
      INNER JOIN resources AS r ON dc.res = r.id
    WHERE
      md.moon = planet_id
      AND r.dispersable = 'true';
  END IF;

  -- Create the message for the specified player.
  PERFORM create_message_for(player_id, 'fight_report_participant', target_player_name, target_name, target_coordinates, (ships_count + defenses_count)::text, (ships_lost + defenses_lost)::text, weapons_tech, shielding_tech, armour_tech);
END
$$ LANGUAGE plpgsql;

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

-- Generate the report indicating the final result
-- of the fight including the generated debris field
-- and the plundered resources if any.
CREATE OR REPLACE FUNCTION fight_report_footer(player_id uuid, pillage json, debris json, rebuilt json) RETURNS VOID AS $$
DECLARE
  resources_pillaged text;
  resources_dispersed text;
  units_rebuilt integer;
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
    string_agg(concat_ws(' unit(s) of ', floor(rp.amount)::integer::text, r.name), ', ')
  INTO
    resources_pillaged
  FROM
    rp
    INNER JOIN resources AS r ON rp.resource = r.id;

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

  -- Generate the count of rebuilt units. We need to
  -- compute the difference between the existing units
  -- and the ones that remain.
  WITH ru AS (
    SELECT
      t.defense,
      t.count
    FROM
      json_to_recordset(rebuilt) AS t(defense uuid, count integer)
    )
  SELECT
    COALESCE(SUM(ru.count), 0)
  INTO
    units_rebuilt
  FROM
    ru;

  -- Create the message entry.
  PERFORM create_message_for(player_id, 'fight_report_footer', resources_pillaged, resources_dispersed, units_rebuilt::text);
END
$$ LANGUAGE plpgsql;

-- General orchestration function allowing to perform
-- the genration of the fight reports for the fleet
-- in input. Reports for both participants will be
-- generated assuming that the fleet is not part of
-- an ACS operation
CREATE OR REPLACE FUNCTION fight_report(fleet_id uuid, outcome text, fleet_remains json, ships_remains json, def_remains json, pillage json, debris json, rebuilt json) RETURNS VOID AS $$
DECLARE
  source_player uuid;

  target_kind text;
  target_id uuid;
  target_player uuid;
BEGIN
  -- Fetch the source and target players.
  SELECT player INTO source_player FROM fleets WHERE id = fleet_id;
  IF NOT FOUND THEN
    RAISE EXCEPTION 'Invalid source player for fleet % in fight report operation', fleet_id;
  END IF;

  SELECT target_type INTO target_kind FROM fleets WHERE id = fleet_id;
  IF NOT FOUND THEN
    RAISE EXCEPTION 'Invalid target player for fleet % in fight report operation', fleet_id;
  END IF;

  IF target_kind = 'planet' THEN
    SELECT
      p.player,
      f.target
    INTO
      target_player,
      target_id
    FROM
      fleets AS f
      INNER JOIN planets AS p ON f.target = p.id
    WHERE
      f.id = fleet_id;
  END IF;

  IF target_kind = 'moon' THEN
    SELECT
      p.player,
      f.target
    INTO
      target_player,
      target_id
    FROM
      fleets AS f
      INNER JOIN moons AS m ON f.target = m.id
      INNER JOIN planets AS p ON m.planet = p.id
    WHERE
      f.id = fleet_id;
  END IF;

  -- Post headers for each player.
  PERFORM fleet_fight_report_header(fleet_id, source_player);
  PERFORM fleet_fight_report_header(fleet_id, target_player);

  -- Post participants for each player.
  PERFORM fleet_fight_report_outsiders_participant(fleet_id, source_player, fleet_remains);
  PERFORM fleet_fight_report_outsiders_participant(fleet_id, target_player, fleet_remains);

  PERFORM fleet_fight_report_indigenous_participant(target_id, target_kind, source_player, ships_remains, def_remains);
  PERFORM fleet_fight_report_indigenous_participant(target_id, target_kind, target_player, ships_remains, def_remains);

  -- Post status for each player.
  PERFORM fleet_fight_report_status(outcome, source_player);
  PERFORM fleet_fight_report_status(outcome, target_player);

  -- Post footer for each player.
  PERFORM fight_report_footer(source_player, pillage, debris, rebuilt);
  PERFORM fight_report_footer(target_player, pillage, debris, rebuilt);

  -- TODO: Add other fleets (and thus reports for other players).
END
$$ LANGUAGE plpgsql;
