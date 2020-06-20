
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
CREATE OR REPLACE FUNCTION fleet_fight_report_outsiders_participant(fleet_id uuid, player_id uuid, remains json) RETURNS VOID AS $$
DECLARE
  source_kind text;

  source_name text;
  source_coordinates text;
  source_player_name text;

  weapon_tech text;
  shielding_tech text;
  armour_tech text;

  units_count text;
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

  -- Fetch the level of the technologies for the player.
  SELECT
    concat_ws('%', 10 * level)
  INTO
    weapon_tech
  FROM
    players_technologies AS pt
    INNER JOIN technologies AS t ON pt.technology = t.id
  WHERE
    player = player_id
    AND t.name = 'weapons';

  SELECT
    concat_ws('%', 10 * level)
  INTO
    shielding_tech
  FROM
    players_technologies AS pt
    INNER JOIN technologies AS t ON pt.technology = t.id
  WHERE
    player = player_id
    AND t.name = 'shielding';

  SELECT
    concat_ws('%', 10 * level)
  INTO
    armour_tech
  FROM
    players_technologies AS pt
    INNER JOIN technologies AS t ON pt.technology = t.id
  WHERE
    player = player_id
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
  PERFORM create_message_for(player_id, 'fight_report_participant', source_player_name, source_name, source_coordinates, units_count, units_lost::text, weapons_tech, shielding_tech, armour_tech);
END
$$ LANGUAGE plpgsql;

-- Generate the list of forces in a fleet fight in
-- the case of the defender. Note that this should
-- only be used for the owner of the planet where
-- the fleet is fighting as it includes the sum
-- of the defense systems existing on the planet.
CREATE OR REPLACE FUNCTION fleet_fight_report_indigenous_participant(fleet_id uuid, def_remains json, ships_remains json) RETURNS VOID AS $$
DECLARE
BEGIN
  -- TODO: Handle this.
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

