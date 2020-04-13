
-- Create the table defining the purpose of a fleet.
CREATE TABLE fleet_objectives (
  id uuid NOT NULL DEFAULT uuid_generate_v4(),
  name text NOT NULL,
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
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  arrival_time TIMESTAMP WITH TIME ZONE NOT NULL,
  PRIMARY KEY (id),
  FOREIGN KEY (uni) REFERENCES universes(id),
  FOREIGN KEY (objective) REFERENCES fleet_objectives(id),
  UNIQUE (uni, name)
);

-- Create the trigger on the table to update the `created_at` field.
CREATE TRIGGER update_fleets_creation BEFORE INSERT ON fleets FOR EACH ROW EXECUTE PROCEDURE update_created_at();

-- Create the table grouping fleet elements with each other.
CREATE TABLE fleet_elements (
  id uuid NOT NULL DEFAULT uuid_generate_v4(),
  fleet uuid NOT NULL,
  player uuid NOT NULL,
  start_galaxy integer NOT NULL,
  start_solar_system integer NOT NULL,
  start_position integer NOT NULL,
  speed numeric(3, 2) NOT NULL,
  joined_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  FOREIGN KEY (fleet) REFERENCES fleets(id)
);

-- Create the trigger on the table to update the `joined_at` field.
CREATE TRIGGER update_fleets_elements_creation BEFORE INSERT ON fleet_elements FOR EACH ROW EXECUTE PROCEDURE update_joined_at();

-- Create the table for vessels belonging to a fleet.
CREATE TABLE fleet_ships (
  id uuid NOT NULL DEFAULT uuid_generate_v4(),
  fleet_element uuid NOT NULL,
  ship uuid NOT NULL,
  amount integer NOT NULL DEFAULT 0,
  PRIMARY KEY (id),
  FOREIGN KEY (fleet_element) REFERENCES fleet_elements(id),
  FOREIGN KEY (ship) REFERENCES ships(id)
);

-- Create the table for resources transported by each fleet element.
CREATE TABLE fleet_resources (
  fleet_elem uuid NOT NULL,
  res uuid NOT NULL,
  amount integer NOT NULL,
  FOREIGN KEY (fleet_elem) REFERENCES fleet_elements(id),
  FOREIGN KEY (res) REFERENCES resources(id)
);

-- Seed the fleet objectives.
INSERT INTO public.fleet_objectives ("name") VALUES('attacking');
INSERT INTO public.fleet_objectives ("name") VALUES('deployment');
INSERT INTO public.fleet_objectives ("name") VALUES('espionage');
INSERT INTO public.fleet_objectives ("name") VALUES('transport');
INSERT INTO public.fleet_objectives ("name") VALUES('colonization');
INSERT INTO public.fleet_objectives ("name") VALUES('harvesting');
INSERT INTO public.fleet_objectives ("name") VALUES('destroy');
INSERT INTO public.fleet_objectives ("name") VALUES('expedition');
INSERT INTO public.fleet_objectives ("name") VALUES('ACS attack');
INSERT INTO public.fleet_objectives ("name") VALUES('ACS defend');
