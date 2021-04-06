
-- Create the table defining dependencies between buildings.
CREATE TABLE tech_tree_buildings_vs_buildings (
  element uuid NOT NULL,
  requirement uuid NOT NULL,
  level integer NOT NULL,
  FOREIGN KEY (element) REFERENCES buildings(id),
  FOREIGN KEY (requirement) REFERENCES buildings(id),
  UNIQUE (element, requirement)
);

-- Create the table defining dependencies between technologies.
CREATE TABLE tech_tree_technologies_vs_technologies (
  element uuid NOT NULL,
  requirement uuid NOT NULL,
  level integer NOT NULL,
  FOREIGN KEY (element) REFERENCES technologies(id),
  FOREIGN KEY (requirement) REFERENCES technologies(id),
  UNIQUE (element, requirement)
);

-- Create the table defining dependencies between buildings and technologies.
CREATE TABLE tech_tree_buildings_vs_technologies (
  element uuid NOT NULL,
  requirement uuid NOT NULL,
  level integer NOT NULL,
  FOREIGN KEY (element) REFERENCES buildings(id),
  FOREIGN KEY (requirement) REFERENCES technologies(id),
  UNIQUE (element, requirement)
);

-- Create the table defining dependencies between technologies and buildings.
CREATE TABLE tech_tree_technologies_vs_buildings (
  element uuid NOT NULL,
  requirement uuid NOT NULL,
  level integer NOT NULL,
  FOREIGN KEY (element) REFERENCES technologies(id),
  FOREIGN KEY (requirement) REFERENCES buildings(id),
  UNIQUE (element, requirement)
);

-- Create the table defining dependencies between ships and buildings.
CREATE TABLE tech_tree_ships_vs_buildings (
  element uuid NOT NULL,
  requirement uuid NOT NULL,
  level integer NOT NULL,
  FOREIGN KEY (element) REFERENCES ships(id),
  FOREIGN KEY (requirement) REFERENCES buildings(id),
  UNIQUE (element, requirement)
);

-- Create the table defining dependencies between ships and technologies.
CREATE TABLE tech_tree_ships_vs_technologies (
  element uuid NOT NULL,
  requirement uuid NOT NULL,
  level integer NOT NULL,
  FOREIGN KEY (element) REFERENCES ships(id),
  FOREIGN KEY (requirement) REFERENCES technologies(id),
  UNIQUE (element, requirement)
);

-- Create the table defining dependencies between defenses and buildings.
CREATE TABLE tech_tree_defenses_vs_buildings (
  element uuid NOT NULL,
  requirement uuid NOT NULL,
  level integer NOT NULL,
  FOREIGN KEY (element) REFERENCES defenses(id),
  FOREIGN KEY (requirement) REFERENCES buildings(id),
  UNIQUE (element, requirement)
);

-- Create the table defining dependencies between defenses and technologies.
CREATE TABLE tech_tree_defenses_vs_technologies (
  element uuid NOT NULL,
  requirement uuid NOT NULL,
  level integer NOT NULL,
  FOREIGN KEY (element) REFERENCES defenses(id),
  FOREIGN KEY (requirement) REFERENCES technologies(id),
  UNIQUE (element, requirement)
);

-- Seed dependencies between buildings.
INSERT INTO public.tech_tree_buildings_vs_buildings ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM buildings WHERE name='fusion reactor'),
    (SELECT id FROM buildings WHERE name='deuterium synthetizer'),
    5
  );

INSERT INTO public.tech_tree_buildings_vs_buildings ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM buildings WHERE name='nanite factory'),
    (SELECT id FROM buildings WHERE name='robotics factory'),
    10
  );

INSERT INTO public.tech_tree_buildings_vs_buildings ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM buildings WHERE name='shipyard'),
    (SELECT id FROM buildings WHERE name='robotics factory'),
    2
  );

INSERT INTO public.tech_tree_buildings_vs_buildings ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM buildings WHERE name='terraformer'),
    (SELECT id FROM buildings WHERE name='nanite factory'),
    1
  );

INSERT INTO public.tech_tree_buildings_vs_buildings ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM buildings WHERE name='missile silo'),
    (SELECT id FROM buildings WHERE name='shipyard'),
    1
  );

INSERT INTO public.tech_tree_buildings_vs_buildings ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM buildings WHERE name='sensor phalanx'),
    (SELECT id FROM buildings WHERE name='moon base'),
    1
  );

INSERT INTO public.tech_tree_buildings_vs_buildings ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM buildings WHERE name='jump gate'),
    (SELECT id FROM buildings WHERE name='moon base'),
    1
  );

-- Seed dependencies between technologies.
INSERT INTO public.tech_tree_technologies_vs_technologies ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM technologies WHERE name='shielding'),
    (SELECT id FROM technologies WHERE name='energy'),
    3
  );

INSERT INTO public.tech_tree_technologies_vs_technologies ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM technologies WHERE name='hyperspace'),
    (SELECT id FROM technologies WHERE name='energy'),
    5
  );
INSERT INTO public.tech_tree_technologies_vs_technologies ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM technologies WHERE name='hyperspace'),
    (SELECT id FROM technologies WHERE name='shielding'),
    5
  );

INSERT INTO public.tech_tree_technologies_vs_technologies ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM technologies WHERE name='combustion drive'),
    (SELECT id FROM technologies WHERE name='energy'),
    1
  );

INSERT INTO public.tech_tree_technologies_vs_technologies ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM technologies WHERE name='impulse drive'),
    (SELECT id FROM technologies WHERE name='energy'),
    1
  );

INSERT INTO public.tech_tree_technologies_vs_technologies ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM technologies WHERE name='hyperspace drive'),
    (SELECT id FROM technologies WHERE name='hyperspace'),
    3
  );

INSERT INTO public.tech_tree_technologies_vs_technologies ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM technologies WHERE name='laser'),
    (SELECT id FROM technologies WHERE name='energy'),
    2
  );

INSERT INTO public.tech_tree_technologies_vs_technologies ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM technologies WHERE name='ions'),
    (SELECT id FROM technologies WHERE name='laser'),
    5
  );
INSERT INTO public.tech_tree_technologies_vs_technologies ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM technologies WHERE name='ions'),
    (SELECT id FROM technologies WHERE name='energy'),
    4
  );

INSERT INTO public.tech_tree_technologies_vs_technologies ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM technologies WHERE name='plasma'),
    (SELECT id FROM technologies WHERE name='energy'),
    8
  );
INSERT INTO public.tech_tree_technologies_vs_technologies ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM technologies WHERE name='plasma'),
    (SELECT id FROM technologies WHERE name='laser'),
    10
  );
INSERT INTO public.tech_tree_technologies_vs_technologies ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM technologies WHERE name='plasma'),
    (SELECT id FROM technologies WHERE name='ions'),
    5
  );

INSERT INTO public.tech_tree_technologies_vs_technologies ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM technologies WHERE name='intergalactic research network'),
    (SELECT id FROM technologies WHERE name='computers'),
    8
  );
INSERT INTO public.tech_tree_technologies_vs_technologies ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM technologies WHERE name='intergalactic research network'),
    (SELECT id FROM technologies WHERE name='hyperspace'),
    8
  );

INSERT INTO public.tech_tree_technologies_vs_technologies ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM technologies WHERE name='astrophysics'),
    (SELECT id FROM technologies WHERE name='espionage'),
    4
  );
INSERT INTO public.tech_tree_technologies_vs_technologies ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM technologies WHERE name='astrophysics'),
    (SELECT id FROM technologies WHERE name='impulse drive'),
    3
  );

-- Seed dependencies between buildings and technologies.
INSERT INTO public.tech_tree_buildings_vs_technologies ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM buildings WHERE name='fusion reactor'),
    (SELECT id FROM technologies WHERE name='energy'),
    3
  );

INSERT INTO public.tech_tree_buildings_vs_technologies ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM buildings WHERE name='nanite factory'),
    (SELECT id FROM technologies WHERE name='computers'),
    10
  );

INSERT INTO public.tech_tree_buildings_vs_technologies ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM buildings WHERE name='terraformer'),
    (SELECT id FROM technologies WHERE name='energy'),
    12
  );

INSERT INTO public.tech_tree_buildings_vs_technologies ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM buildings WHERE name='jump gate'),
    (SELECT id FROM technologies WHERE name='hyperspace'),
    7
  );

-- Seed dependencies between technologies and buildings.
INSERT INTO public.tech_tree_technologies_vs_buildings ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM technologies WHERE name='espionage'),
    (SELECT id FROM buildings WHERE name='research lab'),
    3
  );

INSERT INTO public.tech_tree_technologies_vs_buildings ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM technologies WHERE name='computers'),
    (SELECT id FROM buildings WHERE name='research lab'),
    1
  );

INSERT INTO public.tech_tree_technologies_vs_buildings ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM technologies WHERE name='weapons'),
    (SELECT id FROM buildings WHERE name='research lab'),
    4
  );

INSERT INTO public.tech_tree_technologies_vs_buildings ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM technologies WHERE name='shielding'),
    (SELECT id FROM buildings WHERE name='research lab'),
    6
  );

INSERT INTO public.tech_tree_technologies_vs_buildings ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM technologies WHERE name='armour'),
    (SELECT id FROM buildings WHERE name='research lab'),
    2
  );

INSERT INTO public.tech_tree_technologies_vs_buildings ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM technologies WHERE name='energy'),
    (SELECT id FROM buildings WHERE name='research lab'),
    1
  );

INSERT INTO public.tech_tree_technologies_vs_buildings ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM technologies WHERE name='hyperspace'),
    (SELECT id FROM buildings WHERE name='research lab'),
    7
  );

INSERT INTO public.tech_tree_technologies_vs_buildings ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM technologies WHERE name='combustion drive'),
    (SELECT id FROM buildings WHERE name='research lab'),
    1
  );

INSERT INTO public.tech_tree_technologies_vs_buildings ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM technologies WHERE name='impulse drive'),
    (SELECT id FROM buildings WHERE name='research lab'),
    2
  );

INSERT INTO public.tech_tree_technologies_vs_buildings ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM technologies WHERE name='hyperspace drive'),
    (SELECT id FROM buildings WHERE name='research lab'),
    7
  );

INSERT INTO public.tech_tree_technologies_vs_buildings ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM technologies WHERE name='laser'),
    (SELECT id FROM buildings WHERE name='research lab'),
    1
  );

INSERT INTO public.tech_tree_technologies_vs_buildings ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM technologies WHERE name='ions'),
    (SELECT id FROM buildings WHERE name='research lab'),
    4
  );

INSERT INTO public.tech_tree_technologies_vs_buildings ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM technologies WHERE name='plasma'),
    (SELECT id FROM buildings WHERE name='research lab'),
    4
  );

INSERT INTO public.tech_tree_technologies_vs_buildings ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM technologies WHERE name='intergalactic research network'),
    (SELECT id FROM buildings WHERE name='research lab'),
    10
  );

INSERT INTO public.tech_tree_technologies_vs_buildings ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM technologies WHERE name='graviton'),
    (SELECT id FROM buildings WHERE name='research lab'),
    12
  );

INSERT INTO public.tech_tree_technologies_vs_buildings ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM technologies WHERE name='astrophysics'),
    (SELECT id FROM buildings WHERE name='research lab'),
    3
  );

-- Seed dependencies between ships and buildings.
INSERT INTO public.tech_tree_ships_vs_buildings ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM ships WHERE name='small cargo ship'),
    (SELECT id FROM buildings WHERE name='shipyard'),
    2
  );

INSERT INTO public.tech_tree_ships_vs_buildings ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM ships WHERE name='large cargo ship'),
    (SELECT id FROM buildings WHERE name='shipyard'),
    4
  );

INSERT INTO public.tech_tree_ships_vs_buildings ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM ships WHERE name='light fighter'),
    (SELECT id FROM buildings WHERE name='shipyard'),
    1
  );

INSERT INTO public.tech_tree_ships_vs_buildings ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM ships WHERE name='heavy fighter'),
    (SELECT id FROM buildings WHERE name='shipyard'),
    3
  );

INSERT INTO public.tech_tree_ships_vs_buildings ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM ships WHERE name='cruiser'),
    (SELECT id FROM buildings WHERE name='shipyard'),
    5
  );

INSERT INTO public.tech_tree_ships_vs_buildings ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM ships WHERE name='battleship'),
    (SELECT id FROM buildings WHERE name='shipyard'),
    7
  );

INSERT INTO public.tech_tree_ships_vs_buildings ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM ships WHERE name='colony ship'),
    (SELECT id FROM buildings WHERE name='shipyard'),
    4
  );

INSERT INTO public.tech_tree_ships_vs_buildings ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM ships WHERE name='recycler'),
    (SELECT id FROM buildings WHERE name='shipyard'),
    4
  );

INSERT INTO public.tech_tree_ships_vs_buildings ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM ships WHERE name='espionage probe'),
    (SELECT id FROM buildings WHERE name='shipyard'),
    3
  );

INSERT INTO public.tech_tree_ships_vs_buildings ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM ships WHERE name='bomber'),
    (SELECT id FROM buildings WHERE name='shipyard'),
    8
  );

INSERT INTO public.tech_tree_ships_vs_buildings ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM ships WHERE name='solar satellite'),
    (SELECT id FROM buildings WHERE name='shipyard'),
    1
  );

INSERT INTO public.tech_tree_ships_vs_buildings ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM ships WHERE name='destroyer'),
    (SELECT id FROM buildings WHERE name='shipyard'),
    9
  );

INSERT INTO public.tech_tree_ships_vs_buildings ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM ships WHERE name='deathstar'),
    (SELECT id FROM buildings WHERE name='shipyard'),
    12
  );

INSERT INTO public.tech_tree_ships_vs_buildings ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM ships WHERE name='battlecruiser'),
    (SELECT id FROM buildings WHERE name='shipyard'),
    8
  );

-- Seed dependencies between ships and technologies.
INSERT INTO public.tech_tree_ships_vs_technologies ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM ships WHERE name='small cargo ship'),
    (SELECT id FROM technologies WHERE name='combustion drive'),
    2
  );

INSERT INTO public.tech_tree_ships_vs_technologies ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM ships WHERE name='large cargo ship'),
    (SELECT id FROM technologies WHERE name='combustion drive'),
    6
  );

INSERT INTO public.tech_tree_ships_vs_technologies ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM ships WHERE name='light fighter'),
    (SELECT id FROM technologies WHERE name='combustion drive'),
    1
  );

INSERT INTO public.tech_tree_ships_vs_technologies ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM ships WHERE name='heavy fighter'),
    (SELECT id FROM technologies WHERE name='armour'),
    2
  );
INSERT INTO public.tech_tree_ships_vs_technologies ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM ships WHERE name='heavy fighter'),
    (SELECT id FROM technologies WHERE name='impulse drive'),
    2
  );

INSERT INTO public.tech_tree_ships_vs_technologies ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM ships WHERE name='cruiser'),
    (SELECT id FROM technologies WHERE name='impulse drive'),
    4
  );
INSERT INTO public.tech_tree_ships_vs_technologies ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM ships WHERE name='cruiser'),
    (SELECT id FROM technologies WHERE name='ions'),
    2
  );

INSERT INTO public.tech_tree_ships_vs_technologies ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM ships WHERE name='battleship'),
    (SELECT id FROM technologies WHERE name='hyperspace drive'),
    4
  );

INSERT INTO public.tech_tree_ships_vs_technologies ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM ships WHERE name='colony ship'),
    (SELECT id FROM technologies WHERE name='impulse drive'),
    3
  );

INSERT INTO public.tech_tree_ships_vs_technologies ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM ships WHERE name='recycler'),
    (SELECT id FROM technologies WHERE name='combustion drive'),
    6
  );
INSERT INTO public.tech_tree_ships_vs_technologies ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM ships WHERE name='recycler'),
    (SELECT id FROM technologies WHERE name='shielding'),
    2
  );

INSERT INTO public.tech_tree_ships_vs_technologies ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM ships WHERE name='espionage probe'),
    (SELECT id FROM technologies WHERE name='combustion drive'),
    3
  );
INSERT INTO public.tech_tree_ships_vs_technologies ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM ships WHERE name='espionage probe'),
    (SELECT id FROM technologies WHERE name='espionage'),
    2
  );

INSERT INTO public.tech_tree_ships_vs_technologies ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM ships WHERE name='bomber'),
    (SELECT id FROM technologies WHERE name='impulse drive'),
    6
  );
INSERT INTO public.tech_tree_ships_vs_technologies ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM ships WHERE name='bomber'),
    (SELECT id FROM technologies WHERE name='plasma'),
    5
  );

INSERT INTO public.tech_tree_ships_vs_technologies ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM ships WHERE name='destroyer'),
    (SELECT id FROM technologies WHERE name='hyperspace drive'),
    6
  );
INSERT INTO public.tech_tree_ships_vs_technologies ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM ships WHERE name='destroyer'),
    (SELECT id FROM technologies WHERE name='hyperspace'),
    5
  );

INSERT INTO public.tech_tree_ships_vs_technologies ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM ships WHERE name='deathstar'),
    (SELECT id FROM technologies WHERE name='hyperspace drive'),
    7
  );
INSERT INTO public.tech_tree_ships_vs_technologies ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM ships WHERE name='deathstar'),
    (SELECT id FROM technologies WHERE name='hyperspace'),
    6
  );
INSERT INTO public.tech_tree_ships_vs_technologies ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM ships WHERE name='deathstar'),
    (SELECT id FROM technologies WHERE name='graviton'),
    1
  );

INSERT INTO public.tech_tree_ships_vs_technologies ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM ships WHERE name='battlecruiser'),
    (SELECT id FROM technologies WHERE name='hyperspace drive'),
    5
  );
INSERT INTO public.tech_tree_ships_vs_technologies ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM ships WHERE name='battlecruiser'),
    (SELECT id FROM technologies WHERE name='hyperspace'),
    5
  );
INSERT INTO public.tech_tree_ships_vs_technologies ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM ships WHERE name='battlecruiser'),
    (SELECT id FROM technologies WHERE name='laser'),
    12
  );

-- Seed dependencies between defenses and buildings.
INSERT INTO public.tech_tree_defenses_vs_buildings ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM defenses WHERE name='rocket launcher'),
    (SELECT id FROM buildings WHERE name='shipyard'),
    1
  );

INSERT INTO public.tech_tree_defenses_vs_buildings ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM defenses WHERE name='light laser'),
    (SELECT id FROM buildings WHERE name='shipyard'),
    2
  );

INSERT INTO public.tech_tree_defenses_vs_buildings ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM defenses WHERE name='heavy laser'),
    (SELECT id FROM buildings WHERE name='shipyard'),
    4
  );

INSERT INTO public.tech_tree_defenses_vs_buildings ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM defenses WHERE name='gauss cannon'),
    (SELECT id FROM buildings WHERE name='shipyard'),
    6
  );

INSERT INTO public.tech_tree_defenses_vs_buildings ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM defenses WHERE name='ion cannon'),
    (SELECT id FROM buildings WHERE name='shipyard'),
    4
  );

INSERT INTO public.tech_tree_defenses_vs_buildings ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM defenses WHERE name='plasma turret'),
    (SELECT id FROM buildings WHERE name='shipyard'),
    8
  );

INSERT INTO public.tech_tree_defenses_vs_buildings ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM defenses WHERE name='small shield dome'),
    (SELECT id FROM buildings WHERE name='shipyard'),
    1
  );

INSERT INTO public.tech_tree_defenses_vs_buildings ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM defenses WHERE name='large shield dome'),
    (SELECT id FROM buildings WHERE name='shipyard'),
    6
  );

INSERT INTO public.tech_tree_defenses_vs_buildings ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM defenses WHERE name='anti-ballistic missile'),
    (SELECT id FROM buildings WHERE name='shipyard'),
    1
  );

INSERT INTO public.tech_tree_defenses_vs_buildings ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM defenses WHERE name='anti-ballistic missile'),
    (SELECT id FROM buildings WHERE name='missile silo'),
    2
  );

INSERT INTO public.tech_tree_defenses_vs_buildings ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM defenses WHERE name='interplanetary missile'),
    (SELECT id FROM buildings WHERE name='shipyard'),
    1
  );

INSERT INTO public.tech_tree_defenses_vs_buildings ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM defenses WHERE name='interplanetary missile'),
    (SELECT id FROM buildings WHERE name='missile silo'),
    4
  );

-- Seed dependencies between defenses and technologies.
INSERT INTO public.tech_tree_defenses_vs_technologies ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM defenses WHERE name='light laser'),
    (SELECT id FROM technologies WHERE name='energy'),
    1
  );
INSERT INTO public.tech_tree_defenses_vs_technologies ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM defenses WHERE name='light laser'),
    (SELECT id FROM technologies WHERE name='laser'),
    3
  );

INSERT INTO public.tech_tree_defenses_vs_technologies ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM defenses WHERE name='heavy laser'),
    (SELECT id FROM technologies WHERE name='energy'),
    3
  );
INSERT INTO public.tech_tree_defenses_vs_technologies ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM defenses WHERE name='heavy laser'),
    (SELECT id FROM technologies WHERE name='laser'),
    6
  );

INSERT INTO public.tech_tree_defenses_vs_technologies ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM defenses WHERE name='gauss cannon'),
    (SELECT id FROM technologies WHERE name='energy'),
    6
  );
INSERT INTO public.tech_tree_defenses_vs_technologies ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM defenses WHERE name='gauss cannon'),
    (SELECT id FROM technologies WHERE name='weapons'),
    3
  );
INSERT INTO public.tech_tree_defenses_vs_technologies ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM defenses WHERE name='gauss cannon'),
    (SELECT id FROM technologies WHERE name='shielding'),
    1
  );

INSERT INTO public.tech_tree_defenses_vs_technologies ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM defenses WHERE name='ion cannon'),
    (SELECT id FROM technologies WHERE name='ions'),
    4
  );

INSERT INTO public.tech_tree_defenses_vs_technologies ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM defenses WHERE name='plasma turret'),
    (SELECT id FROM technologies WHERE name='plasma'),
    7
  );

INSERT INTO public.tech_tree_defenses_vs_technologies ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM defenses WHERE name='small shield dome'),
    (SELECT id FROM technologies WHERE name='shielding'),
    2
  );

INSERT INTO public.tech_tree_defenses_vs_technologies ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM defenses WHERE name='large shield dome'),
    (SELECT id FROM technologies WHERE name='shielding'),
    6
  );

INSERT INTO public.tech_tree_defenses_vs_technologies ("element", "requirement", "level")
  VALUES(
    (SELECT id FROM defenses WHERE name='interplanetary missile'),
    (SELECT id FROM technologies WHERE name='impulse drive'),
    1
  );
