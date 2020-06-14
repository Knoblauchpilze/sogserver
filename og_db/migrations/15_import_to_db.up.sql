
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
  INSERT INTO accounts(id, name, mail, password, created_at)
  SELECT *
    FROM json_populate_record(null::accounts, inputs);
END
$$ LANGUAGE plpgsql;

-- Update data for an existing account.
CREATE OR REPLACE FUNCTION update_account(account_id uuid, inputs json) RETURNS VOID AS $$
DECLARE
  acc_name text;
  acc_mail text;
  acc_password text;
BEGIN
  -- Fetch the data from the `inputs` and update only
  -- values that are filled.
  SELECT
    t.name,
    t.mail,
    t.password
  INTO
    acc_name,
    acc_mail,
    acc_password
  FROM
    json_to_record(inputs) AS t(name text, mail text, password text);

  -- Update each prop if it is defined.
  IF acc_name != '' THEN
    UPDATE accounts SET name = acc_name WHERE id = account_id;
  END IF;

  IF acc_mail != '' THEN
    UPDATE accounts SET mail = acc_mail WHERE id = account_id;
  END IF;

  IF acc_password != '' THEN
    UPDATE accounts SET password = acc_password WHERE id = account_id;
  END IF;
END
$$ LANGUAGE plpgsql;

-- Create players from the account and universe data.
CREATE OR REPLACE FUNCTION create_player(inputs json) RETURNS VOID AS $$
BEGIN
  -- Insert the player's data into the dedicated table.
  INSERT INTO players(id, universe, account, name, created_at)
    SELECT *
    FROM json_populate_record(null::players, inputs);

  -- Insert technologies with a `0` level in the table.
  -- The conversion in itself includes retrieving the `json`
  -- key by value and then converting it to a uuid. Here is
  -- a useful link:
  -- https://stackoverflow.com/questions/53567903/postgres-cast-to-uuid-from-json
  INSERT INTO players_technologies(player, technology, level)
    SELECT
      (inputs->>'id')::uuid,
      t.id,
      0
    FROM
      technologies AS t;
END
$$ LANGUAGE plpgsql;

-- Update data for an existing player.
CREATE OR REPLACE FUNCTION update_player(player_id uuid, inputs json) RETURNS VOID AS $$
DECLARE
  p_name text;
BEGIN
  -- Fetch the data from the `inputs` and update only
  -- values that are filled. For now there's only the
  -- name of the player but we could add more later.
  SELECT t.name INTO p_name FROM json_to_record(inputs) AS t(name text);

  -- Update each prop if it is defined.
  IF p_name != '' THEN
    UPDATE players SET name = p_name WHERE id = player_id;
  END IF;
END
$$ LANGUAGE plpgsql;

-- Delete data for an existing player.
CREATE OR REPLACE FUNCTION delete_player(player_id uuid) RETURNS VOID AS $$
BEGIN
  -- TODO: Handle this.
END
$$ LANGUAGE plpgsql;

-- Import planet into the corresponding table.
CREATE OR REPLACE FUNCTION create_planet(planet_data json, resources json, moment TIMESTAMP WITH TIME ZONE) RETURNS VOID AS $$
BEGIN
  -- Insert the planet in the planets table.
  INSERT INTO planets
    SELECT *
    FROM json_populate_record(null::planets, planet_data);

  -- Insert the base resources of the planet.
  INSERT INTO planets_resources(planet, res, amount, production, storage_capacity, updated_at)
    SELECT
      (planet_data->>'id')::uuid,
      res,
      amount,
      production,
      storage_capacity,
      moment
    FROM
      json_populate_recordset(null::planets_resources, resources);

  -- Insert base buildings, ships, defenses on the planet.
  INSERT INTO planets_buildings(planet, building, level)
    SELECT
      (planet_data->>'id')::uuid,
      b.id,
      0
    FROM
      buildings AS b;

  INSERT INTO planets_ships(planet, ship, count)
    SELECT
      (planet_data->>'id')::uuid,
      s.id,
      0
    FROM
      ships AS s;

  INSERT INTO planets_defenses(planet, defense, count)
    SELECT
      (planet_data->>'id')::uuid,
      d.id,
      0
    FROM
      defenses AS d;
END
$$ LANGUAGE plpgsql;

-- Update data for an existing planet.
CREATE OR REPLACE FUNCTION update_planet(planet_id uuid, inputs json) RETURNS VOID AS $$
DECLARE
  p_name text;
BEGIN
  -- Fetch the data from the `inputs` and update only
  -- values that are filled. For now there's only the
  -- name of the planet but we could add more later.
  SELECT t.name INTO p_name FROM json_to_record(inputs) AS t(name text);

  -- Update each prop if it is defined.
  IF p_name != '' THEN
    UPDATE planets SET name = p_name WHERE id = planet_id;
  END IF;
END
$$ LANGUAGE plpgsql;

-- Delete data for an existing planet.
CREATE OR REPLACE FUNCTION delete_planet(planet_id uuid) RETURNS VOID AS $$
BEGIN
  -- TODO: Handle this.
END
$$ LANGUAGE plpgsql;

-- Import moon into the corresponding table.
CREATE OR REPLACE FUNCTION create_moon(moon_id uuid, planet_id uuid, diameter integer) RETURNS VOID AS $$
BEGIN
  -- Insert the moon in the moons table.
  INSERT INTO moons("id", "planet", "name", "fields", "diameter")
    VALUES(moon_id, planet_id, 'moon', 1, diameter);

  -- Insert the base resources of the moon.
  INSERT INTO moons_resources(moon, res, amount)
    SELECT
      moon_id,
      r.id,
      0.0
    FROM
      resources AS r;

  -- Insert base buildings, ships, defenses on the moon.
  INSERT INTO moons_buildings(moon, building, level)
    SELECT
      moon_id,
      b.id,
      0
    FROM
      buildings AS b;

  INSERT INTO moons_ships(moon, ship, count)
    SELECT
      moon_id,
      s.id,
      0
    FROM
      ships AS s;

  INSERT INTO moons_defenses(moon, defense, count)
    SELECT
      moon_id,
      d.id,
      0
    FROM
      defenses AS d;
END
$$ LANGUAGE plpgsql;

-- Update data for an existing moon.
CREATE OR REPLACE FUNCTION update_moon(moon_id uuid, inputs json) RETURNS VOID AS $$
DECLARE
  m_name text;
BEGIN
  -- Fetch the data from the `inputs` and update only
  -- values that are filled. For now there's only the
  -- name of the moon but we could add more later.
  SELECT t.name INTO m_name FROM json_to_record(inputs) AS t(name text);

  -- Update each prop if it is defined.
  IF m_name != '' THEN
    UPDATE moons SET name = m_name WHERE id = moon_id;
  END IF;
END
$$ LANGUAGE plpgsql;

-- Delete a moon from the corresponding table.
CREATE OR REPLACE FUNCTION delete_moon(moon_id uuid) RETURNS VOID AS $$
BEGIN
  -- Delete moon's resources.
  DELETE FROM moons_defenses WHERE moon = moon_id;
  DELETE FROM moons_ships WHERE moon = moon_id;
  DELETE FROM moons_buildings WHERE moon = moon_id;
  DELETE FROM moons_resources WHERE moon = moon_id;

  -- Delete actions queue for this moon.
  DELETE FROM actions_queue AS aq
    USING construction_actions_defenses_moon AS cadm
  WHERE
    aq.action = cadm.id
    AND cadm.moon = moon_id;

  DELETE FROM construction_actions_defenses_moon WHERE moon = moon_id;

  DELETE FROM actions_queue AS aq
    USING construction_actions_ships_moon AS casm
  WHERE
    aq.action = casm.id
    AND casm.moon = moon_id;

  DELETE FROM construction_actions_ships_moon WHERE moon = moon_id;

  DELETE FROM actions_queue AS aq
    USING construction_actions_buildings_moon AS cabm
  WHERE
    aq.action = cabm.id
    AND cabm.moon = moon_id;

  DELETE FROM construction_actions_buildings_moon WHERE moon = moon_id;
  
  -- Reroute fleets to the parent planet when
  -- possible.
  UPDATE fleets
    SET target_type = 'planet'
  FROM
    moons AS m
    INNER JOIN planets AS p ON m.planet = p.id
  WHERE
    f.target_galaxy = p.galaxy
    AND f.target_solar_system = p.solar_system
    AND f.target_position = p.position;

  -- Delete the moon itself.
  DELETE FROM moons WHERE moon = moon_id;
END
$$ LANGUAGE plpgsql;
