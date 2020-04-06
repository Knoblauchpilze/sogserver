
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

-- Import planet into the corresponding table.
CREATE OR REPLACE FUNCTION create_planet(planet json, resources json) RETURNS VOID AS $$
BEGIN
  -- Insert the planet in the planets table.
  INSERT INTO planets SELECT * FROM json_populate_record(null::planets, planet);

  -- Insert the base resources of the planet.
  INSERT INTO planets_resources select * FROM json_populate_recordset(null::planets_resources, resources);
END
$$ LANGUAGE plpgsql;

-- Import building upgrade action in the dedicated table.
CREATE OR REPLACE FUNCTION create_building_upgrade_action(upgrade json) RETURNS VOID AS $$
BEGIN
  INSERT INTO construction_actions_buildings SELECT * FROM json_populate_record(null::construction_actions_buildings, upgrade);
END
$$ LANGUAGE plpgsql;

-- Import technology upgrade action in the dedicated table.
CREATE OR REPLACE FUNCTION create_technology_upgrade_action(upgrade json) RETURNS VOID AS $$
BEGIN
  INSERT INTO construction_actions_technologies SELECT * FROM json_populate_record(null::construction_actions_technologies, upgrade);
END
$$ LANGUAGE plpgsql;

-- Import ship upgrade action in the dedicated table.
CREATE OR REPLACE FUNCTION create_ship_upgrade_action(upgrade json) RETURNS VOID AS $$
BEGIN
  INSERT INTO construction_actions_ships SELECT * FROM json_populate_record(null::construction_actions_ships, upgrade);
END
$$ LANGUAGE plpgsql;

-- Import defense upgrade action in the dedicated table.
CREATE OR REPLACE FUNCTION create_defense_upgrade_action(upgrade json) RETURNS VOID AS $$
BEGIN
  INSERT INTO construction_actions_defenses SELECT * FROM json_populate_record(null::construction_actions_defenses, upgrade);
END
$$ LANGUAGE plpgsql;

-- Import fleets in the relevant table.
CREATE OR REPLACE FUNCTION create_fleet(inputs json) RETURNS VOID AS $$
BEGIN
  INSERT INTO fleets SELECT * FROM json_populate_record(null::fleets, inputs);
END
$$ LANGUAGE plpgsql;

-- Import fleet components in the relevant table.
CREATE OR REPLACE FUNCTION create_fleet_component(component json, ships json) RETURNS VOID AS $$
BEGIN
  -- Insert the fleet element.
  INSERT INTO fleet_elements SELECT * FROM json_populate_record(null::fleet_elements, component);

  -- Insert the ships for this fleet element.
  INSERT INTO fleet_ships SELECT * FROM json_populate_recordset(null::fleet_ships, ships);
END
$$ LANGUAGE plpgsql;
