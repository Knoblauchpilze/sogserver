
-- Drop the table representing the defenses by planets.
DROP TABLE planets_defenses;

-- Create the table containing the ships on each planet.
DROP TABLE planets_ships;

-- Drop the planets referencing buildings on planets.
DROP TABLE planets_buildings;

-- Drop the resources per planet table.
DROP TABLE planets_resources;

-- Drop the planets table and its associated trigger.
DROP TRIGGER update_planet_creation_time ON planets;
DROP TABLE planets;
