
-- Create the table defining planets.
CREATE TABLE planets (
    id uuid NOT NULL DEFAULT uuid_generate_v4(),
    player uuid NOT NULL references players,
    name text,
    min_temperature numeric(5,2) NOT NULL,
    max_temperature numeric(5,2) NOT NULL,
    fields integer NOT NULL,
    galaxy integer NOT NULL,
    solar_system integer NOT NULL,
    position integer NOT NULL,
    diameter integer NOT NULL,
    created_at timestamp with time zone default current_timestamp,
    PRIMARY KEY (id)
);

-- Trigger to update the `created_at` field of the table.
CREATE TRIGGER update_planet_creation_time BEFORE INSERT ON planets FOR EACH ROW EXECUTE PROCEDURE update_created_at_column();

-- Create the buildings per planet table.
CREATE TABLE planets_buildings (
  planet uuid NOT NULL references planets,
  building uuid NOT NULL references buildings,
  level integer NOT NULL default 0
);

-- Create the table referencing resources on each planet.
CREATE TABLE planets_resources (
  planet uuid NOT NULL references planets,
  res uuid NOT NULL  references resources,
  amount numeric(15, 5)
);

-- Create the table containing the defenses on each planet.
CREATE TABLE planets_defenses (
  planet uuid NOT NULL  references planets,
  defense uuid NOT NULL references defenses,
  count integer NOT NULL
);

-- Create the table containing the ships on each planet.
CREATE TABLE planets_ships (
  planet uuid NOT NULL references planets,
  ship uuid NOT NULL references ships,
  count integer NOT NULL
);
