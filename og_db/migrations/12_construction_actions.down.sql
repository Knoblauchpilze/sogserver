
-- Drop the table referencing defenses construction actions and
-- the associated trigger.
DROP TRIGGER update_defense_action_creation_time ON construction_actions_defenses;
DROP TABLE construction_actions_defenses;

-- Drop the table referencing ships construction actions and the
-- associated trigger.
DROP TRIGGER update_ship_action_creation_time ON construction_actions_ships;
DROP TABLE construction_actions_ships;

-- Drop the table referencing technologies construction actions
-- and the associated trigger.
DROP TRIGGER update_technology_action_creation_time ON construction_actions_technologies;
DROP TABLE construction_actions_technologies;

-- Drop the table referencing buildings construction actions and
-- the associated trigger.
DROP TRIGGER update_building_action_creation_time ON construction_actions_buildings;
DROP TABLE construction_actions_buildings;
