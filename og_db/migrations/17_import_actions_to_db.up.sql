
-- Import building upgrade action in the dedicated table.
CREATE OR REPLACE FUNCTION create_building_upgrade_action(upgrade json, costs json, production_effects json, storage_effects json, fields_effects json, kind text) RETURNS VOID AS $$
DECLARE
  processing_time TIMESTAMP WITH TIME ZONE := NOW();
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

    INSERT INTO construction_actions_buildings_fields_effects
      SELECT
        (upgrade->>'id')::uuid,
        additional_fields
      FROM
        json_populate_record(null::construction_actions_buildings_fields_effects, fields_effects)
      WHERE
        additional_fields > 0;

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
    UPDATE planets_resources AS pr
      SET amount = pr.amount - rc.cost
    FROM
      rc
      INNER JOIN resources AS r ON rc.resource = r.id
    WHERE
      pr.planet = (upgrade->>'planet')::uuid
      AND pr.res = rc.resource
      AND r.storable = 'true';

    -- Update the last activity time for this planet.
    UPDATE planets SET last_activity = processing_time WHERE id = (upgrade->>'planet')::uuid;

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

    INSERT INTO construction_actions_buildings_fields_effects_moon
      SELECT
        (upgrade->>'id')::uuid,
        additional_fields
      FROM
        json_populate_record(null::construction_actions_buildings_fields_effects_moon, fields_effects)
      WHERE
        additional_fields > 0;

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
    UPDATE moons_resources AS mr
      SET amount = mr.amount - rc.cost
    FROM
      rc
      INNER JOIN resources AS r ON rc.resource = r.id
    WHERE
      mr.moon = (upgrade->>'planet')::uuid
      AND mr.res = rc.resource
      AND r.storable = 'true';

    -- Update the last activity time for this moon.
    UPDATE moons SET last_activity = processing_time WHERE id = (upgrade->>'planet')::uuid;

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
DECLARE
  processing_time TIMESTAMP WITH TIME ZONE := NOW();
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
  UPDATE planets_resources AS pr
    SET amount = pr.amount - rc.cost
  FROM
    rc
    INNER JOIN resources AS r ON rc.resource = r.id
  WHERE
    pr.planet = (upgrade->>'planet')::uuid
    AND pr.res = rc.resource
    AND r.storable = 'true';

  -- Update the last activity time for this planet.
  UPDATE planets SET last_activity = processing_time WHERE id = (upgrade->>'planet')::uuid;

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
DECLARE
  processing_time TIMESTAMP WITH TIME ZONE := NOW();
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
    UPDATE planets_resources AS pr
      SET amount = pr.amount - rc.cost
    FROM
      rc
      INNER JOIN resources AS r ON rc.resource = r.id
    WHERE
      pr.planet = (upgrade->>'planet')::uuid
      AND pr.res = rc.resource
      AND r.storable = 'true';

    -- Update the last activity time for this planet.
    UPDATE planets SET last_activity = processing_time WHERE id = (upgrade->>'planet')::uuid;

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
    UPDATE moons_resources AS mr
      SET amount = mr.amount - rc.cost
    FROM
      rc
      INNER JOIN resources AS r ON rc.resource = r.id
    WHERE
      mr.moon = (upgrade->>'planet')::uuid
      AND mr.res = rc.resource
      AND r.storable = 'true';

    -- Update the last activity time for this moon.
    UPDATE moons SET last_activity = processing_time WHERE id = (upgrade->>'planet')::uuid;

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
DECLARE
  processing_time TIMESTAMP WITH TIME ZONE := NOW();
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
    UPDATE planets_resources AS pr
      SET amount = pr.amount - rc.cost
    FROM
      rc
      INNER JOIN resources AS r ON rc.resource = r.id
    WHERE
      pr.planet = (upgrade->>'planet')::uuid
      AND pr.res = rc.resource
      AND r.storable = 'true';

    -- Update the last activity time for this planet.
    UPDATE planets SET last_activity = processing_time WHERE id = (upgrade->>'planet')::uuid;

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
    UPDATE moons_resources AS mr
      SET amount = mr.amount - rc.cost
    FROM
      rc
      INNER JOIN resources AS r ON rc.resource = r.id
    WHERE
      mr.moon = (upgrade->>'planet')::uuid
      AND mr.res = rc.resource
      AND r.storable = 'true';

    -- Update the last activity time for this moon.
    UPDATE moons SET last_activity = processing_time WHERE id = (upgrade->>'planet')::uuid;

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
  -- The final amount is obtained by first taking the maximum
  -- value between the current amount and the storage capacity
  -- (so that we keep excess brought by fleets for example)
  -- and we take the minimum of this and the amount that would
  -- exist if the production was still running (assuming that
  -- if the current value is less than that it means that we
  -- reached the storage capacity).
  UPDATE planets_resources
  SET
    amount = LEAST(
      amount + EXTRACT(EPOCH FROM NOW() - updated_at) * production / 3600.0,
      GREATEST(
        amount,
        storage_capacity
      )
    ),
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

    PERFORM update_resources_for_planet_to_time(planet_id, moment);

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

    -- 2.d) Update the fields available on the planet based
    -- on the additional fields produced by the building.
    UPDATE planets AS p
      SET fields = fields + cabfe.additional_fields
    FROM
      construction_actions_buildings_fields_effects AS cabfe
      INNER JOIN construction_actions_buildings AS cab ON cabfe.action = cab.id
    WHERE
      cabfe.action = action_id
      AND p.id = cab.planet;

    -- 2.e) Update the last activity time for this planet.
    UPDATE planets SET last_activity = moment WHERE id = planet_id;

    -- 2.f) Add the cost of the action to the points of the
    -- player in the economy section. The cost is directly
    -- registered in the action.
    UPDATE players_points
      SET economy_points = economy_points + points
    FROM
      construction_actions_buildings AS cab
      INNER JOIN planets AS p ON cab.planet = p.id
    WHERE
      cab.id = action_id
      AND p.player = players_points.player;

      -- 2.g) Add the cost of this building to the total
      -- cost for this building on this planet.
    UPDATE planets_buildings
      SET points = planets_buildings.points + cab.points
    FROM
      construction_actions_buildings AS cab
    WHERE
      cab.id = action_id
      AND planets_buildings.planet = cab.planet
      AND planets_buildings.building = cab.element;

    -- 3. Destroy the processed action effects.
    DELETE FROM construction_actions_buildings_production_effects WHERE action = action_id;

    DELETE FROM construction_actions_buildings_storage_effects WHERE action = action_id;

    DELETE FROM construction_actions_buildings_fields_effects WHERE action = action_id;

    -- 4. Remove the processed action from the events queue.
    DELETE FROM actions_queue WHERE action = action_id;

    -- 5. And finally delete the processed action.
    DELETE FROM construction_actions_buildings WHERE id = action_id;
  END IF;

  IF kind = 'moon' THEN
    -- Fetch the identifier of the moon.
    SELECT moon INTO planet_id FROM construction_actions_buildings_moon WHERE id = action_id;
    IF NOT FOUND THEN
      RAISE EXCEPTION 'Unable to fetch moon id for action %', action_id;
    END IF;

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
    -- that can happen on a moon (at least for now). There
    -- is a need to update the fields for moons though.
    UPDATE moons AS m
      SET fields = fields + cabfem.additional_fields
    FROM
      construction_actions_buildings_fields_effects_moon AS cabfem
      INNER JOIN construction_actions_buildings_moon AS cabm ON cabfe.action = cab.id
    WHERE
      cabfem.action = action_id
      AND m.id = cabm.moon;

    -- 2.e) Update the last activity time for this moon.
    UPDATE moons SET last_activity = moment WHERE id = planet_id;

    -- 2.f) Add the cost of the action to the points of the
    -- player in the economy section.
    UPDATE players_points
      SET economy_points = economy_points + points
    FROM
      construction_actions_buildings_moon AS cabm
      INNER JOIN moons AS m ON cabm.moon = m.id
      INNER JOIN planets AS p ON m.planet = p.id
    WHERE
      cabm.id = action_id
      AND p.player = players_points.player;

      -- 2.g) Add the cost of this building to the total
      -- cost for this building on this moon.
    UPDATE moons_buildings
      SET points = moons_buildings.points + cabm.points
    FROM
      construction_actions_buildings_moon AS cabm
    WHERE
      cabm.id = action_id
      AND moons_buildings.moon = cabm.moon
      AND moons_buildings.building = cabm.element;

    -- 3. Only fields effects can be applied in the case of
    -- moon buildings.
    DELETE FROM construction_actions_buildings_fields_effects_moon WHERE action = action_id;

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

  -- 2.a) Update the last activity time for this moon.
  UPDATE planets AS p
    SET last_activity = cat.completion_time
  FROM
    construction_actions_technologies AS cat
  WHERE
    p.id = cat.planet;

  -- 2.b) Add the cost of the action to the points of the
  -- player in the research section.
  UPDATE players_points
    SET research_points = research_points + points
  FROM
    construction_actions_technologies AS cat
    INNER JOIN planets AS p ON cat.planet = p.id
  WHERE
    cat.id = action_id
    AND p.player = players_points.player;

  WITH points AS (
    SELECT
      SUM(bc.cost)/1000.0 AS sum,
      cat.planet as planet
    FROM
      construction_actions_technologies AS cat
      INNER JOIN buildings AS b ON cat.element = b.id
      INNER JOIN buildings_costs AS bc ON b.id = bc.element
    WHERE
      cat.id = action_id
    GROUP BY
      cat.planet
    )
  UPDATE players_points
    SET research_points = research_points + sum
  FROM
    points AS p
    INNER JOIN planets AS pl ON p.planet = pl.id
  WHERE
    pl.player = players_points.player;

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
          FLOOR(
            EXTRACT(EPOCH FROM processing_time - cas.created_at) / EXTRACT(EPOCH FROM cas.completion_time)
          ),
          CAST(cas.amount AS DOUBLE PRECISION)
        )
      FROM
        construction_actions_ships AS cas
      WHERE
        cas.id = action_id
        AND ps.planet = cas.planet
        AND ps.ship = cas.element;

    -- 1.b) Add the cost of the action to the points of the
    -- player in the military built section. We need to do
    -- that *before* updating the number of remaining elems
    -- to be sure to get the same number of built items and
    -- thus count all the points.
    -- We use the same piece of code to compute how many
    -- items have been built and thus how many points will
    -- be added.
    WITH points AS (
      WITH build AS (
        SELECT
          LEAST(
            FLOOR(
              EXTRACT(EPOCH FROM processing_time - cas.created_at) / EXTRACT(EPOCH FROM cas.completion_time)
            ),
            CAST (cas.amount AS DOUBLE PRECISION)
          ) AS items
        FROM
          construction_actions_ships AS cas
        WHERE
          cas.id = action_id
        )
      SELECT
        sum(sc.cost * items)/1000 AS sum,
        cas.planet AS planet
      FROM
        construction_actions_ships AS cas
        INNER JOIN ships AS s ON cas.element = s.id
        INNER JOIN ships_costs AS sc ON sc.element = s.id
        CROSS JOIN build
      WHERE
        cas.id = action_id
      GROUP BY
        cas.planet
      )
    UPDATE players_points
      SET military_points = military_points + sum,
      military_points_built = military_points_built + sum
    FROM
      points AS p
      INNER JOIN planets AS pl ON p.planet = pl.id
    WHERE
      pl.player = players_points.player;

    -- 2. Update remaining action with an amount decreased by an
    -- amount consistent with the duration elapsed since the creation.
    UPDATE construction_actions_ships
      SET remaining = amount -
        LEAST(
          FLOOR(
            EXTRACT(EPOCH FROM processing_time - created_at) / EXTRACT(EPOCH FROM completion_time)
          ),
          CAST(amount AS DOUBLE PRECISION)
        )
    WHERE
      id = action_id;

    -- 2.a) Update the last activity time for this planet.
    UPDATE planets AS p
      SET last_activity =
        -- This extract the number of items that were
        -- completed capping to the maximum number of
        -- items that were requested.
        LEAST(
          FLOOR(
            EXTRACT(EPOCH FROM processing_time - cas.created_at) / EXTRACT(EPOCH FROM cas.completion_time)
          ),
          CAST(cas.amount AS DOUBLE PRECISION)
        )
        -- From there we can multiply by the duration
        -- of a single completion time to get the
        -- total elapsed duration to build the amount
        -- of ships.
        * cas.completion_time
        -- And add that to the creation time so that
        -- we have the time at which the element was
        -- finished.
        + cas.created_at
    FROM
      construction_actions_ships AS cas
    WHERE
      cas.id = action_id
      AND p.id = cas.planet;

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
          FLOOR(
            EXTRACT(EPOCH FROM processing_time - casm.created_at) / EXTRACT(EPOCH FROM casm.completion_time)
          ),
          CAST(casm.amount AS DOUBLE PRECISION)
        )
      FROM
        construction_actions_ships_moon AS casm
      WHERE
        casm.id = action_id
        AND ms.moon = casm.moon
        AND ms.ship = casm.element;

    -- 1.b) See comment in above section.
    WITH points AS (
      WITH build AS (
        SELECT
          LEAST(
            FLOOR(
              EXTRACT(EPOCH FROM processing_time - casm.created_at) / EXTRACT(EPOCH FROM casm.completion_time)
            ),
            CAST (casm.amount AS DOUBLE PRECISION)
          ) AS items
        FROM
          construction_actions_ships_moon AS casm
        WHERE
          casm.id = action_id
        )
      SELECT
        sum(sc.cost * items)/1000 AS sum,
        casm.moon AS moon
      FROM
        construction_actions_ships_moon AS casm
        INNER JOIN ships AS s ON casm.element = s.id
        INNER JOIN ships_costs AS sc ON sc.element = s.id
        CROSS JOIN build
      WHERE
        casm.id = action_id
      GROUP BY
        casm.moon
      )
    UPDATE players_points
      SET military_points = military_points + sum,
      military_points_built = military_points_built + sum
    FROM
      points AS p
      INNER JOIN moons AS m ON p.moon = m.id
      INNER JOIN planets AS pl ON m.planet = pl.id
    WHERE
      pl.player = players_points.player;

    -- 2. See comment in above section.
    UPDATE construction_actions_ships_moon
      SET remaining = amount -
        LEAST(
          FLOOR(
            EXTRACT(EPOCH FROM processing_time - created_at) / EXTRACT(EPOCH FROM completion_time)
          ),
          CAST(amount AS DOUBLE PRECISION)
        )
    WHERE
      id = action_id;

    -- 2.a) Update the last activity time for this moon.
    UPDATE moons AS m
      SET last_activity =
        -- This extract the number of items that were
        -- completed capping to the maximum number of
        -- items that were requested.
        LEAST(
          FLOOR(
            EXTRACT(EPOCH FROM processing_time - casm.created_at) / EXTRACT(EPOCH FROM casm.completion_time)
          ),
          CAST(casm.amount AS DOUBLE PRECISION)
        )
        -- From there we can multiply by the duration
        -- of a single completion time to get the
        -- total elapsed duration to build the amount
        -- of ships.
        * casm.completion_time
        -- And add that to the creation time so that
        -- we have the time at which the element was
        -- finished.
        + casm.created_at
    FROM
      construction_actions_ships_moon AS casm
    WHERE
      casm.id = action_id
      AND m.id = casm.moon;

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
          FLOOR(
            EXTRACT(EPOCH FROM processing_time - cad.created_at) / EXTRACT(EPOCH FROM cad.completion_time)
          ),
          CAST(cad.amount AS DOUBLE PRECISION)
        )
      FROM
        construction_actions_defenses AS cad
      WHERE
        cad.id = action_id
        AND pd.planet = cad.planet
        AND pd.defense = cad.element;

    -- 1.b) Add the cost of the action to the points of the
    -- player in the military built section. We need to do
    -- that *before* updating the number of remaining elems
    -- to be sure to get the same number of built items and
    -- thus count all the points.
    -- We use the same piece of code to compute how many
    -- items have been built and thus how many points will
    -- be added.
    WITH points AS (
      WITH build AS (
        SELECT
          LEAST(
            FLOOR(
              EXTRACT(EPOCH FROM processing_time - cad.created_at) / EXTRACT(EPOCH FROM cad.completion_time)
            ),
            CAST (cad.amount AS DOUBLE PRECISION)
          ) AS items
        FROM
          construction_actions_defenses AS cad
        WHERE
          cad.id = action_id
        )
      SELECT
        sum(dc.cost * items)/1000 AS sum,
        cad.planet AS planet
      FROM
        construction_actions_defenses AS cad
        INNER JOIN defenses AS d ON cad.element = d.id
        INNER JOIN defenses_costs AS dc ON dc.element = d.id
        CROSS JOIN build
      WHERE
        cad.id = action_id
      GROUP BY
        cad.planet
      )
    UPDATE players_points
      SET military_points = military_points + sum,
      military_points_built = military_points_built + sum
    FROM
      points AS p
      INNER JOIN planets AS pl ON p.planet = pl.id
    WHERE
      pl.player = players_points.player;

    -- 2. Update remaining action with an amount decreased by an
    -- amount consistent with the duration elapsed since the creation.
    UPDATE construction_actions_defenses
      SET remaining = amount -
        LEAST(
          FLOOR(
            EXTRACT(EPOCH FROM processing_time - created_at) / EXTRACT(EPOCH FROM completion_time)
          ),
          CAST(amount AS DOUBLE PRECISION)
        )
    WHERE
      id = action_id;

    -- 2.a) Update the last activity time for this planet.
    -- It should be updated to the last time a defense has
    -- actually been produced.
    UPDATE planets AS p
      SET last_activity =
        -- This extract the number of items that were
        -- completed capping to the maximum number of
        -- items that were requested.
        LEAST(
          FLOOR(
            EXTRACT(EPOCH FROM processing_time - cad.created_at) / EXTRACT(EPOCH FROM cad.completion_time)
          ),
          CAST(cad.amount AS DOUBLE PRECISION)
        )
        -- From there we can multiply by the duration
        -- of a single completion time to get the
        -- total elapsed duration to build the amount
        -- of defense systems.
        * cad.completion_time
        -- And add that to the creation time so that
        -- we have the time at which the element was
        -- finished.
        + cad.created_at
    FROM
      construction_actions_defenses AS cad
    WHERE
      cad.id = action_id
      AND p.id = cad.planet;

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
    UPDATE moons_defenses AS md
      SET count = count - (cadm.amount - cadm.remaining) +
        LEAST(
          FLOOR(
            EXTRACT(EPOCH FROM processing_time - cadm.created_at) / EXTRACT(EPOCH FROM cadm.completion_time)
          ),
          CAST(cadm.amount AS DOUBLE PRECISION)
        )
      FROM
        construction_actions_defenses_moon AS cadm
      WHERE
        cadm.id = action_id
        AND md.moon = cadm.moon
        AND md.defense = cadm.element;

    -- 1.b) See comment in above section.
    WITH points AS (
      WITH build AS (
        SELECT
          LEAST(
            FLOOR(
              EXTRACT(EPOCH FROM processing_time - cadm.created_at) / EXTRACT(EPOCH FROM cadm.completion_time)
            ),
            CAST (cadm.amount AS DOUBLE PRECISION)
          ) AS items
        FROM
          construction_actions_defenses_moon AS cadm
        WHERE
          cadm.id = action_id
        )
      SELECT
        sum(dc.cost * items)/1000 AS sum,
        cadm.moon AS moon
      FROM
        construction_actions_defenses_moon AS cadm
        INNER JOIN defenses AS d ON cadm.element = d.id
        INNER JOIN defenses_costs AS dc ON dc.element = d.id
        CROSS JOIN build
      WHERE
        cadm.id = action_id
      GROUP BY
        cadm.moon
      )
    UPDATE players_points
      SET military_points = military_points + sum,
      military_points_built = military_points_built + sum
    FROM
      points AS p
      INNER JOIN moons AS m ON p.moon = m.id
      INNER JOIN planets AS pl ON m.planet = pl.id
    WHERE
      pl.player = players_points.player;

    -- 2. See comment in above section.
    UPDATE construction_actions_defenses_moon
      SET remaining = amount -
        LEAST(
          FLOOR(
            EXTRACT(EPOCH FROM processing_time - created_at) / EXTRACT(EPOCH FROM completion_time)
          ),
          CAST(amount AS DOUBLE PRECISION)
        )
    WHERE
      id = action_id;

    -- 2.a) Update the last activity time for this moon.
    UPDATE moons AS m
      SET last_activity =
        -- This extract the number of items that were
        -- completed capping to the maximum number of
        -- items that were requested.
        LEAST(
          FLOOR(
            EXTRACT(EPOCH FROM processing_time - cadm.created_at) / EXTRACT(EPOCH FROM cadm.completion_time)
          ),
          CAST(cadm.amount AS DOUBLE PRECISION)
        )
        -- From there we can multiply by the duration
        -- of a single completion time to get the
        -- total elapsed duration to build the amount
        -- of defense systems.
        * cadm.completion_time
        -- And add that to the creation time so that
        -- we have the time at which the element was
        -- finished.
        + cadm.created_at
    FROM
      construction_actions_defenses_moon AS cadm
    WHERE
      cadm.id = action_id
      AND m.id = cadm.moon;

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
