
-- Drop the table referencing defenses construction actions and its associated trigger.
DROP TRIGGER update_defenses_action_creation ON construction_actions_defenses;
DROP TABLE construction_actions_defenses;

-- Drop the table referencing ships construction actions and its associated trigger.
DROP TRIGGER update_ships_action_creation ON construction_actions_ships;
DROP TABLE construction_actions_ships;

-- Drop the table referencing technologies construction and its associated trigger.
DROP TRIGGER update_technologies_action_creation ON construction_actions_technologies;
DROP TABLE construction_actions_technologies;

-- Drop the table referencing buildings construction actions and its associated trigger.
DROP TRIGGER update_buildings_action_creation ON construction_actions_buildings;
DROP TABLE construction_actions_buildings;
