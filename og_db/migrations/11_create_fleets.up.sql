
-- Create the table defining the purpose of a fleet.
CREATE TABLE fleet_objectives (
  id uuid NOT NULL DEFAULT uuid_generate_v4(),
  name text NOT NULL,
  hostile boolean NOT NULL,
  directed boolean NOT NULL,
  PRIMARY KEY (id),
  UNIQUE (name)
);

-- Create the table defining fleets.
CREATE TABLE fleets (
  id uuid NOT NULL DEFAULT uuid_generate_v4(),
  name text,
  uni uuid NOT NULL,
  objective uuid NOT NULL,
  target_galaxy integer NOT NULL,
  target_solar_system integer NOT NULL,
  target_position integer NOT NULL,
  planet uuid,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  arrival_time TIMESTAMP WITH TIME ZONE NOT NULL,
  PRIMARY KEY (id),
  FOREIGN KEY (uni) REFERENCES universes(id),
  FOREIGN KEY (objective) REFERENCES fleet_objectives(id),
  FOREIGN KEY (planet) REFERENCES planets(id),
  UNIQUE (uni, name)
);

-- Create the trigger on the table to update the `created_at` field.
CREATE TRIGGER update_fleets_creation BEFORE INSERT ON fleets FOR EACH ROW EXECUTE PROCEDURE update_created_at();

-- Create the table grouping fleet elements with each other.
CREATE TABLE fleet_elements (
  id uuid NOT NULL DEFAULT uuid_generate_v4(),
  fleet uuid NOT NULL,
  player uuid NOT NULL,
  planet uuid NOT NULL,
  speed numeric(3, 2) NOT NULL,
  joined_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  FOREIGN KEY (fleet) REFERENCES fleets(id),
  FOREIGN KEY (planet) REFERENCES planets(id)
);

-- Create the table for vessels belonging to a fleet.
CREATE TABLE fleet_ships (
  id uuid NOT NULL DEFAULT uuid_generate_v4(),
  fleet_element uuid NOT NULL,
  ship uuid NOT NULL,
  count integer NOT NULL DEFAULT 0,
  PRIMARY KEY (id),
  FOREIGN KEY (fleet_element) REFERENCES fleet_elements(id),
  FOREIGN KEY (ship) REFERENCES ships(id),
  UNIQUE (fleet_element, ship)
);

-- Create the table for resources transported by each fleet element.
CREATE TABLE fleet_resources (
  fleet_elem uuid NOT NULL,
  res uuid NOT NULL,
  amount numeric(15, 5) NOT NULL,
  FOREIGN KEY (fleet_elem) REFERENCES fleet_elements(id),
  FOREIGN KEY (res) REFERENCES resources(id)
);

-- Seed the fleet objectives.
INSERT INTO public.fleet_objectives ("name", "hostile", "directed") VALUES('deployment', 'false', 'true');
INSERT INTO public.fleet_objectives ("name", "hostile", "directed") VALUES('transport', 'false', 'true');
INSERT INTO public.fleet_objectives ("name", "hostile", "directed") VALUES('colonization', 'false', 'false');
INSERT INTO public.fleet_objectives ("name", "hostile", "directed") VALUES('expedition', 'false', 'false');
INSERT INTO public.fleet_objectives ("name", "hostile", "directed") VALUES('ACS defend', 'false', 'true');
INSERT INTO public.fleet_objectives ("name", "hostile", "directed") VALUES('ACS attack', 'true', 'true');
INSERT INTO public.fleet_objectives ("name", "hostile", "directed") VALUES('harvesting', 'false', 'false');
INSERT INTO public.fleet_objectives ("name", "hostile", "directed") VALUES('attacking', 'true', 'true');
INSERT INTO public.fleet_objectives ("name", "hostile", "directed") VALUES('espionage', 'true', 'true');
INSERT INTO public.fleet_objectives ("name", "hostile", "directed") VALUES('destroy', 'true', 'true');
