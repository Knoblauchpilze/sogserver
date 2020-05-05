
-- Import the universes into the corresponding table.
CREATE OR REPLACE FUNCTION create_universe(inputs json) RETURNS VOID AS $$
BEGIN
  INSERT INTO universes
    SELECT *
    FROM json_populate_record(null::universes, inputs);
END
$$ LANGUAGE plpgsql;

-- Import the accounts into the corresponding table.
CREATE OR REPLACE FUNCTION create_account(inputs json) RETURNS VOID AS $$
BEGIN
  INSERT INTO accounts
    SELECT *
    FROM json_populate_record(null::accounts, inputs);
END
$$ LANGUAGE plpgsql;

-- Create players from the account and universe data.
CREATE OR REPLACE FUNCTION create_player(inputs json) RETURNS VOID AS $$
BEGIN
  -- Insert the player's data into the dedicated table.
  INSERT INTO players
    SELECT *
    FROM json_populate_record(null::players, inputs);

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
CREATE OR REPLACE FUNCTION create_planet(planet_data json, resources json) RETURNS VOID AS $$
BEGIN
  -- Insert the planet in the planets table.
  INSERT INTO planets
    SELECT *
    FROM json_populate_record(null::planets, planet_data);

  -- Insert the base resources of the planet.
  INSERT INTO planets_resources
    SELECT
      (planet_data->>'id')::uuid,
      res,
      amount,
      production,
      storage_capacity
    FROM
      json_populate_recordset(null::planets_resources, resources);

  -- Insert base buildings, ships, defenses on the planet.
  INSERT INTO planets_buildings(planet, building, level)
    SELECT (planet_data->>'id')::uuid, b.id, 0
    FROM buildings b;

  INSERT INTO planets_ships(planet, ship, count)
    SELECT (planet_data->>'id')::uuid, s.id, 0
    FROM ships s;

  INSERT INTO planets_defenses(planet, defense, count)
    SELECT (planet_data->>'id')::uuid, d.id, 0
    FROM defenses d;
END
$$ LANGUAGE plpgsql;

-- Import building upgrade action in the dedicated table.
CREATE OR REPLACE FUNCTION create_building_upgrade_action(upgrade json, costs json, production_effects json, storage_effects json) RETURNS VOID AS $$
BEGIN
  -- Create the building upgrade action itself.
  INSERT INTO construction_actions_buildings
    SELECT *
    FROM json_populate_record(null::construction_actions_buildings, upgrade);

  -- Update the construction action effects (both in terms of
  -- storage and production).
  INSERT INTO construction_actions_buildings_production_effects
    SELECT
      (upgrade->>'id')::uuid,
      resource,
      production_change
    FROM
      json_populate_recordset(null::construction_actions_buildings_production_effects, production_effects);

  INSERT INTO construction_actions_buildings_storage_effects
    SELECT
      (upgrade->>'id')::uuid,
      resource,
      storage_capacity_change
    FROM
      json_populate_recordset(null::construction_actions_buildings_storage_effects, storage_effects);

  -- Subtract the cost of the action to the resources existing
  -- the planet so that it's no longer available. We assume a
  -- valid amount of resources remaining (no checks to clamp
  -- to `0` or anything). We will update the existing resource
  -- on the planet before decreasing in order to make sure we
  -- have a valid amount.
  PERFORM update_resources_for_planet((upgrade->>'planet')::uuid);

  WITH rc AS (
    SELECT
      t.resource,
      t.cost
    FROM
      json_to_recordset(costs) AS t(resource uuid, cost numeric(15, 5))
    )
  UPDATE
    planets_resources
  SET
    amount = amount - rc.cost
  FROM
    rc
  WHERE planet = (upgrade->>'planet')::uuid
  AND res = rc.resource;

END
$$ LANGUAGE plpgsql;

-- Import technology upgrade action in the dedicated table.
CREATE OR REPLACE FUNCTION create_technology_upgrade_action(upgrade json, costs json) RETURNS VOID AS $$
BEGIN
  -- Insert the construction action in the related table.
  INSERT INTO construction_actions_technologies
    SELECT *
    FROM json_populate_record(null::construction_actions_technologies, upgrade);

  -- Perform the update of the resources by subtracting
  -- the cost of the action.
  PERFORM update_resources_for_planet((upgrade->>'planet')::uuid);

  WITH rc AS (
    SELECT
      t.resource,
      t.cost
    FROM
      json_to_recordset(costs) AS t(resource uuid, cost numeric(15, 5))
    )
  UPDATE
    planets_resources
  SET
    amount = amount - rc.cost
  FROM
    rc
  WHERE planet = (upgrade->>'planet')::uuid
  AND res = rc.resource;
END
$$ LANGUAGE plpgsql;

-- Import ship upgrade action in the dedicated table.
CREATE OR REPLACE FUNCTION create_ship_upgrade_action(upgrade json, costs json) RETURNS VOID AS $$
BEGIN
  -- Insert the construction action in the related table.
  INSERT INTO construction_actions_ships
    SELECT *
    FROM json_populate_record(null::construction_actions_ships, upgrade);

  -- Perform the update of the resources by subtracting
  -- the cost of the action.
  PERFORM update_resources_for_planet((upgrade->>'planet')::uuid);

  WITH rc AS (
    SELECT
      t.resource,
      t.cost
    FROM
      json_to_recordset(costs) AS t(resource uuid, cost numeric(15, 5))
    )
  UPDATE
    planets_resources
  SET
    amount = amount - rc.cost
  FROM
    rc
  WHERE
    planet = (upgrade->>'planet')::uuid
    AND res = rc.resource;
END
$$ LANGUAGE plpgsql;

-- Import defense upgrade action in the dedicated table.
CREATE OR REPLACE FUNCTION create_defense_upgrade_action(upgrade json, costs json) RETURNS VOID AS $$
BEGIN
  -- Insert the construction action in the related table.
  INSERT INTO construction_actions_defenses
    SELECT *
    FROM json_populate_record(null::construction_actions_defenses, upgrade);

  -- Perform the update of the resources by subtracting
  -- the cost of the action.
  PERFORM update_resources_for_planet((upgrade->>'planet')::uuid);

  WITH rc AS (
    SELECT
      t.resource,
      t.cost
    FROM
      json_to_recordset(costs) AS t(resource uuid, cost numeric(15, 5))
    )
  UPDATE
    planets_resources
  SET
    amount = amount - rc.cost
  FROM
    rc
  WHERE planet = (upgrade->>'planet')::uuid
  AND res = rc.resource;
END
$$ LANGUAGE plpgsql;

-- Import fleets in the relevant table.
CREATE OR REPLACE FUNCTION create_fleet(inputs json) RETURNS VOID AS $$
BEGIN
  INSERT INTO fleets
    SELECT *
    FROM json_populate_record(null::fleets, inputs);
END
$$ LANGUAGE plpgsql;

-- Import fleet components in the relevant table.
CREATE OR REPLACE FUNCTION create_fleet_component(component json, ships json, resources json, consumption json) RETURNS VOID AS $$
BEGIN
  -- Insert the fleet element.
  INSERT INTO fleet_elements
    SELECT *
    FROM json_populate_record(null::fleet_elements, component);

  -- Insert the ships for this fleet element.
  INSERT INTO fleet_ships
    SELECT *
    FROM json_populate_recordset(null::fleet_ships, ships);

  -- Insert the resources for this fleet element.
  INSERT INTO fleet_resources
    SELECT *
    FROM json_populate_recordset(null::fleet_resources, resources);

  -- Reduce the planet's resources from the amount of the fuel.
  WITH cc AS (
    SELECT
      t.resource,
      t.amount AS quantity
    FROM
      json_to_recordset(consumption) AS t(resource uuid, amount numeric(15, 5))
    )
  UPDATE planets_resources
    SET amount = amount - cc.quantity
  FROM
    cc
  WHERE
    planet = (component->>'planet')::uuid
    AND res = cc.resource;

  -- Reduce the planet's resources from the amount that will be moved.
  WITH cr AS (
    SELECT
      t.resource,
      t.amount AS quantity
    FROM
      json_to_recordset(resources) AS t(resource uuid, amount numeric(15, 5))
    )
  UPDATE planets_resources
    SET amount = amount - cr.quantity
  FROM
    cr
  WHERE
    planet = (component->>'planet')::uuid
    AND res = cr.resource;

  -- Reduce the planet's available ships from the ones that will be launched.
  WITH cs AS (
    SELECT
      t.ship AS vessel,
      t.count AS quantity
    FROM
      json_to_recordset(ships) AS t(ship uuid, count integer)
    )
  UPDATE planets_ships
    SET count = count - cs.quantity
  FROM
    cs
  WHERE
    planet = (component->>'planet')::uuid
    AND ship = cs.vessel;
END
$$ LANGUAGE plpgsql;

-- Update resources for a planet.
CREATE OR REPLACE FUNCTION update_resources_for_planet(planet_id uuid) RETURNS VOID AS $$
DECLARE
  -- Save time: this will make sure thatall resources are
  -- updated at the same time.
  processing_time TIMESTAMP := NOW();
BEGIN
  -- Update the amount of resource to be at most the storage
  -- capacity, and otherwise to increase by the duration that
  -- passed between the last update and the current time.
  -- Note that even if the production is expressed in hours,
  -- we need to extract the number of seconds in order to be
  -- able to obtain fractions of an hour to update the value.
  UPDATE planets_resources
    SET amount = amount + EXTRACT(EPOCH FROM processing_time - updated_at) * production / 3600.0
  FROM
    resources AS r
  WHERE
    planet = planet_id
    AND res = r.id
    AND r.storable='true';
END
$$ LANGUAGE plpgsql;

-- Update upgrade action for buildings.
CREATE OR REPLACE FUNCTION update_building_upgrade_action(planet_id uuid) RETURNS VOID AS $$
DECLARE
  -- Save time: this will make sure that we can't run into
  -- problem where for example an action is not complete
  -- when the 1. is performed and complete when the 2. is
  -- performed (resulting in a ship never being built).
  processing_time TIMESTAMP := NOW();
BEGIN
  -- 1. Update the actions by updating the level of each
  -- building having a completed action. Make sure to
  -- order by ascending order of operations in case some
  -- pending operations concerns the same building for
  -- different levels.
  WITH update_data
    AS (
      SELECT *
      FROM construction_actions_buildings
      WHERE
        planet = planet_id AND
        completion_time < processing_time
      ORDER BY
        desired_level ASC
    )
  UPDATE planets_buildings pb
    SET level = ud.desired_level
  FROM update_data AS ud
  WHERE
    pb.planet = planet_id
    AND pb.building = ud.element
    AND pb.level = ud.current_level;

  -- 2. Update the resources on this planet based on the
  -- type of building that has been completed. We will
  -- focus on updating the storage capacity and prod for
  -- each resource.
  -- 2.a) Update resources to reach the current time.
  PERFORM update_resources_for_planet(planet_id);

  -- 2.b) Proceed to update the mines with their new prod
  -- values.
  WITH update_data
    AS (
      SELECT resource, SUM(production_change) AS prod_change
      FROM
        construction_actions_buildings_production_effects cabpe
        INNER JOIN construction_actions_buildings cab ON cab.id = cabpe.action
      WHERE
        cab.planet = planet_id AND
        cab.completion_time < processing_time
      GROUP BY
        cabpe.resource
    )
  UPDATE planets_resources pr
    SET production = production + ud.prod_change
  FROM update_data AS ud
  WHERE
    pr.planet = planet_id
    AND pr.res = ud.resource;

  -- 2.c) Update the storage facilities with their new
  -- values.
  WITH update_data
    AS (
      SELECT resource, SUM(storage_capacity_change) AS capacity_change
      FROM
        construction_actions_buildings_storage_effects cabse
        INNER JOIN construction_actions_buildings cab ON cab.id = cabse.action
      WHERE
        cab.planet = planet_id AND
        cab.completion_time < processing_time
      GROUP BY
        cabse.resource
    )
  UPDATE planets_resources pr
    SET storage_capacity = storage_capacity + ud.capacity_change
  FROM update_data AS ud
  WHERE
    pr.planet = planet_id
    AND pr.res = ud.resource;

  -- 3. Destroy the processed actions effects.
  DELETE FROM
    construction_actions_buildings_production_effects cabpe
    USING construction_actions_buildings cab
  WHERE
    cabpe.action = cab.id AND
    cab.planet = planet_id AND
    cab.completion_time < processing_time;

  DELETE FROM
    construction_actions_buildings_storage_effects cabse
    USING construction_actions_buildings cab
  WHERE
    cabse.action = cab.id AND
    cab.planet = planet_id AND
    cab.completion_time < processing_time;

  -- 4. And finally the processed actions.
  DELETE FROM construction_actions_buildings WHERE planet = planet_id AND completion_time < processing_time;
END
$$ LANGUAGE plpgsql;

-- Update upgrade action for technologies.
CREATE OR REPLACE FUNCTION update_technology_upgrade_action(player_id uuid) RETURNS VOID AS $$
DECLARE
  -- Save time: this will make sure that we can't run into
  -- problem where for example an action is not complete
  -- when the 1. is performed and complete when the 2. is
  -- performed (resulting in a ship never being built).
  processing_time TIMESTAMP := NOW();
BEGIN
  -- 1. Register actions that are now complete.
  WITH update_data
    AS (
      SELECT *
      FROM construction_actions_technologies
      WHERE
        player = player_id AND
        completion_time < processing_time
      ORDER BY
        desired_level ASC
    )
  UPDATE player_technologies pt
    SET level = ud.desired_level
  FROM update_data AS ud
  WHERE
    pt.player = player_id AND
    pt.technology = ud.element AND
    pt.level = ud.current_level;

  -- 2. Delete processed actions.
  DELETE FROM construction_actions_technologies WHERE player = player_id AND completion_time < processing_time;
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
  WITH update_data
    AS (
      SELECT *
      FROM construction_actions_ships
      WHERE
        planet = planet_id AND
        created_at + (amount - (remaining - 1)) * completion_time < processing_time
    )
  UPDATE planets_ships ps
    SET count = count - (ud.amount - ud.remaining) +
      LEAST(
        EXTRACT(EPOCH FROM processing_time - ud.created_at) / EXTRACT(EPOCH FROM ud.completion_time),
        CAST(ud.amount AS DOUBLE PRECISION)
      )
  FROM update_data AS ud
  WHERE
    ps.planet = planet_id AND
    ps.ship = ud.element;

  -- 2. Update remaining actions with an amount decreased by an amount
  -- consistent with the duration elapsed since the creation.
  UPDATE construction_actions_ships
    SET remaining = amount -
      LEAST(
        EXTRACT(EPOCH FROM processing_time - created_at) / EXTRACT(EPOCH FROM completion_time),
        CAST(amount AS DOUBLE PRECISION)
      )
  WHERE
    planet = planet_id AND
    created_at + (amount - (remaining - 1)) * completion_time < processing_time;

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
  WITH update_data
    AS (
      SELECT *
      FROM construction_actions_defenses
      WHERE
        planet = planet_id AND
        created_at + (amount - (remaining - 1)) * completion_time < processing_time
    )
  UPDATE planets_defenses pd
    SET count = count - (ud.amount - ud.remaining) +
      LEAST(
        EXTRACT(EPOCH FROM processing_time - ud.created_at) / EXTRACT(EPOCH FROM ud.completion_time),
        CAST(ud.amount AS DOUBLE PRECISION)
      )
  FROM update_data AS ud
  WHERE
    pd.planet = planet_id AND
    pd.defense = ud.element;

  -- 2. Update remaining actions with an amount decreased by an amount
  -- consistent with the duration elapsed since the creation.
  UPDATE construction_actions_defenses
    SET remaining = amount -
      LEAST(
        EXTRACT(EPOCH FROM processing_time - created_at) / EXTRACT(EPOCH FROM completion_time),
        CAST(amount AS DOUBLE PRECISION)
      )
  WHERE
    planet = planet_id AND
    created_at + (amount - (remaining - 1)) * completion_time < processing_time;

  -- 3.Delete actions that don't have any remaining effect.
  DELETE FROM construction_actions_defenses WHERE planet = planet_id AND remaining = 0;
END
$$ LANGUAGE plpgsql;

-- Utility script allowing to deposit the resources
-- carried by a fleet to the planet it belongs to.
-- The return value indicates whether the planet to
-- deposit the resources to (computed with the data
-- from the input fleet) actually existed.
CREATE OR REPLACE FUNCTION fleet_deposit_resources(fleet_id uuid) RETURNS BOOLEAN AS $$
DECLARE
  planet_id uuid;
BEGIN
  -- Retrieve the index of the planet associated to the
  -- fleet.
  SELECT planet INTO planet_id FROM fleets WHERE id = fleet_id AND planet IS NOT NULL;

  -- If the planet does not exist for this fleet, do not
  -- deposit resources. Whether it is an issue will be
  -- determined by the calling script.
  IF NOT FOUND THEN
    RETURN FALSE;
  END IF;

  -- Perform the update of the resources on the planet
  -- so as to be sure that the player gets the max of
  -- its production in case the new deposit brings the
  -- total over the storage capacity.
  PERFORM update_resources_for_planet(planet_id);

  -- Add the resources carried by the fleet to the
  -- destination planet and remove them from the
  -- fleet's resources.
  UPDATE planets_resources AS pr
    SET amount = pr.amount + fr.amount
  FROM
    fleet_resources fr
    INNER JOIN fleet_elements fe ON fr.fleet_element = fe.id
    INNER JOIN fleets f ON fe.fleet = f.id
  WHERE
    f.id = fleet_id
    AND pr.res = fr.resource
    AND pr.planet = planet_id;

  -- Remove the resources carried by this fleet.
  DELETE FROM
    fleet_resources AS fr
    USING
      fleet_elements AS fe,
      fleets AS f
  WHERE
    fr.fleet_element = fe.id
    AND fe.fleet = f.id
    AND f.id = fleet_id;

  RETURN TRUE;
END
$$ LANGUAGE plpgsql;

-- Perform updates to account for a transport fleet.
CREATE OR REPLACE FUNCTION fleet_transport(fleet_id uuid) RETURNS VOID AS $$
DECLARE
  planet_was_found BOOLEAN;
BEGIN
  -- Use the dedicated script to perform the deposit
  -- of the resources.
  SELECT fleet_deposit_resources(fleet_id) INTO planet_was_found;

  -- Raise an error in case the planet associated to
  -- the fleet does not exist.
  IF NOT planet_was_found THEN
    RAISE EXCEPTION 'Fleet % is not directed towards a planet', fleet_id;
  END IF;
END
$$ LANGUAGE plpgsql;

-- Perform updates to account for a deployment fleet.
CREATE OR REPLACE FUNCTION fleet_deployment(fleet_id uuid) RETURNS VOID AS $$
DECLARE
  planet_was_found BOOLEAN;
  planet_id uuid;
BEGIN
  -- Make the resources carried by the fleet drop on
  -- the target planet.
  SELECT fleet_deposit_resources(fleet_id) INTO planet_was_found;

  -- Raise an error in case the planet associated to
  -- the fleet does not exist.
  IF NOT planet_was_found THEN
    RAISE EXCEPTION 'Fleet % is not directed towards a planet', fleet_id;
  END IF;

  -- At this point we know that the fleet exist so
  -- we can fetch it.
  SELECT planet INTO planet_id FROM fleets WHERE id = fleet_id AND planet IS NOT NULL;

  -- Otherwise, assign the ships of the fleet as
  -- element of the player.
  UPDATE planets_ships AS ps
    SET count = ps.count + fs.count
  FROM
    fleet_ships fs
    INNER JOIN fleet_elements fe ON fs.fleet_element = fe.id
    INNER JOIN fleets f ON fe.fleet = f.id
  WHERE
    f.id = fleet_id
    AND ps.ship = fs.ship
    AND ps.planet = planet_id;

  -- Remove the ships associated to the components
  -- of this fleet.
  DELETE FROM
    fleet_ships AS fs
    USING
      fleet_elements AS fe,
      fleets AS f
  WHERE
    fs.fleet_element = fe.id
    AND fe.fleet = f.id
    AND f.id = fleet_id;

  -- We will also need to remove the components as
  -- they're not relevant anymore.
  DELETE FROM
    fleet_elements AS fe
    USING
      fleets AS f
  WHERE
    fe.fleet = f.id
    AND f.id = fleet_id;

  -- And finally remove the fleet which is now as
  -- empty as my bank account.
  DELETE FROM
    fleets
  WHERE
    id = fleet_id;
END
$$ LANGUAGE plpgsql;
