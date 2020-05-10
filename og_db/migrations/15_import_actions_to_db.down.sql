
-- Drop the function allowing to update defense for a planet or moon.
DROP FUNCTION update_defense_upgrade_action(target_id uuid, kind text);

-- Drop the function allowing to update ship for a planet or moon.
DROP FUNCTION update_ship_upgrade_action(target_id uuid, kind text);

-- Drop the function allowing to update technology for a player.
DROP FUNCTION update_technology_upgrade_action(player_id uuid);

-- Drop the function allowing to update building for a planet or moon.
DROP FUNCTION update_building_upgrade_action(action_id uuid, kind text);

-- Drop the function to update resources on a given planet.
DROP FUNCTION update_resources_for_planet(planet_id uuid);

-- Drop the defense upgrade insertion script.
DROP FUNCTION create_defense_upgrade_action(upgrade json, costs json, kind text);

-- Drop the ship upgrade insertion script.
DROP FUNCTION create_ship_upgrade_action(upgrade json, costs json, kind text);

-- Drop the technology upgrade insertion script.
DROP FUNCTION create_technology_upgrade_action(upgrade json, costs json);

-- Drop the building upgrade insertion script.
DROP FUNCTION create_building_upgrade_action(upgrade json, costs json, production_effects json, storage_effects json, kind text);
