
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

-- Import planet into the corresponding table.
CREATE OR REPLACE FUNCTION create_planet(planet_data json, resources json) RETURNS VOID AS $$
BEGIN
  -- Insert the planet in the planets table.
  INSERT INTO planets
    SELECT *
    FROM json_populate_record(null::planets, planet_data);

  -- Insert the base resources of the planet.
  INSERT INTO planets_resources(planet, res, amount, production, storage_capacity)
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
