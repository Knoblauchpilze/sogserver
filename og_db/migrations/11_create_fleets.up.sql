
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
  uni uuid NOT NULL,
  objective uuid NOT NULL,
  player uuid NOT NULL,
  source uuid NOT NULL,
  source_type text NOT NULL,
  target_galaxy integer NOT NULL,
  target_solar_system integer NOT NULL,
  target_position integer NOT NULL,
  target uuid,
  target_type text NOT NULL,
  speed numeric(3, 2) NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  arrival_time TIMESTAMP WITH TIME ZONE NOT NULL,
  return_time TIMESTAMP WITH TIME ZONE NOT NULL,
  is_returning boolean NOT NULL DEFAULT false,
  PRIMARY KEY (id),
  FOREIGN KEY (uni) REFERENCES universes(id),
  FOREIGN KEY (objective) REFERENCES fleet_objectives(id),
  FOREIGN KEY (player) REFERENCES players(id)
);

-- Create the trigger on the table to update the `created_at` field.
CREATE TRIGGER update_fleets_creation BEFORE INSERT ON fleets FOR EACH ROW EXECUTE PROCEDURE update_created_at();

-- Create the table for vessels belonging to a fleet.
CREATE TABLE fleet_ships (
  id uuid NOT NULL DEFAULT uuid_generate_v4(),
  fleet uuid NOT NULL,
  ship uuid NOT NULL,
  count integer NOT NULL DEFAULT 0,
  PRIMARY KEY (id),
  FOREIGN KEY (fleet) REFERENCES fleets(id),
  FOREIGN KEY (ship) REFERENCES ships(id),
  UNIQUE (fleet, ship)
);

-- Create the table for resources transported by each fleet element.
CREATE TABLE fleet_resources (
  fleet uuid NOT NULL,
  resource uuid NOT NULL,
  amount numeric(15, 5) NOT NULL,
  FOREIGN KEY (fleet) REFERENCES fleets(id),
  FOREIGN KEY (resource) REFERENCES resources(id)
);

-- Create the table defining the ACS operations.
CREATE TABLE fleets_acs (
  id uuid NOT NULL DEFAULT uuid_generate_v4(),
  universe uuid NOT NULL,
  name text NOT NULL,
  objective uuid NOT NULL,
  target uuid NOT NULL,
  target_type text NOT NULL,
  PRIMARY KEY (id),
  FOREIGN KEY (universe) REFERENCES universes(id),
  FOREIGN KEY (objective) REFERENCES fleet_objectives(id),
  UNIQUE (universe, name)
);

-- Create the table registering the participants to an ACS operation.
CREATE TABLE fleets_acs_components (
  acs uuid NOT NULL,
  fleet uuid NOT NULL,
  FOREIGN KEY (acs) REFERENCES fleets_acs(id),
  FOREIGN KEY (fleet) REFERENCES fleets(id)
);

-- Create the table indicating which ship can be used for which missions.
CREATE TABLE ships_usage (
  ship uuid NOT NULL,
  objective uuid NOT NULL,
  usable boolean NOT NULL,
  FOREIGN KEY (ship) REFERENCES ships(id),
  FOREIGN KEY (objective) REFERENCES fleet_objectives(id)
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

-- Seed the ships' usage.
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='small cargo ship'),
    (SELECT id FROM fleet_objectives WHERE name='deployment'),
    'true'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='large cargo ship'),
    (SELECT id FROM fleet_objectives WHERE name='deployment'),
    'true'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='light fighter'),
    (SELECT id FROM fleet_objectives WHERE name='deployment'),
    'true'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='heavy fighter'),
    (SELECT id FROM fleet_objectives WHERE name='deployment'),
    'true'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='cruiser'),
    (SELECT id FROM fleet_objectives WHERE name='deployment'),
    'true'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='battleship'),
    (SELECT id FROM fleet_objectives WHERE name='deployment'),
    'true'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='battlecruiser'),
    (SELECT id FROM fleet_objectives WHERE name='deployment'),
    'true'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='bomber'),
    (SELECT id FROM fleet_objectives WHERE name='deployment'),
    'true'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='destroyer'),
    (SELECT id FROM fleet_objectives WHERE name='deployment'),
    'true'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='deathstar'),
    (SELECT id FROM fleet_objectives WHERE name='deployment'),
    'true'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='recycler'),
    (SELECT id FROM fleet_objectives WHERE name='deployment'),
    'true'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='espionage probe'),
    (SELECT id FROM fleet_objectives WHERE name='deployment'),
    'true'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='solar satellite'),
    (SELECT id FROM fleet_objectives WHERE name='deployment'),
    'false'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='colony ship'),
    (SELECT id FROM fleet_objectives WHERE name='deployment'),
    'true'
  );

INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='small cargo ship'),
    (SELECT id FROM fleet_objectives WHERE name='transport'),
    'true'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='large cargo ship'),
    (SELECT id FROM fleet_objectives WHERE name='transport'),
    'true'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='light fighter'),
    (SELECT id FROM fleet_objectives WHERE name='transport'),
    'true'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='heavy fighter'),
    (SELECT id FROM fleet_objectives WHERE name='transport'),
    'true'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='cruiser'),
    (SELECT id FROM fleet_objectives WHERE name='transport'),
    'true'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='battleship'),
    (SELECT id FROM fleet_objectives WHERE name='transport'),
    'true'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='battlecruiser'),
    (SELECT id FROM fleet_objectives WHERE name='transport'),
    'true'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='bomber'),
    (SELECT id FROM fleet_objectives WHERE name='transport'),
    'true'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='destroyer'),
    (SELECT id FROM fleet_objectives WHERE name='transport'),
    'true'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='deathstar'),
    (SELECT id FROM fleet_objectives WHERE name='transport'),
    'true'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='recycler'),
    (SELECT id FROM fleet_objectives WHERE name='transport'),
    'true'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='espionage probe'),
    (SELECT id FROM fleet_objectives WHERE name='transport'),
    'true'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='solar satellite'),
    (SELECT id FROM fleet_objectives WHERE name='transport'),
    'false'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='colony ship'),
    (SELECT id FROM fleet_objectives WHERE name='transport'),
    'true'
  );

INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='small cargo ship'),
    (SELECT id FROM fleet_objectives WHERE name='colonization'),
    'false'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='large cargo ship'),
    (SELECT id FROM fleet_objectives WHERE name='colonization'),
    'false'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='light fighter'),
    (SELECT id FROM fleet_objectives WHERE name='colonization'),
    'false'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='heavy fighter'),
    (SELECT id FROM fleet_objectives WHERE name='colonization'),
    'false'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='cruiser'),
    (SELECT id FROM fleet_objectives WHERE name='colonization'),
    'false'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='battleship'),
    (SELECT id FROM fleet_objectives WHERE name='colonization'),
    'false'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='battlecruiser'),
    (SELECT id FROM fleet_objectives WHERE name='colonization'),
    'false'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='bomber'),
    (SELECT id FROM fleet_objectives WHERE name='colonization'),
    'false'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='destroyer'),
    (SELECT id FROM fleet_objectives WHERE name='colonization'),
    'false'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='deathstar'),
    (SELECT id FROM fleet_objectives WHERE name='colonization'),
    'false'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='recycler'),
    (SELECT id FROM fleet_objectives WHERE name='colonization'),
    'false'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='espionage probe'),
    (SELECT id FROM fleet_objectives WHERE name='colonization'),
    'false'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='solar satellite'),
    (SELECT id FROM fleet_objectives WHERE name='colonization'),
    'false'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='colony ship'),
    (SELECT id FROM fleet_objectives WHERE name='colonization'),
    'true'
  );

INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='small cargo ship'),
    (SELECT id FROM fleet_objectives WHERE name='expedition'),
    'true'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='large cargo ship'),
    (SELECT id FROM fleet_objectives WHERE name='expedition'),
    'true'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='light fighter'),
    (SELECT id FROM fleet_objectives WHERE name='expedition'),
    'true'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='heavy fighter'),
    (SELECT id FROM fleet_objectives WHERE name='expedition'),
    'true'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='cruiser'),
    (SELECT id FROM fleet_objectives WHERE name='expedition'),
    'true'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='battleship'),
    (SELECT id FROM fleet_objectives WHERE name='expedition'),
    'true'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='battlecruiser'),
    (SELECT id FROM fleet_objectives WHERE name='expedition'),
    'true'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='bomber'),
    (SELECT id FROM fleet_objectives WHERE name='expedition'),
    'true'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='destroyer'),
    (SELECT id FROM fleet_objectives WHERE name='expedition'),
    'true'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='deathstar'),
    (SELECT id FROM fleet_objectives WHERE name='expedition'),
    'true'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='recycler'),
    (SELECT id FROM fleet_objectives WHERE name='expedition'),
    'true'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='espionage probe'),
    (SELECT id FROM fleet_objectives WHERE name='expedition'),
    'true'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='solar satellite'),
    (SELECT id FROM fleet_objectives WHERE name='expedition'),
    'false'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='colony ship'),
    (SELECT id FROM fleet_objectives WHERE name='expedition'),
    'true'
  );

INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='small cargo ship'),
    (SELECT id FROM fleet_objectives WHERE name='ACS defend'),
    'true'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='large cargo ship'),
    (SELECT id FROM fleet_objectives WHERE name='ACS defend'),
    'true'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='light fighter'),
    (SELECT id FROM fleet_objectives WHERE name='ACS defend'),
    'true'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='heavy fighter'),
    (SELECT id FROM fleet_objectives WHERE name='ACS defend'),
    'true'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='cruiser'),
    (SELECT id FROM fleet_objectives WHERE name='ACS defend'),
    'true'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='battleship'),
    (SELECT id FROM fleet_objectives WHERE name='ACS defend'),
    'true'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='battlecruiser'),
    (SELECT id FROM fleet_objectives WHERE name='ACS defend'),
    'true'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='bomber'),
    (SELECT id FROM fleet_objectives WHERE name='ACS defend'),
    'true'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='destroyer'),
    (SELECT id FROM fleet_objectives WHERE name='ACS defend'),
    'true'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='deathstar'),
    (SELECT id FROM fleet_objectives WHERE name='ACS defend'),
    'true'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='recycler'),
    (SELECT id FROM fleet_objectives WHERE name='ACS defend'),
    'true'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='espionage probe'),
    (SELECT id FROM fleet_objectives WHERE name='ACS defend'),
    'true'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='solar satellite'),
    (SELECT id FROM fleet_objectives WHERE name='ACS defend'),
    'false'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='colony ship'),
    (SELECT id FROM fleet_objectives WHERE name='ACS defend'),
    'true'
  );

INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='small cargo ship'),
    (SELECT id FROM fleet_objectives WHERE name='ACS attack'),
    'true'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='large cargo ship'),
    (SELECT id FROM fleet_objectives WHERE name='ACS attack'),
    'true'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='light fighter'),
    (SELECT id FROM fleet_objectives WHERE name='ACS attack'),
    'true'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='heavy fighter'),
    (SELECT id FROM fleet_objectives WHERE name='ACS attack'),
    'true'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='cruiser'),
    (SELECT id FROM fleet_objectives WHERE name='ACS attack'),
    'true'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='battleship'),
    (SELECT id FROM fleet_objectives WHERE name='ACS attack'),
    'true'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='battlecruiser'),
    (SELECT id FROM fleet_objectives WHERE name='ACS attack'),
    'true'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='bomber'),
    (SELECT id FROM fleet_objectives WHERE name='ACS attack'),
    'true'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='destroyer'),
    (SELECT id FROM fleet_objectives WHERE name='ACS attack'),
    'true'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='deathstar'),
    (SELECT id FROM fleet_objectives WHERE name='ACS attack'),
    'true'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='recycler'),
    (SELECT id FROM fleet_objectives WHERE name='ACS attack'),
    'true'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='espionage probe'),
    (SELECT id FROM fleet_objectives WHERE name='ACS attack'),
    'true'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='solar satellite'),
    (SELECT id FROM fleet_objectives WHERE name='ACS attack'),
    'false'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='colony ship'),
    (SELECT id FROM fleet_objectives WHERE name='ACS attack'),
    'true'
  );

INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='small cargo ship'),
    (SELECT id FROM fleet_objectives WHERE name='harvesting'),
    'false'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='large cargo ship'),
    (SELECT id FROM fleet_objectives WHERE name='harvesting'),
    'false'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='light fighter'),
    (SELECT id FROM fleet_objectives WHERE name='harvesting'),
    'false'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='heavy fighter'),
    (SELECT id FROM fleet_objectives WHERE name='harvesting'),
    'false'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='cruiser'),
    (SELECT id FROM fleet_objectives WHERE name='harvesting'),
    'false'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='battleship'),
    (SELECT id FROM fleet_objectives WHERE name='harvesting'),
    'false'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='battlecruiser'),
    (SELECT id FROM fleet_objectives WHERE name='harvesting'),
    'false'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='bomber'),
    (SELECT id FROM fleet_objectives WHERE name='harvesting'),
    'false'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='destroyer'),
    (SELECT id FROM fleet_objectives WHERE name='harvesting'),
    'false'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='deathstar'),
    (SELECT id FROM fleet_objectives WHERE name='harvesting'),
    'false'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='recycler'),
    (SELECT id FROM fleet_objectives WHERE name='harvesting'),
    'true'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='espionage probe'),
    (SELECT id FROM fleet_objectives WHERE name='harvesting'),
    'false'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='solar satellite'),
    (SELECT id FROM fleet_objectives WHERE name='harvesting'),
    'false'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='colony ship'),
    (SELECT id FROM fleet_objectives WHERE name='harvesting'),
    'false'
  );

INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='small cargo ship'),
    (SELECT id FROM fleet_objectives WHERE name='attacking'),
    'true'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='large cargo ship'),
    (SELECT id FROM fleet_objectives WHERE name='attacking'),
    'true'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='light fighter'),
    (SELECT id FROM fleet_objectives WHERE name='attacking'),
    'true'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='heavy fighter'),
    (SELECT id FROM fleet_objectives WHERE name='attacking'),
    'true'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='cruiser'),
    (SELECT id FROM fleet_objectives WHERE name='attacking'),
    'true'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='battleship'),
    (SELECT id FROM fleet_objectives WHERE name='attacking'),
    'true'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='battlecruiser'),
    (SELECT id FROM fleet_objectives WHERE name='attacking'),
    'true'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='bomber'),
    (SELECT id FROM fleet_objectives WHERE name='attacking'),
    'true'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='destroyer'),
    (SELECT id FROM fleet_objectives WHERE name='attacking'),
    'true'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='deathstar'),
    (SELECT id FROM fleet_objectives WHERE name='attacking'),
    'true'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='recycler'),
    (SELECT id FROM fleet_objectives WHERE name='attacking'),
    'true'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='espionage probe'),
    (SELECT id FROM fleet_objectives WHERE name='attacking'),
    'true'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='solar satellite'),
    (SELECT id FROM fleet_objectives WHERE name='attacking'),
    'false'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='colony ship'),
    (SELECT id FROM fleet_objectives WHERE name='attacking'),
    'true'
  );

INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='small cargo ship'),
    (SELECT id FROM fleet_objectives WHERE name='espionage'),
    'false'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='large cargo ship'),
    (SELECT id FROM fleet_objectives WHERE name='espionage'),
    'false'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='light fighter'),
    (SELECT id FROM fleet_objectives WHERE name='espionage'),
    'false'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='heavy fighter'),
    (SELECT id FROM fleet_objectives WHERE name='espionage'),
    'false'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='cruiser'),
    (SELECT id FROM fleet_objectives WHERE name='espionage'),
    'false'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='battleship'),
    (SELECT id FROM fleet_objectives WHERE name='espionage'),
    'false'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='battlecruiser'),
    (SELECT id FROM fleet_objectives WHERE name='espionage'),
    'false'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='bomber'),
    (SELECT id FROM fleet_objectives WHERE name='espionage'),
    'false'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='destroyer'),
    (SELECT id FROM fleet_objectives WHERE name='espionage'),
    'false'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='deathstar'),
    (SELECT id FROM fleet_objectives WHERE name='espionage'),
    'false'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='recycler'),
    (SELECT id FROM fleet_objectives WHERE name='espionage'),
    'false'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='espionage probe'),
    (SELECT id FROM fleet_objectives WHERE name='espionage'),
    'true'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='solar satellite'),
    (SELECT id FROM fleet_objectives WHERE name='espionage'),
    'false'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='colony ship'),
    (SELECT id FROM fleet_objectives WHERE name='espionage'),
    'false'
  );

INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='small cargo ship'),
    (SELECT id FROM fleet_objectives WHERE name='destroy'),
    'false'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='large cargo ship'),
    (SELECT id FROM fleet_objectives WHERE name='destroy'),
    'false'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='light fighter'),
    (SELECT id FROM fleet_objectives WHERE name='destroy'),
    'false'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='heavy fighter'),
    (SELECT id FROM fleet_objectives WHERE name='destroy'),
    'false'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='cruiser'),
    (SELECT id FROM fleet_objectives WHERE name='destroy'),
    'false'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='battleship'),
    (SELECT id FROM fleet_objectives WHERE name='destroy'),
    'false'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='battlecruiser'),
    (SELECT id FROM fleet_objectives WHERE name='destroy'),
    'false'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='bomber'),
    (SELECT id FROM fleet_objectives WHERE name='destroy'),
    'false'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='destroyer'),
    (SELECT id FROM fleet_objectives WHERE name='destroy'),
    'false'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='deathstar'),
    (SELECT id FROM fleet_objectives WHERE name='destroy'),
    'true'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='recycler'),
    (SELECT id FROM fleet_objectives WHERE name='destroy'),
    'false'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='espionage probe'),
    (SELECT id FROM fleet_objectives WHERE name='destroy'),
    'false'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='solar satellite'),
    (SELECT id FROM fleet_objectives WHERE name='destroy'),
    'false'
  );
INSERT INTO public.ships_usage ("ship", "objective", "usable")
  VALUES(
    (SELECT id FROM ships WHERE name='colony ship'),
    (SELECT id FROM fleet_objectives WHERE name='destroy'),
    'false'
  );
