
-- Drop the table defining the ships usage.
DROP TABLE ships_usage;

-- Drop the table representing the resources transported by a fleet
DROP TABLE fleet_resources;

-- Drop the ships belonging to fleets table.
DROP TABLE fleet_ships;

-- Drop the table regrouping the participants to a fleet.
DROP TABLE fleet_elements;

-- Drop the fleets table and its associated trigger.
DROP TRIGGER update_fleets_creation ON fleets;
DROP TABLE fleets;

-- Drop the table referencing fleet objectives.
DROP TABLE fleet_objectives;
