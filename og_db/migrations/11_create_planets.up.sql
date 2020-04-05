
-- Create the table defining planets.
CREATE TABLE planets (
    id uuid NOT NULL DEFAULT uuid_generate_v4(),
    player uuid NOT NULL,
    name text,
    min_temperature integer NOT NULL,
    max_temperature integer NOT NULL,
    fields integer NOT NULL,
    galaxy integer NOT NULL,
    solar_system integer NOT NULL,
    position integer NOT NULL,
    diameter integer NOT NULL,
    created_at timestamp WITH TIME ZONE DEFAULT current_timestamp,
    PRIMARY KEY (id),
    FOREIGN KEY (player) REFERENCES players(id)
);

-- Create the table referencing resources on each planet.
CREATE TABLE planets_resources (
  planet uuid NOT NULL,
  res uuid NOT NULL,
  amount numeric(15, 5) DEFAULT 1.0,
  FOREIGN KEY (planet) REFERENCES planets(id),
  FOREIGN KEY (res) REFERENCES resources(id)
);

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
