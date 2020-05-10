
-- Drop the table referencing defenses construction actions for moons.
DROP TABLE construction_actions_defenses_moon;

-- Drop the table referencing ships construction actions for moons.
DROP TABLE construction_actions_ships_moon;

-- Drop the table referencing buildings construction actions for moons and its trigger.
DROP TRIGGER update_moons_buildings_action_creation ON construction_actions_buildings_moon;
DROP TABLE construction_actions_buildings_moon;

-- Drop the table referencing defenses construction actions.
DROP TABLE construction_actions_defenses;

-- Drop the table referencing ships construction actions.
DROP TABLE construction_actions_ships;

-- Drop the table referencing technologies construction and its associated trigger.
DROP TRIGGER update_technologies_action_creation ON construction_actions_technologies;
DROP TABLE construction_actions_technologies;

-- Drop the table registering effects of a building upgrade action on the storage.
DROP TABLE construction_actions_buildings_storage_effects;

-- Drop the table registering effects of a building upgrade action on the production.
DROP TABLE construction_actions_buildings_production_effects;

-- Drop the table referencing buildings construction actions and its associated trigger.
DROP TRIGGER update_buildings_action_creation ON construction_actions_buildings;
DROP TABLE construction_actions_buildings;
