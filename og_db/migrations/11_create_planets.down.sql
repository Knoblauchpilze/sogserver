
-- Drop the table representing the defenses by moons.
DROP TABLE moons_defenses;

-- Drop the table representing the ships on each moon.
DROP TABLE moons_ships;

-- Drop the planets referencing buildings on moons.
DROP TABLE moons_buildings;

-- Drop the table representing the defenses by planets.
DROP TABLE planets_defenses;

-- Create the table containing the ships on each planet.
DROP TABLE planets_ships;

-- Drop the planets referencing buildings on planets.
DROP TABLE planets_buildings;

-- Drop the resources per debris fields.
DROP TABLE debris_fields_resources;

-- Drop the resources per moons.
DROP TABLE moons_resources;

-- Drop the resources per planet table.
DROP TABLE planets_resources;

-- Drop the debris fields table and its associated trigger.
DROP TRIGGER update_debris_fields_creation ON debris_fields;
DROP TABLE debris_fields;

-- Drop the moons table and its associated trigger.
DROP TRIGGER update_moons_creation ON moons;
DROP TABLE moons;

-- Drop the planets table and its associated trigger.
DROP TRIGGER update_planets_creation ON planets;
DROP TABLE planets;
