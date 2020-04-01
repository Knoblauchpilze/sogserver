
-- Create the table defining the purpose of a fleet.
CREATE TABLE fleet_objectives (
  id uuid NOT NULL DEFAULT uuid_generate_v4(),
  name text NOT NULL,
  PRIMARY KEY (id)
);

-- Create the table defining fleets.
CREATE TABLE fleets (
    id uuid NOT NULL DEFAULT uuid_generate_v4(),
    name text,
    objective uuid NOT NULL,
    galaxy integer NOT NULL,
    solar_system integer NOT NULL,
    position integer NOT NULL,
    created_at timestamp with time zone default current_timestamp,
    arrival_time timestamp with time zone default current_timestamp,
    PRIMARY KEY (id),
    FOREIGN KEY (objective) REFERENCES fleet_objectives(id)
);

-- Trigger to update the `created_at` field of the table.
CREATE TRIGGER update_fleet_creation_time BEFORE INSERT ON fleets FOR EACH ROW EXECUTE PROCEDURE update_created_at_column();

-- Create the table for vessels belonging to a fleet.
CREATE TABLE fleet_ships (
  fleet uuid NOT NULL,
  ship uuid NOT NULL,
  player uuid NOT NULL,
  amount integer NOT NULL default 0,
  start_galaxy integer NOT NULL,
  start_solar_system integer NOT NULL,
  start_position integer NOT NULL,
  speed numeric(3, 2) NOT NULL,
  joined_at timestamp with time zone default current_timestamp,
  FOREIGN KEY (fleet) REFERENCES fleets(id),
  FOREIGN KEY (ship) REFERENCES ships(id),
  FOREIGN KEY (player) REFERENCES players(id)
);

-- Trigger to update the `joined_at` field of the table.
CREATE TRIGGER update_ships_joined_at_time BEFORE INSERT ON fleet_ships FOR EACH ROW EXECUTE PROCEDURE update_created_at_column();

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
