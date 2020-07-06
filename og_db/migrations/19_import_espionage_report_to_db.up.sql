
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

-- Script allowing to perform the registration of an
-- espionage report for the player owning the input
-- fleet with the level of information.
CREATE OR REPLACE FUNCTION espionage_report(fleet_id uuid, counter_espionage integer, info_level integer) RETURNS VOID AS $$
DECLARE
  target_kind text;
  player_id uuid;
  target_name text;
  target_coordinates text;

  source_kind text;
  spyer_id uuid;
  spyer_name text;
  source_name text;
  source_coordinates text;
BEGIN
  -- We need to generate the counter espionage report for
  -- the player that was targeted by the fleet.
  PERFORM generate_counter_espionage_report(fleet_ud, counter_espionage);

  -- TODO: Handle this.

  -- Register the espionage report.
  -- PERFORM create_message_for(spyer_id, 'espionage_report', info_level::text);
END
$$ LANGUAGE plpgsql;