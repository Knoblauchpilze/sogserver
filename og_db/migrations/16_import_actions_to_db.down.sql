
-- Drop the function allowing to update a specific defense action.
DROP FUNCTION update_defense_upgrade_action(action_id uuid, kind text);

-- Drop the function allowing to update a specific ship action.
DROP FUNCTION update_ship_upgrade_action(action_id uuid, kind text);

-- Drop the function allowing to update a specific technology action.
DROP FUNCTION update_technology_upgrade_action(action_id uuid);

-- Drop the function allowing to update a specific building action.
DROP FUNCTION update_building_upgrade_action(action_id uuid, kind text);

-- Drop the function allowing to update the resources of a planet to the current time.
DROP FUNCTION update_resources_for_planet(planet_id uuid);

-- Drop the function to update resources on a given planet.
DROP FUNCTION update_resources_for_planet_to_time(planet_id uuid, moment TIMESTAMP WITH TIME ZONE);

-- Drop the defense upgrade insertion script.
DROP FUNCTION create_defense_upgrade_action(upgrade json, costs json, kind text);

-- Drop the ship upgrade insertion script.
DROP FUNCTION create_ship_upgrade_action(upgrade json, costs json, kind text);

-- Drop the technology upgrade insertion script.
DROP FUNCTION create_technology_upgrade_action(upgrade json, costs json);

-- Drop the building upgrade insertion script.
DROP FUNCTION create_building_upgrade_action(upgrade json, costs json, production_effects json, storage_effects json, fields_effects json, kind text);
