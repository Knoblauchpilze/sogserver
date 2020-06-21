
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

-- Import fleet in the relevant table.
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
      fleets AS f
    WHERE
      f.id = (fleet->>'id')::uuid;
END
$$ LANGUAGE plpgsql;

-- Import ACS operation and perform the needed operations
-- to create the associated component.
CREATE OR REPLACE FUNCTION create_acs_fleet(acs_id uuid, fleet json, ships json, resources json, consumption json) RETURNS VOID AS $$
DECLARE
  acs uuid;
BEGIN
  -- Make sure that the target and source type for this fleet are valid.
  IF fleet->>'target_type' != 'planet' AND fleet->>'target_type' != 'moon' THEN
    RAISE EXCEPTION 'Invalid kind % specified for target of ACS fleet', fleet->>'target_type';
  END IF;

  IF fleet->>'source_type' != 'planet' AND fleet->>'source_type' != 'moon' THEN
    RAISE EXCEPTION 'Invalid kind % specified for source of ACS fleet', fleet->>'source_type';
  END IF;

  -- Perform the creation of the fleet as for a regular fleet.
  PERFORM create_fleet(fleet, ships, resources, consumption);

  -- Register the ACS action if needed.
  SELECT id INTO acs FROM fleets_acs WHERE id = acs_id;
  IF NOT FOUND THEN
    -- The ACS operation does not exist yet, create it from the
    -- fleet's data.
    INSERT INTO fleets_acs("id", "universe", "objective", "target", "target_type")
      VALUES(
        acs_id,
        (fleet->>'universe')::uuid,
        (fleet->>'objective')::uuid,
        (fleet->>'target')::uuid,
        (fleet->>'target_type')::text
      );

    -- Insert the acs as a new action to process in the queue.
    INSERT INTO actions_queue("action", "completion_time", "type")
      SELECT
        acs_id,
        f.arrival_time,
        'acs_fleet'
      FROM
        fleets f
      WHERE
        f.id = (fleet->>'id')::uuid;
  ELSE
    -- We need to update the completion time for the ACS
    -- in the actions queue with the arrival time of this
    -- new fleet.
    UPDATE actions_queue
      SET completion_time = (fleet->>'arrival_time')::timestamp with time zone
    WHERE
      action = acs_id;
  END IF;

  -- Register this fleet as one of the component for the ACS.
  INSERT INTO fleets_acs_components("acs", "fleet", "joined_at")
    SELECT
      acs_id,
      f.id,
      f.created_at
    FROM
      fleets AS f
    WHERE
      f.id = (fleet->>'id')::uuid;

  -- Delete this fleet from the actions queue. Indeed we
  -- handled both the case where the ACS did not exist
  -- yet by creating the action and when the ACS fleet
  -- already existed by updating the arrival time to the
  -- arrival time of the fleet.
  DELETE FROM actions_queue WHERE action = (fleet->>'id')::uuid;

  -- Update the arrival time and return time of existing
  -- fleets of the ACS so that it is consistent with the
  -- new component. The only case where this action has
  -- an effect is when the new fleet is slower than the
  -- ACS: otherwise the `arrival_time` set for the fleet
  -- is the one of the existing ACS components.
  UPDATE fleets AS f
    SET
      return_time = f.return_time + ((fleet->>'arrival_time')::timestamp with time zone - f.arrival_time),
      arrival_time = (fleet->>'arrival_time')::timestamp with time zone
    WHERE
      f.id IN (
        SELECT
          fac.fleet
        FROM
          fleets_acs_components AS fac
        WHERE
          fac.acs = acs_id
      );
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

    PERFORM update_resources_for_planet_to_time(target_id, arrival);
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
  arrival_date timestamp with time zone;
  coordinates text;
BEGIN
  -- Retrieve the arrival time of the fleet: it will be
  -- used as a reference for when the resources of the
  -- planet have been updated.
  SELECT arrival_time INTO arrival_date FROM fleets WHERE id = fleet_id;
  IF NOT FOUND THEN
    RAISE EXCEPTION 'Invalid arrival time for fleet % in colonization operation', fleet_id;
  END IF;

  -- Create the planet as provided in input.
  PERFORM create_planet(planet, resources, arrival_date);

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
  recyclers_count integer;
  recyclers_capacity integer;
  coordinates text;
BEGIN
  -- Attempt to retrieve the player as it will be
  -- needed afterwards anyways.
  SELECT player INTO player_id FROM fleets WHERE id = fleet_id;
  IF NOT FOUND THEN
    RAISE EXCEPTION 'Invalid player for fleet % in harvesting operation', fleet_id;
  END IF;

  -- Retrieve the information needed for the
  -- harvesting report.
  SELECT
    fs.count INTO recyclers_count
  FROM
    fleets_ships AS fs
    INNER JOIN ships AS s ON fs.ship = s.id
  WHERE
    fs.fleet = fleet_id
    AND s.name = 'recycler';

  IF NOT FOUND THEN
    RAISE EXCEPTION 'Invalid recyclers count for fleet % in harvesting operation', fleet_id;
  END IF;

  SELECT
    s.cargo * fs.count INTO recyclers_capacity
  FROM
    fleets_ships AS fs
    INNER JOIN ships AS s ON fs.ship = s.id
  WHERE
    fs.fleet = fleet_id
    AND s.name = 'recycler';

  IF NOT FOUND THEN
    RAISE EXCEPTION 'Invalid recyclers capacity for fleet % in harvesting operation', fleet_id;
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
  PERFORM create_message_for(player_id, 'harvesting_report', recyclers_count, recyclers_capacity, coordinates, dispersed, gathered);
END
$$ LANGUAGE plpgsql;

-- Handle the destruction operation of a moon by a fleet
-- of deathstars. We will both perform the deletion of a
-- moon (which includes removing the moon from the list
-- of bodies registered but also rerouting the fleets to
-- the parent planet for example).
CREATE OR REPLACE FUNCTION fleet_destroy(fleet_id uuid, moon_destroyed boolean, fleet_destroyed boolean) RETURNS VOID AS $$
DECLARE
  player_id uuid;
  moon_id uuid;
  deathstars_count integer;
  coordinates text;
BEGIN
  -- Attempt to retrieve the player as it will be needed
  -- afterwards anyways. We will also retrieve the target
  -- of the fleet.
  SELECT player INTO player_id FROM fleets WHERE id = fleet_id;
  IF NOT FOUND THEN
    RAISE EXCEPTION 'Invalid player for fleet % in destruction operation', fleet_id;
  END IF;

  SELECT
    m.id INTO moon_id
  FROM
    fleets AS f
    INNER JOIN moons AS m ON m.id = f.target
  WHERE
    id = fleet_id;

  IF NOT FOUND THEN
    RAISE EXCEPTION 'Invalid moon for fleet % in destruction operation', fleet_id;
  END IF;

  -- Retrieve the information needed for the
  -- destruction report.
  SELECT
    fs.count INTO deathstars_count
  FROM
    fleets_ships AS fs
    INNER JOIN ships AS s ON fs.ship = s.id
  WHERE
    fs.fleet = fleet_id
    AND s.name = 'deathstar';

  IF NOT FOUND THEN
    RAISE EXCEPTION 'Invalid deathstars count for fleet % in destruction operation', fleet_id;
  END IF;

  SELECT
    concat_ws(':', target_galaxy, target_solar_system, target_position)
  INTO
    coordinates
  FROM
    fleets
  WHERE
    id = fleet_id;

  -- In case the fleet is destroyed, we need to remove
  -- any deathstar from the attacking fleet.
  IF fleet_destroyed THEN
    DELETE FROM fleets_ships AS fs
      USING ships AS s 
    WHERE
      fs.ships = s.id
      AND s.name = 'deathstar';

    -- Delete the fleet in case it does not contain any
    -- ship anymore. Note that we will use the fact that
    -- no ACS operation can be used to destroy a moon so
    -- we assume that the fleet will not be existing in
    -- the ACS tables.
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
  END IF;

  -- In case the moon is destroyed we need to remove the
  -- moon and reroute any fleet to the parent planet.
  IF moon_destroyed THEN
    PERFORM delete_moon(moon_id);
  END IF;

  -- We need to register a new message indicating whether
  -- the fleet/moon were destroyed.
  IF moon_destroyed AND fleet_destroyed THEN
    PERFORM create_message_for(player_id, 'destruction_report_all_destroyed', deathstars_count, coordinates);
  ELSEIF moon_destroyed THEN
    PERFORM create_message_for(player_id, 'destruction_report_moon_destroyed', deathstars_count, coordinates);
  ELSEIF fleet_destroyed THEN
    PERFORM create_message_for(player_id, 'destruction_report_fleet_destroyed', deathstars_count, coordinates);
  ELSE
    PERFORM create_message_for(player_id, 'destruction_report_failed', deathstars_count, coordinates);
  END IF;
END
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION planet_fight_aftermath(target_id uuid, kind text, planet_ships json, planet_defenses json, debris json, moon boolean, diameter integer) RETURNS VOID AS $$
DECLARE
  field_id uuid;
  target_galaxy integer;
  target_system integer;
  target_position integer;
  universe_id uuid;
  moon_id uuid;
BEGIN
  -- Update the ships and defenses for the target planet
  -- or moon: we assume that the input arguments are the
  -- total final count of each system/ship. We will also
  -- handle the creation of the moon in the case of a
  -- planet and if a moon does not already exist.
  IF kind = 'planet' THEN
    WITH rs AS (
      SELECT
        t.ship,
        t.count
      FROM
        json_to_recordset(planet_ships) AS t(ship uuid, count integer)
      )
    UPDATE planets_ships AS ps
      SET count = rs.count
    FROM
      rs
    WHERE
      ps.planet = target_id
      AND ps.ship = rs.ship;

    WITH rs AS (
      SELECT
        t.defense,
        t.count
      FROM
        json_to_recordset(planet_defenses) AS t(defense uuid, count integer)
      )
    UPDATE planets_defenses AS pd
      SET count = rs.count
    FROM
      rs
    WHERE
      pd.planet = target_id
      AND pd.defense = rs.defense;

    -- Attempt to create a moon if needed.
    IF moon THEN
      -- Check whether a moon already exist for the
      -- parent planet: if it is the case we won't
      -- create a new one.
      SELECT id INTO moon_id FROM moons WHERE planet = target_id;

      IF NOT FOUND THEN
        -- The moon does not exist yet, create it.
        moon_id = uuid_generate_v4();
        PERFORM create_moon(moon_id, target_id, diameter);
      END IF;
    END IF;
  END IF;

  IF kind = 'moon' THEN
    WITH rs AS (
      SELECT
        t.ship,
        t.count
      FROM
        json_to_recordset(planet_ships) AS t(ship uuid, count integer)
      )
    UPDATE moons_ships AS ms
      SET count = rs.count
    FROM
      rs
    WHERE
      ms.moon = target_id
      AND ms.ship = rs.ship;

    WITH rs AS (
      SELECT
        t.defense,
        t.count
      FROM
        json_to_recordset(planet_defenses) AS t(defense uuid, count integer)
      )
    UPDATE moons_defenses AS md
      SET count = rs.count
    FROM
      rs
    WHERE
      md.planet = target_id
      AND md.defense = rs.defense;
  END IF;

  -- Create the debris field and insert the resources in
  -- it: we will perform an addition of the resources if
  -- the field already exists.
  SELECT
    p.galaxy,
    p.solar_system,
    p.position,
    pl.universe
  INTO
    target_galaxy,
    target_system,
    target_position,
    universe_id
  FROM
    planets AS p
    INNER JOIN players AS pl ON p.player = pl.id
  WHERE
    p.id = target_id;

  IF NOT FOUND THEN
    RAISE EXCEPTION 'Invalid target coordinates for fleet % in fleet fight operation', fleet_id;
  END IF;

  SELECT
    id
  INTO
    field_id
  FROM
    debris_fields
  WHERE
    universe = universe_id
    AND galaxy = target_galaxy
    AND solar_system = target_system
    AND position = target_position;

  IF NOT FOUND THEN
    -- The debris field does not exist yet, create it.
    -- Note that as we create entries for all the res
    -- that might be dispersed in a debris field we
    -- don't have to worry about a resource not being
    -- present later on (i.e. when debris are added to
    -- the field).
    field_id = uuid_generate_v4();

    INSERT INTO debris_fields("id", "universe", "galaxy", "solar_system", "position")
      VALUES(field_id, universe_id, target_galaxy, target_system, target_position);

    INSERT INTO debris_fields_resources("field", "res", "amount")
      SELECT
        field_id,
        r.id,
        0.0
      FROM
        resources AS r
      WHERE
        r.dispersable = 'true';
  END IF;

  WITH dr AS (
    SELECT
      t.resource,
      t.amount
    FROM
      json_to_recordset(debris) AS t(resource uuid, amount numeric(15, 5))
    )
  UPDATE debris_fields_resources AS dfr
    SET amount = dr.amount + dfr.amount
  FROM
    dr
  WHERE
    dfr.field = field_id
    AND dfr.res = dr.resource;
END
$$ LANGUAGE plpgsql;

-- Handle the aftermath of a fleet fight on a planet. We
-- have to update the resources pillaged by the fleet,
-- create the debris field if needed (even if it does not
-- contain any resources) and remove any destroyed ships
-- and defenses from the fleet and the planet.
CREATE OR REPLACE FUNCTION fleet_fight_aftermath(fleet_id uuid, ships json, pillage json, outcome text) RETURNS VOID AS $$
DECLARE
  check_fleet_id uuid;
BEGIN
  -- Update the resources carried by the fleet with the
  -- input values. The `pillage` actually describes all
  -- the resources carried by the fleet.
  DELETE FROM fleets_resources WHERE fleet = fleet_id;

  INSERT INTO fleets_resources("fleet", "resource", "amount")
    SELECT
      fleet_id,
      t.resource,
      t.amount
    FROM
      json_to_recordset(pillage) AS t(resource uuid, amount numeric(15, 5))
    WHERE
      t.amount > 0;

  WITH rs AS (
    SELECT
      t.ship,
      t.count
    FROM
      json_to_recordset(ships) AS t(ship uuid, count integer)
    )
  UPDATE fleets_ships AS fs
    SET count = rs.count
  FROM
    rs
  WHERE
    fs.fleet = fleet_id
    AND fs.ship = rs.ship;

  -- Update the fleet's data: the input `fleet` should
  -- contain the description of remaining ships.
  WITH rs AS (
    SELECT
      t.ship,
      t.count
    FROM
      json_to_recordset(ships) AS t(ship uuid, count integer)
    )
  UPDATE fleets_ships AS fs
    SET count = rs.count
  FROM
    rs
  WHERE
    fs.fleet = fleet_id
    AND fs.ship = rs.ship;

  -- Handle the creation of the fight report for the
  -- attacking player before deleting the fleet. We
  -- don't know for sure that this will happen but
  -- we handle this now to be on the safe side.
  PERFORM fleet_fight_report(fleet_id);

  -- Delete empty entries in the `fleets_ships` table.
  DELETE FROM fleets_ships WHERE fleet = fleet_id AND count = 0;

  -- Delete the fleet in case it does not contain any
  -- ship anymore. We won't consider the ACS case in
  -- here because the process is to first update the
  -- individual fleets of an ACS after a fight and to
  -- then update the ACS where it is easy to determine
  -- whether there are still some components to the
  -- ACS.
  -- So the removal from the `actions_queue` table is
  -- only working when the fleet is not an ACS comp
  -- but it does not hurt (except performance a bit)
  -- to perform the removal.
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

  -- After a fight has been processed we can trigger
  -- the update of the return process. This will also
  -- play nicely in case the fleet is deployed to an
  -- allied planet. We need to handle this only if
  -- the fleet has not been completely wiped out by
  -- the fight.
  -- To determine that we will attempt to select the
  -- fleet's identifier into a local variable to be
  -- certain that the fleet still exists.
  SELECT id INTO check_fleet_id FROM fleets WHERE id = fleet_id;
  IF FOUND THEN
    PERFORM fleet_return_to_base(fleet_id);
  END IF;
END
$$ LANGUAGE plpgsql;

-- In case an ACS operation has just been performed we
-- need to perform its removal: this will then break
-- the individual fleets as regular elements going back
-- to their home planets.
CREATE OR REPLACE FUNCTION acs_fleet_fight_aftermath(acs_id uuid) RETURNS VOID AS $$
BEGIN
  -- We need to remove all the components of the ACS
  -- fleet and update the corresponding fleets so as
  -- to remove references to the ACS.
  DELETE FROM fleets_acs_components WHERE acs = acs_id;

  DELETE FROM fleets_acs WHERE id = acs_id;
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
  -- Attempt to retrieve the player owning the fleet.
  -- Retrieve the kind of the target: this is the first
  -- step in determining the target player information.
  SELECT source_type INTO source_kind FROM fleets WHERE id = fleet_id;
  IF NOT FOUND THEN
    RAISE EXCEPTION 'Invalid source kind for fleet % in espionage operation', fleet_id;
  END IF;

  SELECT player INTO spyer_id FROM fleets WHERE id = fleet_id;

  SELECT
    name
  INTO
    spyer_name
  FROM
    fleets AS f
    INNER JOIN players AS p on f.player = p.id
  WHERE
    f.id = fleet_id;

  IF source_kind = 'planet' THEN
    SELECT
      p.name
    INTO
      source_name
    FROM
      fleets AS f
      INNER JOIN planets AS p ON f.source = p.id
    WHERE
      f.id = fleet_id;

    SELECT
      concat_ws(':', p.galaxy,  p.solar_system,  p.position)
    INTO
      source_coordinates
    FROM
      fleets AS f
      INNER JOIN planets AS p ON f.source = p.id
    WHERE
      id = fleet_id;
  END IF;

  IF source_kind = 'moon' THEN
    SELECT
      m.name
    INTO
      source_name
    FROM
      fleets AS f
      INNER JOIN moons AS m ON f.source = m.id
    WHERE
      f.id = fleet_id;

    SELECT
      concat_ws(':', p.galaxy,  p.solar_system,  p.position)
    INTO
      source_coordinates
    FROM
      fleets AS f
      INNER JOIN moons AS m ON f.source = m.id
      INNER JOIN planets AS p ON m.planet = p.id
    WHERE
      id = fleet_id;
  END IF;

  -- Now fetch information about the spied player which
  -- is described by the target of the fleet.
  SELECT target_type INTO target_kind FROM fleets WHERE id = fleet_id;
  IF NOT FOUND THEN
    RAISE EXCEPTION 'Invalid target kind for fleet % in espionage operation', fleet_id;
  END IF;

  IF target_kind = 'planet' THEN
    SELECT
      p.name,
      p.player
    INTO
      target_name,
      player_id
    FROM
      fleets AS f
      INNER JOIN planets AS m ON f.target = p.id
    WHERE
      f.id = fleet_id;
  END IF;

  IF target_kind = 'moon' THEN
    SELECT
      m.name,
      p.player
    INTO
      target_name,
      player_id
    FROM
      fleets AS f
      INNER JOIN moons AS m ON f.target = m.id
      INNER JOIN planets AS p ON m.planet = p.id
    WHERE
      f.id = fleet_id;
  END IF;

  SELECT
    concat_ws(':', target_galaxy,  target_solar_system,  target_position)
  INTO
    target_coordinates
  FROM
    fleets
  WHERE
    id = fleet_id;

  -- Register the counter espionage report.
  PERFORM create_message_for(player_id, 'counter_espionage_report', source_name, source_coordinates, spyer_name, target_name, target_coordinates, counter_espionage::text);

  -- Register the espionage report.
  PERFORM create_message_for(spyer_id, 'espionage_report', info_level::text);
END
$$ LANGUAGE plpgsql;