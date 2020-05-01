
-- Create the table defining planets.
CREATE TABLE planets (
  id uuid NOT NULL DEFAULT uuid_generate_v4(),
  player uuid NOT NULL,
  name text NOT NULL,
  min_temperature integer NOT NULL,
  max_temperature integer NOT NULL,
  fields integer NOT NULL,
  galaxy integer NOT NULL,
  solar_system integer NOT NULL,
  position integer NOT NULL,
  diameter integer NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  FOREIGN KEY (player) REFERENCES players(id)
);

-- Create the trigger on the table to update the `created_at` field.
CREATE TRIGGER update_planets_creation BEFORE INSERT ON planets FOR EACH ROW EXECUTE PROCEDURE update_created_at();

-- Create the table referencing resources on each planet.
CREATE TABLE planets_resources (
  planet uuid NOT NULL,
  res uuid NOT NULL,
  amount numeric(15, 5) NOT NULL,
  production numeric(15, 5) NOT NULL,
  storage_capacity numeric(15, 5) NOT NULL,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (planet) REFERENCES planets(id),
  FOREIGN KEY (res) REFERENCES resources(id)
);

-- Create the trigger on the table to update the `updated_at` field.
CREATE TRIGGER update_resources_refresh BEFORE INSERT OR UPDATE ON planets_resources FOR EACH ROW EXECUTE PROCEDURE update_updated_at();

-- Create the buildings per planet table.
CREATE TABLE planets_buildings (
  planet uuid NOT NULL,
  building uuid NOT NULL,
  level integer NOT NULL DEFAULT 0,
  FOREIGN KEY (planet) REFERENCES planets(id),
  FOREIGN KEY (building) REFERENCES buildings(id)
);

-- Create the table containing the ships on each planet.
CREATE TABLE planets_ships (
  planet uuid NOT NULL,
  ship uuid NOT NULL,
  count integer NOT NULL,
  FOREIGN KEY (planet) REFERENCES planets(id),
  FOREIGN KEY (ship) REFERENCES ships(id)
);

-- Create the table containing the defenses on each planet.
CREATE TABLE planets_defenses (
  planet uuid NOT NULL,
  defense uuid NOT NULL,
  count integer NOT NULL,
  FOREIGN KEY (planet) REFERENCES planets(id),
  FOREIGN KEY (defense) REFERENCES defenses(id)
);
