
-- Drop the function allowing to update technology for a player.
DROP FUNCTION update_technology_upgrade_action(player_id uuid);

-- Drop the function to update resources on a given planet.
DROP FUNCTION update_resources_for_planet(resources json);

-- Drop the fleet components import function.
DROP FUNCTION create_fleet_component(component json, ships json);

-- Drop the fleet import function.
DROP FUNCTION create_fleet(inputs json);

-- Drop the defense upgrade insertion script.
DROP FUNCTION create_defense_upgrade_action(upgrade json);

-- Drop the ship upgrade insertion script.
DROP FUNCTION create_ship_upgrade_action(upgrade json);

-- Drop the technology upgrade insertion script.
DROP FUNCTION create_technology_upgrade_action(upgrade json);

-- Drop the building upgrade insertion script.
DROP FUNCTION create_building_upgrade_action(upgrade json);

-- Drop the planet's creation script.
DROP FUNCTION create_planet(planet json, resources json);

-- Drop the player's creation script.
DROP FUNCTION create_player(inputs json);

-- Drop the account's creation script.
DROP FUNCTION create_account(inputs json);

-- Drop the universe's creation script.
DROP FUNCTION create_universe(inputs json);
