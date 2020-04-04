
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
  INSERT INTO players SELECT * FROM json_populate_record(null::players, inputs);
END
$$ LANGUAGE plpgsql;

-- Import planet into the correspoding table.
CREATE OR REPLACE FUNCTION create_planet(planet json, resources json) RETURNS VOID AS $$
BEGIN
  -- Insert the planet in the planets table.
  INSERT INTO planets SELECT * FROM json_populate_record(null::planets, planet);

  -- Insert the base resources of the planet.
  INSERT INTO planets_resources select * FROM json_populate_recordset(null::planets_resources, resources);
END
$$ LANGUAGE plpgsql;
