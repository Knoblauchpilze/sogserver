
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

    -- Decrease the amount of resources existing on the planet
    -- after the construction of this building. Note that we
    -- do not update the resources to the current time which
    -- might lead to negative values if it hasn't been done
    -- in a long time. We assume the update will be enforced
    -- by other processes.
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
        construction_actions_buildings cab
      WHERE
        cab.id = (upgrade->>'id')::uuid;
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
        'building_upgrade_moon' AS type
      FROM
        construction_actions_buildings_moon cabm
      WHERE
        cabm.id = (upgrade->>'id')::uuid;
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

    -- Decrease the amount of resources existing on the planet
    -- after the construction of this technology. Note that we
    -- do not update the resources to the current time which
    -- might lead to negative values if it hasn't been done
    -- in a long time. We assume the update will be enforced
    -- by other processes.
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
      construction_actions_technologies cat
    WHERE
      cat.id = (upgrade->>'id')::uuid;
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

    -- Decrease the amount of resources existing on the planet
    -- after the construction of this ship. Note that we do not
    -- update the resources to the current time which might lead
    -- to negative values if it hasn't been done in a long time.
    -- We assume the update will be enforced by other processes.
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
        construction_actions_ships cas
      WHERE
        cas.id = (upgrade->>'id')::uuid;
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
        'ship_upgrade_moon' AS type
      FROM
        construction_actions_ships_moon casm
      WHERE
        casm.id = (upgrade->>'id')::uuid;
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

    -- Decrease the amount of resources existing on the planet
    -- after the construction of this defense. Note that we
    -- do not update the resources to the current time which
    -- might lead to negative values if it hasn't been done
    -- in a long time. We assume the update will be enforced
    -- by other processes.
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
        construction_actions_defenses cad
      WHERE
        cad.id = (upgrade->>'id')::uuid;
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
        'defense_upgrade_moon' AS type
      FROM
        construction_actions_defenses_moon cadm
      WHERE
        cadm.id = (upgrade->>'id')::uuid;
  END IF;
END
$$ LANGUAGE plpgsql;

-- Update the resources available on a planet until the
-- provided `moment` in time. Note that if the `moment`
-- is not consistent with the last time this planet has
-- been updated we might run into trouble.
CREATE OR REPLACE FUNCTION update_resources_for_planet_to_time(planet_id uuid, moment TIMESTAMP WITH TIME ZONE) RETURNS VOID AS $$
BEGIN
  -- Update the amount of resource to be at most the storage
  -- capacity, and otherwise to increase by the duration that
  -- passed between the last update and the current time.
  -- Note that even if the production is expressed in hours,
  -- we need to extract the number of seconds in order to be
  -- able to obtain fractions of an hour to update the value.
  UPDATE planets_resources
  SET
    amount = amount + EXTRACT(EPOCH FROM moment - updated_at) * production / 3600.0,
    updated_at = moment
  FROM
    resources AS r
  WHERE
    planet = planet_id
    AND res = r.id
    AND r.storable='true';
END
$$ LANGUAGE plpgsql;

-- Update resources for a planet. This method will pick
-- the current time as the base to update the resources.
CREATE OR REPLACE FUNCTION update_resources_for_planet(planet_id uuid) RETURNS VOID AS $$
DECLARE
  processing_time TIMESTAMP WITH TIME ZONE := NOW();
BEGIN
  -- Just use the base script.
  PERFORM update_resources_for_planet_to_time(planet_id, processing_time);
END
$$ LANGUAGE plpgsql;

-- Update upgrade action for buildings.
CREATE OR REPLACE FUNCTION update_building_upgrade_action(action_id uuid, kind text) RETURNS VOID AS $$
DECLARE
  moment timestamp with time zone;
  planet_id uuid;
BEGIN
  -- We can have building upgrades both for planets and moons.
  -- These actions are stored in different tables and we don't
  -- have a way to determine which beforehand. The `kind` helps
  -- define this: we just need to make sure it is correct.
  IF kind != 'planet' AND kind != 'moon' THEN
    RAISE EXCEPTION 'Invalid kind % specified for %', kind, action_id;
  END IF;

  IF kind = 'planet' THEN
    -- 1. Update the level of the building described by the
    -- action on the corresponding planet.
    UPDATE planets_buildings AS pb
      SET level = cab.desired_level
    FROM
      construction_actions_buildings AS cab
    WHERE
      cab.id = action_id
      AND pb.planet = cab.planet
      AND pb.building = cab.element
      AND pb.level = cab.current_level;

    -- 2. Update the resources on this planet based on the
    -- type of building that has been completed. Before it
    -- can happen we need to bring the production of the
    -- planet to its value at the time of the update of
    -- the building.
    -- 2.a) Update resources to reach the current time.
    SELECT completion_time INTO moment FROM construction_actions_buildings WHERE id = action_id;

    IF NOT FOUND THEN
      RAISE EXCEPTION 'Unable to fetch completion time for action %', action_id;
    END IF;

    SELECT planet INTO planet_id FROM construction_actions_buildings WHERE id = action_id;

    IF NOT FOUND THEN
      RAISE EXCEPTION 'Unable to fetch planet id for action %', action_id;
    END IF;

    PERFORM update_resources_for_planet(planet_id, moment);

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
    UPDATE moons_buildings AS mb
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
CREATE OR REPLACE FUNCTION update_technology_upgrade_action(action_id uuid) RETURNS VOID AS $$
BEGIN
  -- 1. Register actions that are now complete.
  UPDATE players_technologies AS pt
    SET level = cat.desired_level
  FROM
    construction_actions_technologies AS cat
  WHERE
    cat.id = action_id
    AND pt.player = cat.player
    AND pt.technology = cat.element
    AND pt.level = cat.current_level;

  -- 2. Remove the processed action from the events queue.
  DELETE FROM actions_queue WHERE action = action_id;

  -- 3. And finally delete the processed action.
  DELETE FROM construction_actions_technologies WHERE id = action_id;
END
$$ LANGUAGE plpgsql;

-- Update upgrade action for ships.
CREATE OR REPLACE FUNCTION update_ship_upgrade_action(action_id uuid, kind text) RETURNS VOID AS $$
DECLARE
  -- Save time: this will make sure that we can't run into
  -- problem where for example an action is not complete
  -- when the 1. is performed and complete when the 2. is
  -- performed (resulting in a ship never being built).
  processing_time timestamp with time zone := NOW();
BEGIN
  -- We can have building upgrades both for planets and moons.
  -- These actions are stored in different tables and we don't
  -- have a way to determine which beforehand. The `kind` helps
  -- define this: we just need to make sure it is correct.
  IF kind != 'planet' AND kind != 'moon' THEN
    RAISE EXCEPTION 'Invalid kind % specified for %', kind, action_id;
  END IF;

  IF kind = 'planet' THEN
    -- 1. Register ships that are now complete. We need to account
    -- for the fact that several elements might have completed while
    -- we were not updating this action.
    -- The algorithm is basically:
    --   - compute the number of intervals that have elapsed.
    --   - subtract the number of already built elements.
    --   - clamp to make sure that we don't create too many elements.
    UPDATE planets_ships AS ps
      SET count = count - (cas.amount - cas.remaining) +
        LEAST(
          EXTRACT(EPOCH FROM processing_time - cas.created_at) / EXTRACT(EPOCH FROM cas.completion_time),
          CAST(cas.amount AS DOUBLE PRECISION)
        )
      FROM
        construction_actions_ships AS cas
      WHERE
        cas.id = action_id
        AND ps.planet = cas.planet
        AND ps.ship = cas.element;

    -- 2. Update remaining action with an amount decreased by an
    -- amount consistent with the duration elapsed since the creation.
    UPDATE construction_actions_ships
      SET remaining = amount -
        LEAST(
          EXTRACT(EPOCH FROM processing_time - created_at) / EXTRACT(EPOCH FROM completion_time),
          CAST(amount AS DOUBLE PRECISION)
        )
    WHERE
      id = action_id;

    -- 3. Update elements in actions queue based on the next completion time.
    UPDATE actions_queue AS aq
      SET completion_time = cas.created_at + (1 + cas.amount - cas.remaining) * cas.completion_time
      FROM
        construction_actions_ships AS cas
      WHERE
        aq.action = action_id
        AND cas.id = action_id;

    DELETE FROM
      actions_queue AS aq
      USING construction_actions_ships AS cas
    WHERE
      aq.action = cas.id
      AND aq.action = action_id
      AND cas.remaining = 0;

    -- 4. Delete actions that don't have any remaining effect.
    DELETE FROM construction_actions_ships WHERE id = action_id AND remaining = 0;
  END IF;

  IF kind = 'moon' THEN
    -- 1. See comment in above section.
    UPDATE moons_ships AS ms
      SET count = count - (casm.amount - casm.remaining) +
        LEAST(
          EXTRACT(EPOCH FROM processing_time - casm.created_at) / EXTRACT(EPOCH FROM casm.completion_time),
          CAST(casm.amount AS DOUBLE PRECISION)
        )
      FROM
        construction_actions_ships_moon AS casm
      WHERE
        casm.id = action_id
        AND ms.planet = casm.planet
        AND ms.ship = casm.element;

    -- 2. See comment in above section.
    UPDATE construction_actions_ships_moon
      SET remaining = amount -
        LEAST(
          EXTRACT(EPOCH FROM processing_time - created_at) / EXTRACT(EPOCH FROM completion_time),
          CAST(amount AS DOUBLE PRECISION)
        )
    WHERE
      id = action_id;

    -- 3. See comment in above section.
    UPDATE actions_queue AS aq
      SET completion_time = casm.created_at + (1 + casm.amount - casm.remaining) * casm.completion_time
      FROM
        construction_actions_ships_moon AS casm
      WHERE
        aq.action = action_id
        AND casm.id = action_id;

    DELETE FROM
      actions_queue AS aq
      USING construction_actions_ships_moon AS casm
    WHERE
      aq.action = casm.id
      AND aq.action = action_id
      AND casm.remaining = 0;

    -- 4. See comment in above section.
    DELETE FROM construction_actions_ships_moon WHERE id = action_id AND remaining = 0;
  END IF;
END
$$ LANGUAGE plpgsql;

-- Update upgrade action for defenses.
CREATE OR REPLACE FUNCTION update_defense_upgrade_action(action_id uuid, kind text) RETURNS VOID AS $$
DECLARE
  -- Similar mechanism to the one used for ships.
  processing_time timestamp with time zone := NOW();
BEGIN
  -- We can have building upgrades both for planets and moons.
  -- These actions are stored in different tables and we don't
  -- have a way to determine which beforehand. The `kind` helps
  -- define this: we just need to make sure it is correct.
  IF kind != 'planet' AND kind != 'moon' THEN
    RAISE EXCEPTION 'Invalid kind % specified for %', kind, action_id;
  END IF;

  IF kind = 'planet' THEN
    -- 1. Register defenses that are now complete. We need to
    -- account for the fact that several elements might have
    -- completed while we were not updating this action.
    -- The algorithm is basically:
    --   - compute the number of intervals that have elapsed.
    --   - subtract the number of already built elements.
    --   - clamp to make sure that we don't create too many elements.
    UPDATE planets_defenses AS pd
      SET count = count - (cad.amount - cad.remaining) +
        LEAST(
          EXTRACT(EPOCH FROM processing_time - cad.created_at) / EXTRACT(EPOCH FROM cad.completion_time),
          CAST(cad.amount AS DOUBLE PRECISION)
        )
      FROM
        construction_actions_defenses AS cad
      WHERE
        cad.id = action_id
        AND pd.planet = cad.planet
        AND pd.defense = cad.element;

    -- 2. Update remaining action with an amount decreased by an
    -- amount consistent with the duration elapsed since the creation.
    UPDATE construction_actions_defenses
      SET remaining = amount -
        LEAST(
          EXTRACT(EPOCH FROM processing_time - created_at) / EXTRACT(EPOCH FROM completion_time),
          CAST(amount AS DOUBLE PRECISION)
        )
    WHERE
      id = action_id;

    -- 3. Update elements in actions queue based on the next completion time.
    UPDATE actions_queue AS aq
      SET completion_time = cad.created_at + (1 + cad.amount - cad.remaining) * cad.completion_time
      FROM
        construction_actions_defenses AS cad
      WHERE
        aq.action = action_id
        AND cad.id = action_id;

    DELETE FROM
      actions_queue AS aq
      USING construction_actions_defenses AS cad
    WHERE
      aq.action = cad.id
      AND aq.action = action_id
      AND cad.remaining = 0;

    -- 4. Delete actions that don't have any remaining effect.
    DELETE FROM construction_actions_defenses WHERE id = action_id AND remaining = 0;
  END IF;

  IF kind = 'moon' THEN
    -- 1. See comment in above section.
    UPDATE planets_defenses AS pd
      SET count = count - (cadm.amount - cadm.remaining) +
        LEAST(
          EXTRACT(EPOCH FROM processing_time - cadm.created_at) / EXTRACT(EPOCH FROM cadm.completion_time),
          CAST(cadm.amount AS DOUBLE PRECISION)
        )
      FROM
        construction_actions_defenses_moon AS cadm
      WHERE
        cadm.id = action_id
        AND pd.planet = cadm.planet
        AND pd.defense = cadm.element;

    -- 2. See comment in above section.
    UPDATE construction_actions_defenses_moon
      SET remaining = amount -
        LEAST(
          EXTRACT(EPOCH FROM processing_time - created_at) / EXTRACT(EPOCH FROM completion_time),
          CAST(amount AS DOUBLE PRECISION)
        )
    WHERE
      id = action_id;

    -- 3. Update elements in actions queue based on the next completion time.
    UPDATE actions_queue AS aq
      SET completion_time = cadm.created_at + (1 + cadm.amount - cadm.remaining) * cadm.completion_time
      FROM
        construction_actions_ships AS cadm
      WHERE
        aq.action = action_id
        AND cadm.id = action_id;

    DELETE FROM
      actions_queue AS aq
      USING construction_actions_defenses_moon AS cadm
    WHERE
      aq.action = cadm.id
      AND aq.action = action_id
      AND cadm.remaining = 0;

    -- 4. Delete actions that don't have any remaining effect.
    DELETE FROM construction_actions_defenses_moon WHERE id = action_id AND remaining = 0;
  END IF;
END
$$ LANGUAGE plpgsql;