
-- Drop the ships belonging to fleets table.
DROP TRIGGER update_ships_joined_at_time ON fleet_ships;
DROP TABLE fleet_ships;

-- Drop the fleets table.
DROP TRIGGER update_fleet_creation_time ON fleets;
DROP TABLE fleets;

-- Drop the table referencing fleet objectives.
DROP TABLE fleet_objectives;
