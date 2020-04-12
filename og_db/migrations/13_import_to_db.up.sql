
-- Import the universes into the corresponding table.
CREATE OR REPLACE FUNCTION create_universe(inputs json) RETURNS VOID AS $$
BEGIN
  INSERT INTO universes SELECT * FROM json_populate_record(null::universes, inputs);
END
$$ LANGUAGE plpgsql;

-- Import the accounts into the corresponding table.
CREATE OR REPLACE FUNCTION create_account(inputs json) RETURNS VOID AS $$
BEGIN
  INSERT INTO accounts SELECT * FROM json_populate_record(null::accounts, inputs);
END
$$ LANGUAGE plpgsql;

-- Create players from the account and universe data.
CREATE OR REPLACE FUNCTION create_player(inputs json) RETURNS VOID AS $$
BEGIN
  -- Insert the player's data into the dedicated table.
  INSERT INTO players SELECT * FROM json_populate_record(null::players, inputs);

  -- Insert technologies with a `0` level in the table.
  -- The conversion in itself includes retrieving the `json`
  -- key by value and then converting it to a uuid. Here is
  -- a useful link:
  -- https://stackoverflow.com/questions/53567903/postgres-cast-to-uuid-from-json
  INSERT INTO player_technologies(player, technology, level)
    SELECT (inputs->>'id')::uuid, t.id, 0
    FROM technologies t;
END
$$ LANGUAGE plpgsql;

-- Import planet into the corresponding table.
CREATE OR REPLACE FUNCTION create_planet(planet json, resources json) RETURNS VOID AS $$
BEGIN
  -- Insert the planet in the planets table.
  INSERT INTO planets SELECT * FROM json_populate_record(null::planets, planet);

  -- Insert the base resources of the planet.
  INSERT INTO planets_resources select * FROM json_populate_recordset(null::planets_resources, resources);

  -- Insert base buildings, ships, defenses on the planet.
  INSERT INTO planets_buildings(planet, building, level)
    SELECT (planet->>'id')::uuid, b.id, 0
    FROM buildings b;

  INSERT INTO planets_ships(planet, ship, count)
    SELECT (planet->>'id')::uuid, s.id, 0
    FROM ships s;

  INSERT INTO planets_defenses(planet, defense, count)
    SELECT (planet->>'id')::uuid, d.id, 0
    FROM defenses d;
END
$$ LANGUAGE plpgsql;

-- Import building upgrade action in the dedicated table.
CREATE OR REPLACE FUNCTION create_building_upgrade_action(upgrade json) RETURNS VOID AS $$
BEGIN
  INSERT INTO construction_actions_buildings SELECT * FROM json_populate_record(null::construction_actions_buildings, upgrade);
END
$$ LANGUAGE plpgsql;

-- Import technology upgrade action in the dedicated table.
CREATE OR REPLACE FUNCTION create_technology_upgrade_action(upgrade json) RETURNS VOID AS $$
BEGIN
  INSERT INTO construction_actions_technologies SELECT * FROM json_populate_record(null::construction_actions_technologies, upgrade);
END
$$ LANGUAGE plpgsql;

-- Import ship upgrade action in the dedicated table.
CREATE OR REPLACE FUNCTION create_ship_upgrade_action(upgrade json) RETURNS VOID AS $$
BEGIN
  INSERT INTO construction_actions_ships SELECT * FROM json_populate_record(null::construction_actions_ships, upgrade);
END
$$ LANGUAGE plpgsql;

-- Import defense upgrade action in the dedicated table.
CREATE OR REPLACE FUNCTION create_defense_upgrade_action(upgrade json) RETURNS VOID AS $$
BEGIN
  INSERT INTO construction_actions_defenses SELECT * FROM json_populate_record(null::construction_actions_defenses, upgrade);
END
$$ LANGUAGE plpgsql;

-- Import fleets in the relevant table.
CREATE OR REPLACE FUNCTION create_fleet(inputs json) RETURNS VOID AS $$
BEGIN
  INSERT INTO fleets SELECT * FROM json_populate_record(null::fleets, inputs);
END
$$ LANGUAGE plpgsql;

-- Import fleet components in the relevant table.
CREATE OR REPLACE FUNCTION create_fleet_component(component json, ships json) RETURNS VOID AS $$
BEGIN
  -- Insert the fleet element.
  INSERT INTO fleet_elements SELECT * FROM json_populate_record(null::fleet_elements, component);

  -- Insert the ships for this fleet element.
  INSERT INTO fleet_ships SELECT * FROM json_populate_recordset(null::fleet_ships, ships);
END
$$ LANGUAGE plpgsql;

-- Update resources for a planet.
CREATE OR REPLACE FUNCTION update_resources_for_planet(resources json) RETURNS VOID AS $$
BEGIN
  WITH updateData
    AS (SELECT * FROM jsonb_populate_recordset(resources))
  UPDATE planets_resources pr
    SET amount = ud.amount
  FROM updateData AS ud
  WHERE pr.planet = ud.planet AND pr.res = ud.res;
END
$$ LANGUAGE plpgsql;

-- Update upgrade action for buildings.
CREATE OR REPLACE FUNCTION update_building_upgrade_action(planet_id uuid) RETURNS VOID AS $$
BEGIN
  WITH updateData
    AS (SELECT * FROM construction_actions_buildings cab WHERE cab.planet=planet_id)
  UPDATE planets_buildings pb
    SET level = ud.desired_level
  FROM updateData AS ud
  WHERE
    pb.planet = planet_id
    AND pb.building = ud.building
    AND pb.level = ud.current_level
    AND ud.completion_time < NOW();

  DELETE FROM construction_actions_buildings WHERE planet = planet_id AND completion_time < NOW();
END
$$ LANGUAGE plpgsql;

-- Update upgrade action for technologies.
CREATE OR REPLACE FUNCTION update_technology_upgrade_action(player_id uuid) RETURNS VOID AS $$
BEGIN
  WITH updateData
    AS (SELECT * FROM construction_actions_technologies cat WHERE cat.player=player_id)
  UPDATE player_technologies pt
    SET level = ud.desired_level
  FROM updateData AS ud
  WHERE
    pt.player = player_id
    AND pt.technology = ud.technology
    AND pt.level = ud.current_level
    AND ud.completion_time < NOW();

  DELETE FROM construction_actions_technologies WHERE player = player_id AND completion_time < NOW();
END
$$ LANGUAGE plpgsql;

-- Update upgrade action for ships.
CREATE OR REPLACE FUNCTION update_ship_upgrade_action(planet_id uuid) RETURNS VOID AS $$
DECLARE
  -- Save time: this will make sure that we can't run into
  -- problem where for example an action is not complete
  -- when the 1. is performed and complete when the 2. is
  -- performed (resulting in a ship never being built).
  processing_time TIMESTAMP := NOW();
BEGIN
  -- 1. Register ships that are now complete. We need to account for
  -- the fact that several elements might have completed while we
  -- were not updating this action.
  -- The algorithm is basically:
  --   - compute the number of intervals that have elapsed.
  --   - subtract the number of already built elements.
  --   - clamp tp make sure that we don't create too many elements.
  WITH updateData
    AS (SELECT * FROM construction_actions_ships WHERE planet = planet_id)
  UPDATE planets_ships ps
    SET count = count - (ud.amount - ud.remaining) +
      LEAST(
        EXTRACT(MILLISECONDS FROM processing_time - ud.created_at) / EXTRACT(MILLISECOND FROM ud.completion_time),
        CAST(ud.amount AS DOUBLE PRECISION)
      )
  FROM updateData AS ud
  WHERE
    ps.planet = planet_id
    AND ps.ship = ud.ship
    AND ud.created_at + (ud.amount - (ud.remaining - 1)) * ud.completion_time < processing_time;

  -- 2. Update remaining actions with an amount decreased by an amount
  -- consistent with the duration elapsed since the creation.
  UPDATE construction_actions_ships
    SET remaining = amount -
      LEAST(
        EXTRACT(MILLISECONDS FROM processing_time - created_at) / EXTRACT(MILLISECOND FROM completion_time),
        CAST(amount AS DOUBLE PRECISION)
      )
  WHERE
    planet = planet_id
    AND created_at + (amount - (remaining - 1)) * completion_time < processing_time;

  -- 3. Delete actions that don't have any remaining effect.
  DELETE FROM construction_actions_ships WHERE planet = planet_id AND remaining = 0;
END
$$ LANGUAGE plpgsql;

-- Update upgrade action for defenses.
CREATE OR REPLACE FUNCTION update_defense_upgrade_action(planet_id uuid) RETURNS VOID AS $$
DECLARE
  -- Similar mechanism to the one used for ships.
  processing_time TIMESTAMP := NOW();
BEGIN
  -- 1. Register defenses that are now complete. We need to account for
  -- the fact that several elements might have completed while we were
  -- not updating this action.
  -- The algorithm is basically:
  --   - compute the number of intervals that have elapsed.
  --   - subtract the number of already built elements.
  --   - clamp tp make sure that we don't create too many elements.
  WITH updateData
    AS (SELECT * FROM construction_actions_defenses WHERE planet = planet_id)
  UPDATE planets_defenses pd
    SET count = count - (ud.amount - ud.remaining) +
      LEAST(
        EXTRACT(MILLISECONDS FROM processing_time - ud.created_at) / EXTRACT(MILLISECOND FROM ud.completion_time),
        CAST(ud.amount AS DOUBLE PRECISION)
      )
  FROM updateData AS ud
  WHERE
    pd.planet = planet_id
    AND pd.defense = ud.defense
    AND ud.created_at + (ud.amount - (ud.remaining - 1)) * ud.completion_time < processing_time;

  -- 2. Update remaining actions with an amount decreased by an amount
  -- consistent with the duration elapsed since the creation.
  UPDATE construction_actions_defenses
    SET remaining = amount -
      LEAST(
        EXTRACT(MILLISECONDS FROM processing_time - created_at) / EXTRACT(MILLISECOND FROM completion_time),
        CAST(amount AS DOUBLE PRECISION)
      )
  WHERE
    planet = planet_id
    AND created_at + (amount - (remaining - 1)) * completion_time < processing_time;

  -- 3.Delete actions that don't have any remaining effect.
  DELETE FROM construction_actions_defenses WHERE planet = planet_id AND remaining = 0;
END
$$ LANGUAGE plpgsql;
