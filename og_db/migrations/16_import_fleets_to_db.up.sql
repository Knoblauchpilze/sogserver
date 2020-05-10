
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
  INSERT INTO fleet_ships
    SELECT
      uuid_generate_v4() AS id,
      (fleet->>'id')::uuid AS fleet,
      t.ship AS ship,
      t.count AS count
    FROM
      json_to_recordset(ships) AS t(ship uuid, count integer);

  -- Insert the resources for this fleet element.
  INSERT INTO fleet_resources
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
  -- Register this action in the actions system.
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
CREATE OR REPLACE FUNCTION fleet_deposit_resources(fleet_id uuid) RETURNS BOOLEAN AS $$
DECLARE
  target_id uuid;
  target_kind text;
BEGIN
  -- Retrieve the ID of the target associated to the
  -- fleet.
  SELECT target INTO target_id FROM fleets WHERE id = fleet_id AND target IS NOT NULL;

  -- If the target does not exist for this fleet, do not
  -- deposit resources. Whether it is an issue will be
  -- determined by the calling script.
  IF NOT FOUND THEN
    RETURN FALSE;
  END IF;

  -- Fetch the target type: if the target type does not
  -- exist it means that the fleet's identifier is not
  -- valid.
  SELECT target_type INTO target_kind FROM fleets WHERE id = fleet_id;

  IF NOT FOUND THEN
    RETURN FALSE;
  END IF;

  -- Perform the update of the resources on the planet
  -- so as to be sure that the player gets the max of
  -- its production in case the new deposit brings the
  -- total over the storage capacity.
  -- This is only relevant in case the target is indeed
  -- a planet.
  IF target_kind = 'planet' THEN
    PERFORM update_resources_for_planet(target_id);
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
      fleet_resources fr
      INNER JOIN fleets f ON fr.fleet = f.id
    WHERE
      f.id = fleet_id
      AND pr.res = fr.resource
      AND pr.planet = target_id;

    -- Remove the resources carried by this fleet.
    DELETE FROM
      fleet_resources AS fr
      USING
        fleets AS f
    WHERE
      fr.fleet = f.id
      AND f.id = fleet_id;
  END IF;

  IF target_kind = 'moon' THEN
    UPDATE moons_resources AS mr
      SET amount = mr.amount + fr.amount
    FROM
      fleet_resources fr
      INNER JOIN fleets f ON fr.fleet = f.id
    WHERE
      f.id = fleet_id
      AND mr.res = fr.resource
      AND mr.moon = target_id;

    -- Remove the resources carried by this fleet.
    DELETE FROM
      fleet_resources AS fr
      USING
        fleets AS f
    WHERE
      fr.fleet_element = f.id
      AND f.id = fleet_id;
  END IF;

  RETURN TRUE;
END
$$ LANGUAGE plpgsql;

-- Perform updates to account for a transport fleet.
CREATE OR REPLACE FUNCTION fleet_transport(fleet_id uuid) RETURNS VOID AS $$
DECLARE
  target_was_found BOOLEAN;
BEGIN
  -- Use the dedicated script to perform the deposit
  -- of the resources.
  SELECT fleet_deposit_resources(fleet_id) INTO target_was_found;

  -- Raise an error in case the target associated to
  -- the fleet does not exist.
  IF NOT target_was_found THEN
    RAISE EXCEPTION 'Fleet % is not directed towards a valid target', fleet_id;
  END IF;
END
$$ LANGUAGE plpgsql;

-- Perform updates to account for a deployment fleet.
CREATE OR REPLACE FUNCTION fleet_deployment(fleet_id uuid) RETURNS VOID AS $$
DECLARE
  target_was_found BOOLEAN;
  target_id uuid;
  target_kind text;
BEGIN
  -- Make the resources carried by the fleet drop on
  -- the target element.
  SELECT fleet_deposit_resources(fleet_id) INTO target_was_found;

  -- Raise an error in case the target associated to
  -- the fleet does not exist.
  IF NOT target_was_found THEN
    RAISE EXCEPTION 'Fleet % is not directed towards a valid target', fleet_id;
  END IF;

  -- At this point we know that the fleet exists and
  -- that its target also exists so we can fetch it.
  SELECT target INTO target_id FROM fleets WHERE id = fleet_id AND target IS NOT NULL;
  SELECT target_type INTO target_kind FROM fleets WHERE id = fleet_id;

  -- Assign the ships to the target: either the planet
  -- or the moon based on the kind of the target.
  IF target_kind = 'planet' THEN
    UPDATE planets_ships AS ps
      SET count = ps.count + fs.count
    FROM
      fleet_ships fs
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
      fleet_ships fs
      INNER JOIN fleets f ON fs.fleet = f.id
    WHERE
      f.id = fleet_id
      AND ms.ship = fs.ship
      AND ms.moon = target_id;
  END IF;

  -- Remove the ships associated to this fleet.
  DELETE FROM
    fleet_ships AS fs
    USING
      fleets AS f
  WHERE
    fs.fleet = f.id
    AND f.id = fleet_id;

  -- Remove this fleet from any ACS operation.
  DELETE FROM
    fleets_acs_components
  WHERE
    fleet = fleet_id;

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

  -- And finally remove the fleet which is now as
  -- empty as my bank account.
  DELETE FROM
    fleets
  WHERE
    id = fleet_id;
END
$$ LANGUAGE plpgsql;

-- Perform updates to account for a harvesting fleet.
CREATE OR REPLACE FUNCTION fleet_harvesting(fleet_id uuid) RETURNS VOID AS $$
BEGIN
  -- We need to update the resources carried by the
  -- fleet with the content of the debris fields.
  -- It is required to make sure that we don't use
  -- more cargo space than available on the fleet.
  -- We also need to take care of both resources
  -- that are existing as cargo in the fleet and
  -- insert the resources that don't exist.
  UPDATE fleet_resources AS fr
    SET amount = fr.amount + dfr.amount
  FROM
    debris_fields_resources dfr
    INNER JOIN debris_fields df ON dfr.field=df.id
    INNER JOIN fleets f ON (
      df.universe = f.uni
      AND df.galaxy = f.target_galaxy
      AND df.solar_system = f.target_solar_system
      AND df.position = f.target_position
      AND f.target_type = 'debris'
    )
  WHERE
    f.id = fleet_id;

  -- Handle resources that are not yet part of the
  -- cargo for this fleet.
  INSERT INTO fleet_resources
  SELECT
    -- TODO: This won't work. As the resources are registered per fleet element
    -- we need to find a way to either move this on a per fleet (and make the
    -- split later, probably when the components arrive) or already divide the
    -- total amount of the debris fields based on fleet components.
    fleet_element_id ???,
    dfr.res AS resource,
    dfr.amount AS amount
  FROM
    debris_fields_resources dfr
    INNER JOIN debris_fields df ON dfr.field = df.id
    INNER JOIN fleets f ON (
      df.universe = f.uni
      AND df.galaxy = f.target_galaxy
      AND df.solar_system = f.target_solar_system
      AND df.position = f.target_position
      AND f.target_type = 'debris'
    )
  WHERE
    f.id = fleet_id;

  -- Remove resources from the debris field.
  -- TODO: Handle this.
  -- TODO: Handle the cargo.

  -- Remove the empty lines in resources table.
  DELETE FROM debris_fields_resources WHERE amount <= 0.0;

  -- Remove the empty debris fields.
  DELETE FROM
    debris_fields
  WHERE
    id NOT IN (
      SELECT
        field
      FROM
        debris_fields_resources
      GROUP BY
        field
      HAVING
        count(*) > 0
    );
END
$$ LANGUAGE plpgsql;