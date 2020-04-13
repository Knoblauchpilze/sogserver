
-- Create the building types.
CREATE TABLE building_types (
  id uuid NOT NULL DEFAULT uuid_generate_v4(),
  name text NOT NULL,
  PRIMARY KEY (id),
  UNIQUE (name)
);

-- Create the table defining buildings.
CREATE TABLE buildings (
  id uuid NOT NULL DEFAULT uuid_generate_v4(),
  name text NOT NULL,
  type uuid NOT NULL,
  PRIMARY KEY (id),
  FOREIGN KEY (type) REFERENCES building_types(id)
);

-- Create the table defining the cost of a building.
CREATE TABLE buildings_costs (
  building uuid NOT NULL,
  res uuid NOT NULL,
  cost integer NOT NULL,
  FOREIGN KEY (building) REFERENCES buildings(id),
  FOREIGN KEY (res) REFERENCES resources(id)
);

-- Create the table defining the law of progression of cost of a building.
CREATE TABLE buildings_costs_progress (
  building uuid NOT NULL,
  progress numeric(15, 5) NOT NULL,
  FOREIGN KEY (building) REFERENCES buildings(id)
);

-- Create the table defining the law of progression of gains of a building.
CREATE TABLE buildings_gains_progress (
  building uuid NOT NULL,
  res uuid NOT NULL,
  base integer NOT NULL,
  progress numeric(15, 5) NOT NULL,
  temperature_coeff numeric(15, 5) NOT NULL,
  temperature_offset numeric(15, 5) NOT NULL,
  FOREIGN KEY (building) REFERENCES buildings(id),
  FOREIGN KEY (res) REFERENCES resources(id)
);

-- Create the table defining the law of progression of storage of a building.
CREATE TABLE buildings_storage_progress (
  building uuid NOT NULL,
  base integer NOT NULL,
  multiplier numeric(15, 5) NOT NULL,
  progress numeric(15, 5) NOT NULL,
  FOREIGN KEY (building) REFERENCES buildings(id)
);

-- Seed the available building types.
INSERT INTO public.building_types ("name") VALUES('mine');
INSERT INTO public.building_types ("name") VALUES('hangar');
INSERT INTO public.building_types ("name") VALUES('none');

-- Seed the available buildings.
INSERT INTO public.buildings ("name", "type")
  VALUES(
    'metal mine',
    (SELECT id FROM building_types WHERE name='mine')
  );
INSERT INTO public.buildings ("name", "type")
  VALUES(
    'crystal mine',
    (SELECT id FROM building_types WHERE name='mine')
  );
INSERT INTO public.buildings ("name", "type")
  VALUES(
    'deuterium synthetizer',
    (SELECT id FROM building_types WHERE name='mine')
  );

INSERT INTO public.buildings ("name", "type")
  VALUES(
    'metal storage',
    (SELECT id FROM building_types WHERE name='hangar')
  );
INSERT INTO public.buildings ("name", "type")
  VALUES(
    'crystal storage',
    (SELECT id FROM building_types WHERE name='hangar')
  );
INSERT INTO public.buildings ("name", "type")
  VALUES(
    'deuterium tank',
    (SELECT id FROM building_types WHERE name='hangar')
  );

INSERT INTO public.buildings ("name", "type")
  VALUES(
    'solar plant',
    (SELECT id FROM building_types WHERE name='mine')
  );
INSERT INTO public.buildings ("name", "type")
  VALUES(
    'fusion reactor',
    (SELECT id FROM building_types WHERE name='mine')
  );

INSERT INTO public.buildings ("name", "type")
  VALUES(
    'robotics factory',
    (SELECT id FROM building_types WHERE name='none')
  );
INSERT INTO public.buildings ("name", "type")
  VALUES(
    'shipyard',
    (SELECT id FROM building_types WHERE name='none')
  );
INSERT INTO public.buildings ("name", "type")
  VALUES(
    'research lab',
    (SELECT id FROM building_types WHERE name='none')
  );
INSERT INTO public.buildings ("name", "type")
  VALUES(
    'alliance depot',
    (SELECT id FROM building_types WHERE name='none')
  );
INSERT INTO public.buildings ("name", "type")
  VALUES(
    'missile silo',
    (SELECT id FROM building_types WHERE name='none')
  );
INSERT INTO public.buildings ("name", "type")
  VALUES(
    'nanite factory',
    (SELECT id FROM building_types WHERE name='none')
  );
INSERT INTO public.buildings ("name", "type")
  VALUES(
    'terraformer',
    (SELECT id FROM building_types WHERE name='none')
  );
INSERT INTO public.buildings ("name", "type")
  VALUES(
    'space dock',
    (SELECT id FROM building_types WHERE name='none')
  );

INSERT INTO public.buildings ("name", "type")
  VALUES(
    'moon base',
    (SELECT id FROM building_types WHERE name='none')
  );
INSERT INTO public.buildings ("name", "type")
  VALUES(
    'jump gate',
    (SELECT id FROM building_types WHERE name='none')
  );
INSERT INTO public.buildings ("name", "type")
  VALUES(
    'sensor phalanx',
    (SELECT id FROM building_types WHERE name='none')
  );

-- Seed the building costs.
-- Mines.
INSERT INTO public.buildings_costs ("building", "res", "cost")
  VALUES(
    (SELECT id FROM buildings WHERE name='metal mine'),
    (SELECT id FROM resources WHERE name='metal'),
    60
  );

INSERT INTO public.buildings_costs ("building", "res", "cost")
  VALUES(
    (SELECT id FROM buildings WHERE name='metal mine'),
    (SELECT id FROM resources WHERE name='crystal'),
    15
  );

INSERT INTO public.buildings_costs ("building", "res", "cost")
  VALUES(
    (SELECT id FROM buildings WHERE name='crystal mine'),
    (SELECT id FROM resources WHERE name='metal'),
    48
  );
INSERT INTO public.buildings_costs ("building", "res", "cost")
  VALUES(
    (SELECT id FROM buildings WHERE name='crystal mine'),
    (SELECT id FROM resources WHERE name='crystal'),
    24
  );

INSERT INTO public.buildings_costs ("building", "res", "cost")
  VALUES(
    (SELECT id FROM buildings WHERE name='deuterium synthetizer'),
    (SELECT id FROM resources WHERE name='metal'),
    225
  );
INSERT INTO public.buildings_costs ("building", "res", "cost")
  VALUES(
    (SELECT id FROM buildings WHERE name='deuterium synthetizer'),
    (SELECT id FROM resources WHERE name='crystal'),
    75
  );

-- Storages.
INSERT INTO public.buildings_costs ("building", "res", "cost")
  VALUES(
    (SELECT id FROM buildings WHERE name='metal storage'),
    (SELECT id FROM resources WHERE name='metal'),
    1000
  );

INSERT INTO public.buildings_costs ("building", "res", "cost")
  VALUES(
    (SELECT id FROM buildings WHERE name='crystal storage'),
    (SELECT id FROM resources WHERE name='metal'),
    1000
  );
INSERT INTO public.buildings_costs ("building", "res", "cost")
  VALUES(
    (SELECT id FROM buildings WHERE name='crystal storage'),
    (SELECT id FROM resources WHERE name='crystal'),
    500
  );

INSERT INTO public.buildings_costs ("building", "res", "cost")
  VALUES(
    (SELECT id FROM buildings WHERE name='deuterium tank'),
    (SELECT id FROM resources WHERE name='metal'),
    1000
  );
INSERT INTO public.buildings_costs ("building", "res", "cost")
  VALUES(
    (SELECT id FROM buildings WHERE name='deuterium tank'),
    (SELECT id FROM resources WHERE name='crystal'),
    1000
  );

-- Power plants.
INSERT INTO public.buildings_costs ("building", "res", "cost")
  VALUES(
    (SELECT id FROM buildings WHERE name='solar plant'),
    (SELECT id FROM resources WHERE name='metal'),
    75
  );
INSERT INTO public.buildings_costs ("building", "res", "cost")
  VALUES(
    (SELECT id FROM buildings WHERE name='solar plant'),
    (SELECT id FROM resources WHERE name='crystal'),
    30
  );

INSERT INTO public.buildings_costs ("building", "res", "cost")
  VALUES(
    (SELECT id FROM buildings WHERE name='fusion reactor'),
    (SELECT id FROM resources WHERE name='metal'),
    900
  );
INSERT INTO public.buildings_costs ("building", "res", "cost")
  VALUES(
    (SELECT id FROM buildings WHERE name='fusion reactor'),
    (SELECT id FROM resources WHERE name='crystal'),
    360
  );
INSERT INTO public.buildings_costs ("building", "res", "cost")
  VALUES(
    (SELECT id FROM buildings WHERE name='fusion reactor'),
    (SELECT id FROM resources WHERE name='deuterium'),
    180
  );

-- General buildings.
INSERT INTO public.buildings_costs ("building", "res", "cost")
  VALUES(
    (SELECT id FROM buildings WHERE name='robotics factory'),
    (SELECT id FROM resources WHERE name='metal'),
    400
  );
INSERT INTO public.buildings_costs ("building", "res", "cost")
  VALUES(
    (SELECT id FROM buildings WHERE name='robotics factory'),
    (SELECT id FROM resources WHERE name='crystal'),
    120
  );
INSERT INTO public.buildings_costs ("building", "res", "cost")
  VALUES(
    (SELECT id FROM buildings WHERE name='robotics factory'),
    (SELECT id FROM resources WHERE name='deuterium'),
    200
  );

INSERT INTO public.buildings_costs ("building", "res", "cost")
  VALUES(
    (SELECT id FROM buildings WHERE name='shipyard'),
    (SELECT id FROM resources WHERE name='metal'),
    400
  );
INSERT INTO public.buildings_costs ("building", "res", "cost")
  VALUES(
    (SELECT id FROM buildings WHERE name='shipyard'),
    (SELECT id FROM resources WHERE name='crystal'),
    200
  );
INSERT INTO public.buildings_costs ("building", "res", "cost")
  VALUES(
    (SELECT id FROM buildings WHERE name='shipyard'),
    (SELECT id FROM resources WHERE name='deuterium'),
    100
  );

INSERT INTO public.buildings_costs ("building", "res", "cost")
  VALUES(
    (SELECT id FROM buildings WHERE name='research lab'),
    (SELECT id FROM resources WHERE name='metal'),
    200
  );
INSERT INTO public.buildings_costs ("building", "res", "cost")
  VALUES(
    (SELECT id FROM buildings WHERE name='research lab'),
    (SELECT id FROM resources WHERE name='crystal'),
    400
  );
INSERT INTO public.buildings_costs ("building", "res", "cost")
  VALUES(
    (SELECT id FROM buildings WHERE name='research lab'),
    (SELECT id FROM resources WHERE name='deuterium'),
    200
  );

INSERT INTO public.buildings_costs ("building", "res", "cost")
  VALUES(
    (SELECT id FROM buildings WHERE name='alliance depot'),
    (SELECT id FROM resources WHERE name='metal'),
    20000
  );
INSERT INTO public.buildings_costs ("building", "res", "cost")
  VALUES(
    (SELECT id FROM buildings WHERE name='alliance depot'),
    (SELECT id FROM resources WHERE name='crystal'),
    40000
  );

INSERT INTO public.buildings_costs ("building", "res", "cost")
  VALUES(
    (SELECT id FROM buildings WHERE name='missile silo'),
    (SELECT id FROM resources WHERE name='metal'),
    20000
  );
INSERT INTO public.buildings_costs ("building", "res", "cost")
  VALUES(
    (SELECT id FROM buildings WHERE name='missile silo'),
    (SELECT id FROM resources WHERE name='crystal'),
    20000
  );
INSERT INTO public.buildings_costs ("building", "res", "cost")
  VALUES(
    (SELECT id FROM buildings WHERE name='missile silo'),
    (SELECT id FROM resources WHERE name='deuterium'),
    1000
  );

INSERT INTO public.buildings_costs ("building", "res", "cost")
  VALUES(
    (SELECT id FROM buildings WHERE name='nanite factory'),
    (SELECT id FROM resources WHERE name='metal'),
    1000000
  );
INSERT INTO public.buildings_costs ("building", "res", "cost")
  VALUES(
    (SELECT id FROM buildings WHERE name='nanite factory'),
    (SELECT id FROM resources WHERE name='crystal'),
    500000
  );
INSERT INTO public.buildings_costs ("building", "res", "cost")
  VALUES(
    (SELECT id FROM buildings WHERE name='nanite factory'),
    (SELECT id FROM resources WHERE name='deuterium'),
    100000
  );

INSERT INTO public.buildings_costs ("building", "res", "cost")
  VALUES(
    (SELECT id FROM buildings WHERE name='terraformer'),
    (SELECT id FROM resources WHERE name='crystal'),
    50000
  );
INSERT INTO public.buildings_costs ("building", "res", "cost")
  VALUES(
    (SELECT id FROM buildings WHERE name='terraformer'),
    (SELECT id FROM resources WHERE name='deuterium'),
    100000
  );
INSERT INTO public.buildings_costs ("building", "res", "cost")
  VALUES(
    (SELECT id FROM buildings WHERE name='terraformer'),
    (SELECT id FROM resources WHERE name='energy'),
    1000
  );

INSERT INTO public.buildings_costs ("building", "res", "cost")
  VALUES(
    (SELECT id FROM buildings WHERE name='space dock'),
    (SELECT id FROM resources WHERE name='metal'),
    200
  );
INSERT INTO public.buildings_costs ("building", "res", "cost")
  VALUES(
    (SELECT id FROM buildings WHERE name='space dock'),
    (SELECT id FROM resources WHERE name='deuterium'),
    50
  );
INSERT INTO public.buildings_costs ("building", "res", "cost")
  VALUES(
    (SELECT id FROM buildings WHERE name='space dock'),
    (SELECT id FROM resources WHERE name='energy'),
    50
  );

-- Moon facilities.
INSERT INTO public.buildings_costs ("building", "res", "cost")
  VALUES(
    (SELECT id FROM buildings WHERE name='moon base'),
    (SELECT id FROM resources WHERE name='metal'),
    20000
  );
INSERT INTO public.buildings_costs ("building", "res", "cost")
  VALUES(
    (SELECT id FROM buildings WHERE name='moon base'),
    (SELECT id FROM resources WHERE name='crystal'),
    40000
  );
INSERT INTO public.buildings_costs ("building", "res", "cost")
  VALUES(
    (SELECT id FROM buildings WHERE name='moon base'),
    (SELECT id FROM resources WHERE name='deuterium'),
    20000
  );

INSERT INTO public.buildings_costs ("building", "res", "cost")
  VALUES(
    (SELECT id FROM buildings WHERE name='jump gate'),
    (SELECT id FROM resources WHERE name='metal'),
    2000000
  );
INSERT INTO public.buildings_costs ("building", "res", "cost")
  VALUES(
    (SELECT id FROM buildings WHERE name='jump gate'),
    (SELECT id FROM resources WHERE name='crystal'),
    4000000
  );
INSERT INTO public.buildings_costs ("building", "res", "cost")
  VALUES(
    (SELECT id FROM buildings WHERE name='jump gate'),
    (SELECT id FROM resources WHERE name='deuterium'),
    2000000
  );

INSERT INTO public.buildings_costs ("building", "res", "cost")
  VALUES(
    (SELECT id FROM buildings WHERE name='sensor phalanx'),
    (SELECT id FROM resources WHERE name='metal'),
    20000
  );
INSERT INTO public.buildings_costs ("building", "res", "cost")
  VALUES(
    (SELECT id FROM buildings WHERE name='sensor phalanx'),
    (SELECT id FROM resources WHERE name='crystal'),
    40000
  );
INSERT INTO public.buildings_costs ("building", "res", "cost")
  VALUES(
    (SELECT id FROM buildings WHERE name='sensor phalanx'),
    (SELECT id FROM resources WHERE name='deuterium'),
    20000
  );

-- Seed the building costs progress.
-- Mines.
INSERT INTO public.buildings_costs_progress ("building", "progress")
  VALUES(
    (SELECT id FROM buildings WHERE name='metal mine'),
    1.5
  );
INSERT INTO public.buildings_costs_progress ("building", "progress")
  VALUES(
    (SELECT id FROM buildings WHERE name='crystal mine'),
    1.6
  );
INSERT INTO public.buildings_costs_progress ("building", "progress")
  VALUES(
    (SELECT id FROM buildings WHERE name='deuterium synthetizer'),
    1.5
  );

-- Storages.
INSERT INTO public.buildings_costs_progress ("building", "progress")
  VALUES(
    (SELECT id FROM buildings WHERE name='metal storage'),
    2
  );
INSERT INTO public.buildings_costs_progress ("building", "progress")
  VALUES(
    (SELECT id FROM buildings WHERE name='crystal storage'),
    2
  );
INSERT INTO public.buildings_costs_progress ("building", "progress")
  VALUES(
    (SELECT id FROM buildings WHERE name='deuterium tank'),
    2
  );

-- Power plants.
INSERT INTO public.buildings_costs_progress ("building", "progress")
  VALUES(
    (SELECT id FROM buildings WHERE name='solar plant'),
    1.5
  );
INSERT INTO public.buildings_costs_progress ("building", "progress")
  VALUES(
    (SELECT id FROM buildings WHERE name='fusion reactor'),
    1.8
  );

-- General buildings.
INSERT INTO public.buildings_costs_progress ("building", "progress")
  VALUES(
    (SELECT id FROM buildings WHERE name='robotics factory'),
    2
  );
INSERT INTO public.buildings_costs_progress ("building", "progress")
  VALUES(
    (SELECT id FROM buildings WHERE name='shipyard'),
    2
  );
INSERT INTO public.buildings_costs_progress ("building", "progress")
  VALUES(
    (SELECT id FROM buildings WHERE name='research lab'),
    2
  );
INSERT INTO public.buildings_costs_progress ("building", "progress")
  VALUES(
    (SELECT id FROM buildings WHERE name='alliance depot'),
    2
  );
INSERT INTO public.buildings_costs_progress ("building", "progress")
  VALUES(
    (SELECT id FROM buildings WHERE name='missile silo'),
    2
  );
INSERT INTO public.buildings_costs_progress ("building", "progress")
  VALUES(
    (SELECT id FROM buildings WHERE name='nanite factory'),
    2
  );
INSERT INTO public.buildings_costs_progress ("building", "progress")
  VALUES(
    (SELECT id FROM buildings WHERE name='terraformer'),
    2
  );
INSERT INTO public.buildings_costs_progress ("building", "progress")
  VALUES(
    (SELECT id FROM buildings WHERE name='space dock'),
    2
  );

-- Moon facilities.
INSERT INTO public.buildings_costs_progress ("building", "progress")
  VALUES(
    (SELECT id FROM buildings WHERE name='moon base'),
    2
  );
INSERT INTO public.buildings_costs_progress ("building", "progress")
  VALUES(
    (SELECT id FROM buildings WHERE name='jump gate'),
    2
  );
INSERT INTO public.buildings_costs_progress ("building", "progress")
  VALUES(
    (SELECT id FROM buildings WHERE name='sensor phalanx'),
    2
  );

-- Seed the building gains progress.
INSERT INTO public.buildings_gains_progress ("building", "res", "base", "progress", "temperature_coeff", "temperature_offset")
  VALUES(
    (SELECT id FROM buildings WHERE name='metal mine'),
    (SELECT id FROM resources WHERE name='metal'),
    30,
    1.1,
    0.0,
    1.0
  );
INSERT INTO public.buildings_gains_progress ("building", "res", "base", "progress", "temperature_coeff", "temperature_offset")
  VALUES(
    (SELECT id FROM buildings WHERE name='metal mine'),
    (SELECT id FROM resources WHERE name='energy'),
    -10,
    1.1,
    0.0,
    1.0
  );

INSERT INTO public.buildings_gains_progress ("building", "res", "base", "progress", "temperature_coeff", "temperature_offset")
  VALUES(
    (SELECT id FROM buildings WHERE name='crystal mine'),
    (SELECT id FROM resources WHERE name='crystal'),
    20,
    1.1,
    0.0,
    1.0
  );
INSERT INTO public.buildings_gains_progress ("building", "res", "base", "progress", "temperature_coeff", "temperature_offset")
  VALUES(
    (SELECT id FROM buildings WHERE name='crystal mine'),
    (SELECT id FROM resources WHERE name='energy'),
    -10,
    1.1,
    0.0,
    1.0
  );

INSERT INTO public.buildings_gains_progress ("building", "res", "base", "progress", "temperature_coeff", "temperature_offset")
  VALUES(
    (SELECT id FROM buildings WHERE name='deuterium synthetizer'),
    (SELECT id FROM resources WHERE name='deuterium'),
    10,
    1.1,
    -0.004,
    1.44
  );
INSERT INTO public.buildings_gains_progress ("building", "res", "base", "progress", "temperature_coeff", "temperature_offset")
  VALUES(
    (SELECT id FROM buildings WHERE name='deuterium synthetizer'),
    (SELECT id FROM resources WHERE name='energy'),
    -20,
    1.1,
    0.0,
    1.0
  );

INSERT INTO public.buildings_gains_progress ("building", "res", "base", "progress", "temperature_coeff", "temperature_offset")
  VALUES(
    (SELECT id FROM buildings WHERE name='solar plant'),
    (SELECT id FROM resources WHERE name='energy'),
    20,
    1.1,
    0.0,
    1.0
  );

INSERT INTO public.buildings_gains_progress ("building", "res", "base", "progress", "temperature_coeff", "temperature_offset")
  VALUES(
    (SELECT id FROM buildings WHERE name='fusion reactor'),
    (SELECT id FROM resources WHERE name='energy'),
    30,
    1.05,
    0.0,
    1.0
  );
INSERT INTO public.buildings_gains_progress ("building", "res", "base", "progress", "temperature_coeff", "temperature_offset")
  VALUES(
    (SELECT id FROM buildings WHERE name='fusion reactor'),
    (SELECT id FROM resources WHERE name='deuterium'),
    -10,
    1.1,
    0.0,
    1.0
  );

-- Seed the building storage progress.
INSERT INTO public.buildings_gains_progress ("building", "res", "base", "progress", "temperature_coeff", "temperature_offset")
  VALUES(
    (SELECT id FROM buildings WHERE name='fusion reactor'),
    (SELECT id FROM resources WHERE name='deuterium'),
    -10,
    1.1,
    0.0,
    1.0
  );

INSERT INTO public.buildings_storage_progress ("building", "base", "multiplier", "progress")
  VALUES(
    (SELECT id FROM buildings WHERE name='metal storage'),
    5000,
    2.5,
    0.606060606
    -- Corresponds to 20/33.
  );

INSERT INTO public.buildings_storage_progress ("building", "base", "multiplier", "progress")

  VALUES(
    (SELECT id FROM buildings WHERE name='crystal storage'),
    5000,
    2.5,
    0.606060606
    -- Corresponds to 20/33.
  );

INSERT INTO public.buildings_storage_progress ("building", "base", "multiplier", "progress")
  VALUES(
    (SELECT id FROM buildings WHERE name='deuterium tank'),
    5000,
    2.5,
    0.606060606
  );
