
-- Create a function allowing to register a message with
-- the specified type for a given player.
CREATE OR REPLACE FUNCTION create_message_for(player_id uuid, message_name text, VARIADIC args text[]) RETURNS VOID AS $$
DECLARE
  msg_id uuid := uuid_generate_v4();
  pos integer := 0;
  arg text;
BEGIN
  -- Insert the message itself.
  INSERT INTO messages_players(id, player, message)
    SELECT
      msg_id,
      player_id,
      mi.id
    FROM
      messages_ids AS mi
    WHERE
      mi.name = message_name;

  -- And then all its arguments. We need a counter to
  -- determine the position of the arg and preserve
  -- the input order.
  FOREACH arg IN ARRAY args
  LOOP
    INSERT INTO messages_arguments("message", "position", "argument")
      VALUES(msg_id, pos, arg);

    pos := pos + 1;
  END LOOP;
END
$$ LANGUAGE plpgsql;

-- Import fleet components in the relevant table.
CREATE OR REPLACE FUNCTION create_fleet(fleet json, ships json, resources json, consumption json) RETURNS VOID AS $$
BEGIN
  -- Make sure that the target and source type for this fleet are valid.
  IF fleet->>'target_type' != 'planet' AND fleet->>'target_type' != 'moon' AND fleet->>'target_type' != 'debris' THEN
    RAISE EXCEPTION 'Invalid kind % specified for target of fleet', fleet->>'target_type';
  END IF;

  IF fleet->>'source_type' != 'planet' AND fleet->>'source_type' != 'moon' THEN
    RAISE EXCEPTION 'Invalid kind % specified for source of fleet', fleet->>'source_type';
  END IF;

  -- Insert the fleet element.
  INSERT INTO fleets
    SELECT *
    FROM json_populate_record(null::fleets, fleet);

  -- Insert the ships for this fleet element.
  INSERT INTO fleets_ships
    SELECT
      uuid_generate_v4() AS id,
      (fleet->>'id')::uuid AS fleet,
      t.ship AS ship,
      t.count AS count
    FROM
      json_to_recordset(ships) AS t(ship uuid, count integer);

  -- Insert the resources for this fleet element.
  INSERT INTO fleets_resources
    SELECT
      (fleet->>'id')::uuid AS fleet,
      t.resource AS resource,
      t.amount AS amount
    FROM
      json_to_recordset(resources) AS t(resource uuid, amount numeric(15, 5));

  -- Reduce the planet's resources from the amount of the fuel.
  -- Note that depending on the starting location of the fleet
  -- we might have to subtract from the planet or the moon that
  -- is associated to it.
  -- This can be checked using the `source_type` field in the
  -- input `fleet` element.
  IF (fleet->>'source_type') = 'planet' THEN
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
      planet = (fleet->>'source')::uuid
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
      planet = (fleet->>'source')::uuid
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
      planet = (fleet->>'source')::uuid
      AND ship = cs.vessel;
  END IF;

  IF (fleet->>'source_type') = 'moon' THEN
    WITH cc AS (
      SELECT
        t.resource,
        t.amount AS quantity
      FROM
        json_to_recordset(consumption) AS t(resource uuid, amount numeric(15, 5))
      )
    UPDATE moons_resources
      SET amount = amount - cc.quantity
    FROM
      cc
    WHERE
      moon = (fleet->>'source')::uuid
      AND res = cc.resource;

    -- Reduce the planet's resources from the amount that will be moved.
    WITH cr AS (
      SELECT
        t.resource,
        t.amount AS quantity
      FROM
        json_to_recordset(resources) AS t(resource uuid, amount numeric(15, 5))
      )
    UPDATE moons_resources
      SET amount = amount - cr.quantity
    FROM
      cr
    WHERE
      moon = (fleet->>'source')::uuid
      AND res = cr.resource;

    -- Reduce the planet's available ships from the ones that will be launched.
    WITH cs AS (
      SELECT
        t.ship AS vessel,
        t.count AS quantity
      FROM
        json_to_recordset(ships) AS t(ship uuid, count integer)
      )
    UPDATE moons_ships
      SET count = count - cs.quantity
    FROM
      cs
    WHERE
      moon = (fleet->>'source')::uuid
      AND ship = cs.vessel;
  END IF;

  -- Register this fleet as part of the actions system.
  INSERT INTO actions_queue
    SELECT
      f.id AS action,
      f.arrival_time AS completion_time,
      'fleet' AS type
    FROM
      fleets f;
END
$$ LANGUAGE plpgsql;

-- Utility script allowing to deposit the resources
-- carried by a fleet to the target it belongs to.
-- The target can either be a planet or a moon as
-- defined by its objective.
-- The return value indicates whether the target to
-- deposit the resources to (computed with the data
-- from the input fleet) actually existed.
CREATE OR REPLACE FUNCTION fleet_deposit_resources(fleet_id uuid, target_id uuid, target_kind text) RETURNS VOID AS $$
DECLARE
  arrival timestamp with time zone;
BEGIN
  -- Perform the update of the resources on the planet
  -- so as to be sure that the player gets the max of
  -- its production in case the new deposit brings the
  -- total over the storage capacity.
  -- This is only relevant in case the target is indeed
  -- a planet.
  IF target_kind = 'planet' THEN
    SELECT arrival_time INTO arrival FROM fleets WHERE id = fleet_id;

    IF NOT FOUND THEN
      RAISE EXCEPTION 'Unable to fetch arrival time for fleet %', fleet_id;
    END IF;

    PERFORM update_resources_for_planet(target_id, arrival);
  END IF;

  -- Add the resources carried by the fleet to the
  -- destination target and remove them from the
  -- fleet's resources.
  -- The table that will be updated depends on the
  -- type of the target.
  -- Note that as we only update existing resources
  -- for planets (or moons) it means that if the
  -- fleet brings new resources to a planet it will
  -- not be added correctly to the planet's stocks.
  -- This is note a problem for now though as for
  -- now we register all the resources for any new
  -- planet so all possible resources should already
  -- be created beforehand.
  IF target_kind = 'planet' THEN
    UPDATE planets_resources AS pr
      SET amount = pr.amount + fr.amount
    FROM
      fleets_resources AS fr
      INNER JOIN fleets f ON fr.fleet = f.id
    WHERE
      f.id = fleet_id
      AND pr.res = fr.resource
      AND pr.planet = target_id;

    -- Remove the resources carried by this fleet.
    DELETE FROM
      fleets_resources AS fr
      USING fleets AS f
    WHERE
      fr.fleet = f.id
      AND f.id = fleet_id;
  END IF;

  IF target_kind = 'moon' THEN
    UPDATE moons_resources AS mr
      SET amount = mr.amount + fr.amount
    FROM
      fleets_resources AS fr
      INNER JOIN fleets f ON fr.fleet = f.id
    WHERE
      f.id = fleet_id
      AND mr.res = fr.resource
      AND mr.moon = target_id;

    -- Remove the resources carried by this fleet.
    DELETE FROM
      fleets_resources AS fr
      USING fleets AS f
    WHERE
      fr.fleet_element = f.id
      AND f.id = fleet_id;
  END IF;
END
$$ LANGUAGE plpgsql;

-- Performs the registration of the ships of a fleet
-- to the specified target (defined by its identifier
-- and its kind).
CREATE OR REPLACE FUNCTION fleet_ships_deployment(fleet_id uuid, target_id uuid, target_kind text) RETURNS VOID AS $$
BEGIN
  -- Now we can add the ships composing the fleet to the
  -- destination celestial body.
  IF target_kind = 'planet' THEN
    UPDATE planets_ships AS ps
      SET count = ps.count + fs.count
    FROM
      fleets_ships AS fs
      INNER JOIN fleets f ON fs.fleet = f.id
    WHERE
      f.id = fleet_id
      AND ps.ship = fs.ship
      AND ps.planet = target_id;
  END IF;

  IF target_kind = 'moon' THEN
    UPDATE moons_ships AS ms
      SET count = ms.count + fs.count
    FROM
      fleets_ships AS fs
      INNER JOIN fleets f ON fs.fleet = f.id
    WHERE
      f.id = fleet_id
      AND ms.ship = fs.ship
      AND ms.moon = target_id;
  END IF;
END
$$ LANGUAGE plpgsql;

-- Performs the deletion of the fleet from the DB
-- along with its erasing from ACS and other tables.
CREATE OR REPLACE FUNCTIOn fleet_deletion(fleet_id uuid) RETURNS VOID AS $$
BEGIN
  -- Remove the resources carried by the fleet.
  DELETE FROM
    fleets_resources AS fr
    USING fleets AS f
  WHERE
    fr.fleet = f.id
    AND f.id = fleet_id;

  -- Remove the ships associated to this fleet.
  DELETE FROM
    fleets_ships AS fs
    USING fleets AS f
  WHERE
    fs.fleet = f.id
    AND f.id = fleet_id;

  -- Remove this fleet from any ACS operation.
  DELETE FROM fleets_acs_components WHERE fleet = fleet_id;

  -- Remove empty ACS operation.
  DELETE FROM
    fleets_acs
  WHERE
    id NOT IN (
      SELECT
        acs
      FROM
        fleets_acs_components
      GROUP BY
        acs
      HAVING
        count(*) > 0
    );

  -- Remove from the actions' queue.
  DELETE FROM actions_queue WHERE action = fleet_id;

  -- And finally remove the fleet which is now as
  -- empty as my bank account.
  DELETE FROM fleets WHERE id = fleet_id;
END
$$ LANGUAGE plpgsql;

-- Perform the update of the entry related to the fleet
-- in the actions queue to be equal to the return time
-- of the fleet.
CREATE OR REPLACE FUNCTION fleet_update_to_return_time(fleet_id uuid) RETURNS VOID AS $$
DECLARE
  wait_time integer;
BEGIN
  -- Select the wait time for this fleet at its target
  -- destination. This will indicate whether we should
  -- make the fleet return immediately or wait at its
  -- destination.
  SELECT deployment_time INTO wait_time FROM fleets WHERE id = fleet_id;
  IF NOT FOUND THEN
    RAISE EXCEPTION 'Invalid deployment time for fleet % in update return time operation', fleet_id;
  END IF;

  IF wait_time > 0 THEN
    UPDATE actions_queue
      SET completion_time = arrival_time + make_interval(secs := CAST(wait_time AS DOUBLE PRECISION))
    FROM
      fleets AS f
    WHERE
      f.id = fleet_id
      AND action = fleet_id;
  ELSE
    -- Update the corresponding entry in the actions queue.
    UPDATE actions_queue
      SET completion_time = return_time
    FROM
      fleets AS f
    WHERE
      f.id = fleet_id
      AND action = fleet_id;

    -- Indicate that this fleet is now returning to its
    -- source.
    UPDATE fleets SET is_returning = 'true' WHERE id = fleet_id;
  END IF;
END
$$ LANGUAGE plpgsql;

-- Perform the deletion of the fleet and the assignement
-- of the resources carried by it to the source object.
CREATE OR REPLACE FUNCTION fleet_return_to_base(fleet_id uuid) RETURNS VOID AS $$
DECLARE
  processing_time timestamp with time zone = NOW();
  target_id uuid;
  target_kind text;
  arrival_date timestamp with time zone;
  return_date timestamp with time zone;
  deployment_date timestamp with time zone;
  deployment_duration integer;
BEGIN
  SELECT arrival_time INTO arrival_date FROM fleets WHERE id = fleet_id;
  IF NOT FOUND THEN
    RAISE EXCEPTION 'Invalid arrival time for fleet % in return to base operation', fleet_id;
  END IF;
  SELECT return_time INTO return_date FROM fleets WHERE id = fleet_id;
  IF NOT FOUND THEN
    RAISE EXCEPTION 'Invalid return time for fleet % in return to base operation', fleet_id;
  END IF;

  SELECT
    arrival_time + make_interval(secs := CAST(deployment_time AS DOUBLE PRECISION)),
    deployment_time
  INTO
    deployment_date,
    deployment_duration
  FROM
    fleets
  WHERE
    id = fleet_id;
  IF NOT FOUND THEN
    RAISE EXCEPTION 'Invalid deployment time for fleet % in return to base operation', fleet_id;
  END IF;

  -- Update the next activation time for this fleet if it
  -- is consistent with the current time.
  IF arrival_date < processing_time THEN
    PERFORM fleet_update_to_return_time(fleet_id);
  END IF;

  -- Update the next activation time in case the deployment
  -- is not set to `0` and we reached the end of it.
  IF deployment_duration > 0 AND deployment_date < processing_time THEN
    PERFORM fleet_update_to_return_time(fleet_id);
  END IF;

  -- Handle the return of the fleet to its source in case
  -- the processing time indicates so.
  IF return_date < processing_time THEN
    -- Fetch the source's data.
    SELECT source INTO target_id FROM fleets WHERE id = fleet_id;
    IF NOT FOUND THEN
      RAISE EXCEPTION 'Invalid source destination for fleet % in harvesting operation', fleet_id;
    END IF;

    SELECT source_type INTO target_kind FROM fleets WHERE id = fleet_id;
    IF NOT FOUND THEN
      RAISE EXCEPTION 'Invalid source kind for fleet % in harvesting operation', fleet_id;
    END IF;

    -- Deposit the resources that were fetched from the
    -- debris field to the source location.
    PERFORM fleet_deposit_resources(fleet_id, target_id, target_kind);

    -- Restore the ships to the source.
    PERFORM fleet_ships_deployment(fleet_id, target_id, target_kind);

    -- Delete the fleet from the DB.
    PERFORM fleet_deletion(fleet_id);
  END IF;
END
$$ LANGUAGE plpgsql;

-- Perform updates to account for a transport fleet.
CREATE OR REPLACE FUNCTION fleet_transport(fleet_id uuid) RETURNS VOID AS $$
DECLARE
  processing_time timestamp with time zone = NOW();
  target_id uuid;
  target_kind text;
  arrival_date timestamp with time zone;
  return_date timestamp with time zone;
BEGIN
  -- The transport mission has two main events associated
  -- to it: the first one corresponds to when it arrives
  -- to the target and the second one when it returns back
  -- to its source.
  -- In the first case the resources carried by the fleet
  -- should be dumped to the target while on the second
  -- case the ships should be added to the source object
  -- and the fleet destroyed.
  SELECT arrival_time INTO arrival_date FROM fleets WHERE id = fleet_id;
  IF NOT FOUND THEN
    RAISE EXCEPTION 'Invalid arrival time for fleet % in transport operation', fleet_id;
  END IF;
  SELECT return_time INTO return_date FROM fleets WHERE id = fleet_id;
  IF NOT FOUND THEN
    RAISE EXCEPTION 'Invalid return time for fleet % in transport operation', fleet_id;
  END IF;

  -- In case the current time is posterior to the arrival
  -- time, dump the resources to the target element.
  IF arrival_date < processing_time THEN
    -- Retrieve the ID of the target associated to this
    -- fleet along with its type.
    SELECT target INTO target_id FROM fleets WHERE id = fleet_id AND target IS NOT NULL;
    IF NOT FOUND THEN
      RAISE EXCEPTION 'Invalid target destination for fleet % in transport operation', fleet_id;
    END IF;

    SELECT target_type INTO target_kind FROM fleets WHERE id = fleet_id;
    IF NOT FOUND THEN
      RAISE EXCEPTION 'Invalid target kind for fleet % in transport operation', fleet_id;
    END IF;

    PERFORM fleet_deposit_resources(fleet_id, target_id, target_kind);

    -- Update the next time the fleet needs processing
    -- to be the return time.
    PERFORM fleet_update_to_return_time(fleet_id);
  END IF;

  -- Handle the return of the fleet to its source in case
  -- the processing time indicates so.
  IF return_date < processing_time THEN
    -- Fetch the source's data.
    SELECT source INTO target_id FROM fleets WHERE id = fleet_id;
    IF NOT FOUND THEN
      RAISE EXCEPTION 'Invalid source destination for fleet % in transport operation', fleet_id;
    END IF;

    SELECT source_type INTO target_kind FROM fleets WHERE id = fleet_id;
    IF NOT FOUND THEN
      RAISE EXCEPTION 'Invalid source kind for fleet % in transport operation', fleet_id;
    END IF;

    -- Restore the ships to the source.
    PERFORM fleet_ships_deployment(fleet_id, target_id, target_kind);

    -- Delete the fleet from the DB.
    PERFORM fleet_deletion(fleet_id);
  END IF;
END
$$ LANGUAGE plpgsql;

-- Perform updates to account for a deployment fleet.
CREATE OR REPLACE FUNCTION fleet_deployment(fleet_id uuid) RETURNS VOID AS $$
DECLARE
  target_id uuid;
  target_kind text;
BEGIN
  -- Fetch the target of the fleet along with its kind.
  SELECT target INTO target_id FROM fleets WHERE id = fleet_id AND target IS NOT NULL;
  IF NOT FOUND THEN
    RAISE EXCEPTION 'Invalid target destination for fleet % in deploy operation', fleet_id;
  END IF;

  SELECT target_type INTO target_kind FROM fleets WHERE id = fleet_id;
  IF NOT FOUND THEN
    RAISE EXCEPTION 'Invalid target kind for fleet % in deploy operation', fleet_id;
  END IF;

  -- Deposit the resources of the fleet at the target
  -- destination.
  PERFORM fleet_deposit_resources(fleet_id, target_id, target_kind);

  -- Add the ships in the target destination.
  PERFORM fleet_ships_deployment(fleet_id, target_id, target_kind);

  -- Delete the fleet from the DB as its mission is
  -- now complete.
  PERFORM fleet_deletion(fleet_id);
END
$$ LANGUAGE plpgsql;

-- In case a colonization succeeeded, we need to register
-- the new planet along with providing a message to the
-- player explaining the success of the operation.
CREATE OR REPLACE FUNCTION fleet_colonization_success(fleet_id uuid, planet json, resources json) RETURNS VOID AS $$
DECLARE
  player_id uuid;
  coordinates text;
BEGIN
  -- Create the planet as provided in input.
  PERFORM create_planet(planet, resources);

  -- Register the message indicating that the colonization
  -- was sucessful.
  SELECT player INTO player_id FROM fleets WHERE id = fleet_id;
  IF NOT FOUND THEN
    RAISE EXCEPTION 'Invalid player for fleet % in colonization operation', fleet_id;
  END IF;

  -- Create the string representing the coordinates which
  -- are used in the colonization message.
  SELECT
    concat_ws(':', target_galaxy, target_solar_system, target_position)
  INTO
    coordinates
  FROM
    fleets
  WHERE
    id = fleet_id;

  PERFORM create_message_for(player_id, 'colonization_suceeded', coordinates);

  -- Dump the resources transported by the fleet to the
  -- new planet.
  PERFORM fleet_deposit_resources(fleet_id, (planet->>'id')::uuid, 'planet');

  -- Remove one colony ship from the fleet. We know that
  -- there should be at least one.
  UPDATE fleets_ships AS fs
    SET count = count - 1
  FROM
    ships AS s
  WHERE
    s.id = fs.ship
    AND fs.fleet = fleet_id
    AND s.name = 'colony ship';

  -- Delete empty entries in the `fleets_ships` table.
  DELETE FROM
    fleets_ships
  WHERE
    fleet = fleet_id
    AND count = 0;

  -- Delete the fleet in case it does not contain any
  -- ship anymore. Note that we will use the fact that
  -- no ACS operation can be used to colonize a planet
  -- so we assume that the fleet will not be existing
  -- in the ACS tables.
  -- We will also handle first the deletion from the
  -- actions queue before deleting the fleet as it is
  -- the only way have to determine whether the fleet
  -- should be removed.
  DELETE FROM
    actions_queue
  WHERE
    action NOT IN (
      SELECT
        fleet
      FROM
        fleets_ships
      GROUP BY
        fleet
      HAVING
        count(*) > 0
    )
    AND type = 'fleet';

  DELETE FROM
    fleets
  WHERE
    id NOT IN (
      SELECT
        fleet
      FROM
        fleets_ships
      GROUP BY
        fleet
      HAVING
        count(*) > 0
    );
END
$$ LANGUAGE plpgsql;

-- In case a colonization fails, we need to register
-- a new message to the player and make the fleet
-- return to its source.
CREATE OR REPLACE FUNCTION fleet_colonization_failed(fleet_id uuid) RETURNS VOID AS $$
DECLARE
  player_id uuid;
  coordinates text;
BEGIN
  -- We need to register a new message indicating the
  -- coordinate that was not colonizable.
  SELECT player INTO player_id FROM fleets WHERE id = fleet_id;
  IF NOT FOUND THEN
    RAISE EXCEPTION 'Invalid player for fleet % in colonization operation', fleet_id;
  END IF;

  -- Create the string representing the coordinates which
  -- are used in the colonization message.
  SELECT
    concat_ws(':', target_galaxy, target_solar_system, target_position)
  INTO
    coordinates
  FROM
    fleets
  WHERE
    id = fleet_id;

  PERFORM create_message_for(player_id, 'colonization_failed', coordinates);
END
$$ LANGUAGE plpgsql;

-- In case a harvesting mission manages to collect at
-- least a single resource we need to update the data
-- of the field and the cargo carried by the fleet.
CREATE OR REPLACE FUNCTION fleet_harvesting_success(fleet_id uuid, debris_id uuid, resources json, dispersed text, gathered text) RETURNS VOID AS $$
DECLARE
  player_id uuid;
  coordinates text;
BEGIN
  -- Attempt to retrieve the player as it will be
  -- needed afterwards anyways.
  SELECT player INTO player_id FROM fleets WHERE id = fleet_id;
  IF NOT FOUND THEN
    RAISE EXCEPTION 'Invalid player for fleet % in harvesting operation', fleet_id;
  END IF;

  -- Add the resources to the fleet's data. We need
  -- to account both for resources that are already
  -- carried by the fleet and the ones that should
  -- be added.
  WITH rc AS (
    SELECT
      t.resource,
      t.amount
    FROM
      json_to_recordset(resources) AS t(resource uuid, amount numeric(15, 5))
  )
  UPDATE fleets_resources AS fr
    SET amount = fr.amount + rc.amount
  FROM
    rc
  WHERE
    fr.fleet = fleet_id
    AND fr.resource = rc.resource;

  -- Insert the resources for this fleet element.
  INSERT INTO fleets_resources ("fleet", "resource", "amount")
    SELECT
      fleet_id,
      t.resource,
      t.amount
    FROM
      json_to_recordset(resources) AS t(resource uuid, amount numeric(15, 5))
    WHERE
      t.resource NOT IN (
        SELECT
          resource
        FROM
          fleets_resources
        WHERE
          fleet = fleet_id
      );

  -- Remove resources from the debris field.
  WITH rc AS (
    SELECT
      t.resource,
      t.amount
    FROM
      json_to_recordset(resources) AS t(resource uuid, amount numeric(15, 5))
  )
  UPDATE debris_fields_resources AS dfr
    SET amount = dfr.amount - rc.amount
  FROM
    rc
  WHERE
    dfr.field = debris_id
    AND rc.resource = dfr.res;

  -- Delete empty lines in debris field resources.
  DELETE FROM debris_fields_resources
  WHERE
    field = debris_id
    AND amount <= 0.0;

  -- Create the string representing the coordinates which
  -- are used in the harvesting message.
  SELECT
    concat_ws(':', target_galaxy, target_solar_system, target_position)
  INTO
    coordinates
  FROM
    fleets
  WHERE
    id = fleet_id;

  -- We need to register a new message indicating the
  -- resources that were harvested.
  PERFORM create_message_for(player_id, 'harvesting_report', coordinates, dispersed, gathered);
END
$$ LANGUAGE plpgsql;