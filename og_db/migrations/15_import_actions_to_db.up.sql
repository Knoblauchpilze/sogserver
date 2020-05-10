
-- Import building upgrade action in the dedicated table.
CREATE OR REPLACE FUNCTION create_building_upgrade_action(upgrade json, costs json, production_effects json, storage_effects json, kind text) RETURNS VOID AS $$
BEGIN
  -- The `kind` can reference either a planet or a moon.
  -- We have to make sure that it's a valid value before
  -- attempting to use it.
  IF kind != 'planet' AND kind != 'moon' THEN
    RAISE EXCEPTION 'Invalid kind % specified for building action', kind;
  END IF;

  IF kind = 'planet' THEN
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
    UPDATE planets_resources
      SET amount = amount - rc.cost
    FROM
      rc
    WHERE
      planet = (upgrade->>'planet')::uuid
      AND res = rc.resource;

    -- Register this action in the actions system.
    INSERT INTO actions_queue
      SELECT
        cab.id AS action,
        cab.completion_time AS completion_time,
        'building_upgrade' AS type
      FROM
        construction_actions_buildings cab;
  END IF;

  IF kind = 'moon' THEN
    -- Create the building upgrade action itself.
    INSERT INTO construction_actions_buildings_moon
      SELECT *
      FROM json_populate_record(null::construction_actions_buildings_moon, upgrade);

    -- No production or storage effects available
    -- on moons as most of the buildings are not
    -- allowed to be built. We still need to take
    -- out the resources needed by the action.
    WITH rc AS (
      SELECT
        t.resource,
        t.cost
      FROM
        json_to_recordset(costs) AS t(resource uuid, cost numeric(15, 5))
      )
    UPDATE moons_resources
      SET amount = amount - rc.cost
    FROM
      rc
    WHERE
      moon = (upgrade->>'planet')::uuid
      AND res = rc.resource;

    -- Register this action in the actions system.
    INSERT INTO actions_queue
      SELECT
        cabm.id AS action,
        cabm.completion_time AS completion_time,
        'building_upgrade' AS type
      FROM
        construction_actions_buildings_moon cabm;
  END IF;
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
  UPDATE planets_resources
    SET amount = amount - rc.cost
  FROM
    rc
  WHERE
    planet = (upgrade->>'planet')::uuid
    AND res = rc.resource;

  -- Register this action in the actions system.
  INSERT INTO actions_queue
    SELECT
      cat.id AS action,
      cat.completion_time AS completion_time,
      'technology_upgrade' AS type
    FROM
      construction_actions_technologies cat;
END
$$ LANGUAGE plpgsql;

-- Import ship upgrade action in the dedicated table.
CREATE OR REPLACE FUNCTION create_ship_upgrade_action(upgrade json, costs json, kind text) RETURNS VOID AS $$
BEGIN
  -- Make sure the kind describes a known action.
  IF kind != 'planet' AND kind != 'moon' THEN
    RAISE EXCEPTION 'Invalid kind % specified for ship action', kind;
  END IF;

  IF kind = 'planet' THEN
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
    UPDATE planets_resources
      SET amount = amount - rc.cost
    FROM
      rc
    WHERE
      planet = (upgrade->>'planet')::uuid
      AND res = rc.resource;

    -- Register this action in the actions system. Note
    -- that the completion time will be computed from
    -- the actual creation time for this action and the
    -- duration of the construction of a single element.
    INSERT INTO actions_queue
      SELECT
        cas.id AS action,
        cas.created_at + cas.completion_time AS completion_time,
        'ship_upgrade' AS type
      FROM
        consturction_actions_ships cas;
  END IF;

  IF kind = 'moon' THEN
    -- Insert the construction action in the related table.
    INSERT INTO construction_actions_ships_moon
      SELECT *
      FROM json_populate_record(null::construction_actions_ships_moon, upgrade);

    -- No need to update the resources of the moon but we
    -- need to deduct the cost of the action.
    WITH rc AS (
      SELECT
        t.resource,
        t.cost
      FROM
        json_to_recordset(costs) AS t(resource uuid, cost numeric(15, 5))
      )
    UPDATE moons_resources
      SET amount = amount - rc.cost
    FROM
      rc
    WHERE
      moon = (upgrade->>'planet')::uuid
      AND res = rc.resource;

    -- See comment in above section.
    INSERT INTO actions_queue
      SELECT
        casm.id AS action,
        casm.created_at + casm.completion_time AS completion_time,
        'ship_upgrade' AS type
      FROM
        construction_actions_ships_moon casm;
  END IF;
END
$$ LANGUAGE plpgsql;

-- Import defense upgrade action in the dedicated table.
CREATE OR REPLACE FUNCTION create_defense_upgrade_action(upgrade json, costs json, kind text) RETURNS VOID AS $$
BEGIN
  -- The `kind` can reference either a planet or a moon.
  -- We have to make sure that it's a valid value before
  -- attempting to use it.
  IF kind != 'planet' AND kind != 'moon' THEN
    RAISE EXCEPTION 'Invalid kind % specified for defense action', kind;
  END IF;

  IF kind = 'planet' THEN
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
    UPDATE planets_resources
      SET amount = amount - rc.cost
    FROM
      rc
    WHERE
      planet = (upgrade->>'planet')::uuid
      AND res = rc.resource;

    -- Register this action in the actions system. Note
    -- that the completion time will be computed from
    -- the actual creation time for this action and the
    -- duration of the construction of a single element.
    INSERT INTO actions_queue
      SELECT
        cad.id AS action,
        cad.created_at + cad.completion_time AS completion_time,
        'defense_upgrade' AS type
      FROM
        construction_actions_defenses cad;
  END IF;

  IF kind = 'moon' THEN
    -- Insert the construction action in the related table.
    INSERT INTO construction_actions_defenses_moon
      SELECT *
      FROM json_populate_record(null::construction_actions_defenses_moon, upgrade);

    -- No need to update the resources of the moon but we
    -- need to deduct the cost of the action.
    WITH rc AS (
      SELECT
        t.resource,
        t.cost
      FROM
        json_to_recordset(costs) AS t(resource uuid, cost numeric(15, 5))
      )
    UPDATE moons_resources
      SET amount = amount - rc.cost
    FROM
      rc
    WHERE
      moon = (upgrade->>'planet')::uuid
      AND res = rc.resource;

    -- See comment in above section.
    INSERT INTO actions_queue
      SELECT
        cadm.id AS action,
        cadm.created_at + cadm.completion_time AS completion_time,
        'defense_upgrade' AS type
      FROM
        construction_actions_defenses_moon cadm;
  END IF;
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
CREATE OR REPLACE FUNCTION update_building_upgrade_action(action_id uuid, kind text) RETURNS VOID AS $$
BEGIN
  -- The `action_id` can reference an action that is
  -- existing either for a planet or a moon. This is
  -- specified by the `kind` in input which allows
  -- to select between both cases.
  -- The process to update the actions is very similar
  -- in any of these two cases, the only variable is
  -- the name of the table. For now we will rely on
  -- the `kind` and copy paste the code for both cases.
  IF kind != 'planet' AND kind != 'moon' THEN
    RAISE EXCEPTION 'Invalid kind % specified for %', kind, action_id;
  END IF;

  IF kind = 'planet' THEN
    -- 1. Update the action by updating the level of
    -- the building described by the action.
    UPDATE planets_buildings pb
      SET level = cab.desired_level
    FROM
      construction_actions_buildings AS cab
    WHERE
      cab.id = action_id
      AND pb.planet = cab.planet
      AND pb.building = cab.element
      AND pb.level = cab.current_level;

    -- 2. Update the resources on this planet based on the
    -- type of building that has been completed. We will
    -- focus on updating the storage capacity and prod for
    -- each resource.
    -- 2.a) Update resources to reach the current time.
    -- TODO: Should provide a processing time for this action
    -- so that we update until this specified time. Maybe it
    -- could correspond to the completion time of the action.
    PERFORM update_resources_for_planet(target_id);

    -- 2.b) Proceed to update the mines with their new prod
    -- values.
    UPDATE planets_resources AS pr
      SET production = production + cabpe.production_change
    FROM
      construction_actions_buildings_production_effects AS cabpe
      INNER JOIN construction_actions_buildings AS cab ON cabpe.action = cab.id
    WHERE
      cabpe.action = action_id
      AND pr.planet = cab.planet
      AND pr.res = cabpe.resource;

    -- 2.c) Update the storage facilities with their new
    -- values.
    UPDATE planets_resources AS pr
      SET storage_capacity = storage_capacity + cabse.storage_capacity_change
    FROM
      construction_actions_buildings_storage_effects AS cabse
      INNER JOIN construction_actions_buildings AS cab ON cabse.action = cab.id
    WHERE
      cabse.action = action_id
      AND pr.planet = cab.planet
      AND pr.res = cabse.resource;

    -- 3. Destroy the processed action effects.
    DELETE FROM
      construction_actions_buildings_production_effects cabpe
      USING construction_actions_buildings cab
    WHERE
      cabpe.action = cab.id AND
      cab.id = action_id;

    DELETE FROM
      construction_actions_buildings_storage_effects cabse
      USING construction_actions_buildings cab
    WHERE
      cabse.action = cab.id AND
      cab.id = action_id;

    -- 4. Remove the processed action from the events queue.
    DELETE FROM actions_queue WHERE action = action_id;

    -- 5. And finally delete the processed action.
    DELETE FROM construction_actions_buildings WHERE id = action_id;
  END IF;

  IF kind = 'moon' THEN
    -- 1. See comment in above section.
    UPDATE moons_buildings mb
      SET level = cabm.desired_level
    FROM
      construction_actions_buildings_moon AS cabm
    WHERE
      cabm.id = action_id
      AND mb.moon = cabm.moon
      AND mb.building = cabm.element
      AND mb.level = cabm.current_level;

    -- 2. No need to update the resources, there's no prod
    -- that can happen on a moon (at least for now).

    -- 3. As no effects can be applied there's no need to
    -- delete the corresponding lines in tables.

    -- 4. See comment in above section.
    DELETE FROM actions_queue WHERE action = action_id;

    -- 5. See comment in above section.
    DELETE FROM construction_actions_buildings_moon WHERE id = action_id;
  END IF;
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
  UPDATE players_technologies pt
    SET level = ud.desired_level
  FROM update_data AS ud
  WHERE
    pt.player = player_id AND
    pt.technology = ud.element AND
    pt.level = ud.current_level;

  -- 2. Remove the processed actions from the events queue.
    DELETE FROM
      actions_queue
      USING construction_actions_technologies cat
    WHERE
      cat.planet = target_id
      AND cat.completion_time < processing_time;

  -- 3. Delete processed actions.
  DELETE FROM construction_actions_technologies WHERE player = player_id AND completion_time < processing_time;
END
$$ LANGUAGE plpgsql;

-- Update upgrade action for ships.
CREATE OR REPLACE FUNCTION update_ship_upgrade_action(target_id uuid, kind text) RETURNS VOID AS $$
DECLARE
  -- Save time: this will make sure that we can't run into
  -- problem where for example an action is not complete
  -- when the 1. is performed and complete when the 2. is
  -- performed (resulting in a ship never being built).
  processing_time TIMESTAMP := NOW();
BEGIN
  -- The `target_id` can reference either a planet or a moon.
  -- See comments in `update_building_upgrade_action` to get
  -- more info.
  IF kind != 'planet' AND kind != 'moon' THEN
    RAISE EXCEPTION 'Invalid kind % specified for %', kind, target_id;
  END IF;

  IF kind = 'planet' THEN
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
          planet = target_id AND
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
      ps.planet = target_id AND
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
      planet = target_id AND
      created_at + (amount - (remaining - 1)) * completion_time < processing_time;

    -- 3. Update elements in actions queue based on the next completion time.
    -- TODO: Handle this.

    -- 4. Delete actions that don't have any remaining effect.
    DELETE FROM construction_actions_ships WHERE planet = target_id AND remaining = 0;
  END IF;

  IF kind = 'moon' THEN
    -- 1. See comment in above section.
    WITH update_data
      AS (
        SELECT *
        FROM construction_actions_ships_moon
        WHERE
          moon = target_id AND
          created_at + (amount - (remaining - 1)) * completion_time < processing_time
      )
    UPDATE moons_ships ms
      SET count = count - (ud.amount - ud.remaining) +
        LEAST(
          EXTRACT(EPOCH FROM processing_time - ud.created_at) / EXTRACT(EPOCH FROM ud.completion_time),
          CAST(ud.amount AS DOUBLE PRECISION)
        )
    FROM update_data AS ud
    WHERE
      ms.moon = target_id AND
      ms.ship = ud.element;

    -- 2. See comment in above section.
    UPDATE construction_actions_ships_moon
      SET remaining = amount -
        LEAST(
          EXTRACT(EPOCH FROM processing_time - created_at) / EXTRACT(EPOCH FROM completion_time),
          CAST(amount AS DOUBLE PRECISION)
        )
    WHERE
      moon = target_id AND
      created_at + (amount - (remaining - 1)) * completion_time < processing_time;

    -- 3. See comment in above section.
    -- TODO: Handle this.

    -- 4. See comment in above section.
    DELETE FROM construction_actions_ships_moon WHERE moon = target_id AND remaining = 0;
  END IF;
END
$$ LANGUAGE plpgsql;

-- Update upgrade action for defenses.
CREATE OR REPLACE FUNCTION update_defense_upgrade_action(target_id uuid, kind text) RETURNS VOID AS $$
DECLARE
  -- Similar mechanism to the one used for ships.
  processing_time TIMESTAMP := NOW();
BEGIN
  -- The `target_id` can reference either a planet or a moon.
  -- See comments in `update_building_upgrade_action` to get
  -- more info.
  IF kind != 'planet' AND kind != 'moon' THEN
    RAISE EXCEPTION 'Invalid kind % specified for %', kind, target_id;
  END IF;

  IF kind = 'planet' THEN
    -- 1. Register defenses that are now complete. We need to
    -- account for the fact that several elements might have
    -- completed while we were not updating this action.
    -- The algorithm is basically:
    --   - compute the number of intervals that have elapsed.
    --   - subtract the number of already built elements.
    --   - clamp tp make sure that we don't create too many
    --     elements.
    WITH update_data
      AS (
        SELECT *
        FROM construction_actions_defenses
        WHERE
          planet = target_id AND
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
      pd.planet = target_id AND
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
      planet = target_id AND
      created_at + (amount - (remaining - 1)) * completion_time < processing_time;


    -- 3. Update elements in actions queue based on the next completion time.
    -- TODO: Handle this.

    -- 4. Delete actions that don't have any remaining effect.
    DELETE FROM construction_actions_defenses WHERE planet = target_id AND remaining = 0;
  END IF;

  IF kind = 'moon' THEN
    -- 1. See comment in above section.
    WITH update_data
      AS (
        SELECT *
        FROM construction_actions_defenses_moon
        WHERE
          moon = target_id AND
          created_at + (amount - (remaining - 1)) * completion_time < processing_time
      )
    UPDATE moons_defenses md
      SET count = count - (ud.amount - ud.remaining) +
        LEAST(
          EXTRACT(EPOCH FROM processing_time - ud.created_at) / EXTRACT(EPOCH FROM ud.completion_time),
          CAST(ud.amount AS DOUBLE PRECISION)
        )
    FROM update_data AS ud
    WHERE
      md.moon = target_id AND
      md.defense = ud.element;

    -- 2. See comment in above section.
    UPDATE construction_actions_defenses_moon
      SET remaining = amount -
        LEAST(
          EXTRACT(EPOCH FROM processing_time - created_at) / EXTRACT(EPOCH FROM completion_time),
          CAST(amount AS DOUBLE PRECISION)
        )
    WHERE
      moon = target_id AND
      created_at + (amount - (remaining - 1)) * completion_time < processing_time;

    -- 3. Update elements in actions queue based on the next completion time.
    -- TODO: Handle this.

    -- 4. See comment in above section.
    DELETE FROM construction_actions_defenses_moon WHERE moon = target_id AND remaining = 0;
  END IF;
END
$$ LANGUAGE plpgsql;
