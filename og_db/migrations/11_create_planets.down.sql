
-- Drop the resources per planet table.
DROP TABLE planets_resources;

-- Drop the planets referencing buildings on planets.
DROP TABLE planets_buildings;

-- Drop the planets table and its associated trigger.
DROP TRIGGER update_planet_creation_time ON planets;
DROP TABLE planets;
