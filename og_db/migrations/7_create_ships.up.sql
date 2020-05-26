
-- Create the table defining ships.
CREATE TABLE ships (
  id uuid NOT NULL DEFAULT uuid_generate_v4(),
  name text NOT NULL,
  cargo integer NOT NULL,
  shield integer NOT NULL,
  weapon integer NOT NULL,
  PRIMARY KEY (id),
  UNIQUE (name)
);

-- Create the table defining the cost of a ship.
CREATE TABLE ships_costs (
  element uuid NOT NULL,
  res uuid NOT NULL,
  cost integer NOT NULL,
  FOREIGN KEY (element) REFERENCES ships(id),
  FOREIGN KEY (res) REFERENCES resources(id),
  UNIQUE (element, res)
);

-- Create the table referencing the propulsion system for ships.
CREATE TABLE ships_propulsion (
  ship uuid NOT NULL,
  propulsion uuid NOT NULL,
  speed integer NOT NULL,
  min_level integer NOT NULL,
  rank integer NOT NULL,
  FOREIGN KEY (ship) REFERENCES ships(id),
  FOREIGN KEY (propulsion) REFERENCES technologies(id),
  UNIQUE (ship, propulsion),
  UNIQUE (ship, rank)
);

-- Create the table representing the increase in propulsion speed for
-- various propulsion technologies.
CREATE TABLE ships_propulsion_increase (
  propulsion uuid,
  increase integer NOT NULL,
  FOREIGN KEY (propulsion) REFERENCES technologies(id),
  UNIQUE (propulsion)
);

-- Create the table defining the consumption of fuel for each ship.
CREATE TABLE ships_propulsion_cost (
  ship uuid NOT NULL,
  res uuid NOT NULL,
  amount integer NOT NULL,
  FOREIGN KEY (ship) REFERENCES ships(id),
  FOREIGN KEY (res) REFERENCES resources(id),
  UNIQUE (ship, res)
);

-- Create the table defining the rapid fire between each ship and any
-- other ship.
CREATE TABLE ships_rapid_fire (
  ship uuid NOT NULL,
  target uuid NOT NULL,
  rapid_fire integer NOT NULL,
  FOREIGN KEY (ship) REFERENCES ships(id),
  FOREIGN KEY (target) REFERENCES ships(id),
  UNIQUE (ship, target)
);

-- Create the table defining the rapid fire between ships and any
-- defense system.
CREATE TABLE ships_rapid_fire_defenses (
  ship uuid NOT NULL,
  target uuid NOT NULL,
  rapid_fire integer NOT NULL,
  FOREIGN KEY (ship) REFERENCES ships(id),
  FOREIGN KEY (target) REFERENCES defenses(id),
  UNIQUE (ship, target)
);

-- Seed the available ships.
INSERT INTO public.ships ("name", "cargo", "shield", "weapon")
  VALUES(
    'small cargo ship',
    5000,
    10,
    5
  );

INSERT INTO public.ships ("name", "cargo", "shield", "weapon")
  VALUES(
    'large cargo ship',
    25000,
    25,
    5
  );

INSERT INTO public.ships ("name", "cargo", "shield", "weapon")
  VALUES(
    'light fighter',
    50,
    10,
    50
  );

INSERT INTO public.ships ("name", "cargo", "shield", "weapon")
  VALUES(
    'heavy fighter',
    100,
    25,
    150
  );

INSERT INTO public.ships ("name", "cargo", "shield", "weapon")
  VALUES(
    'cruiser',
    800,
    50,
    400
  );

INSERT INTO public.ships ("name", "cargo", "shield", "weapon")
  VALUES(
    'battleship',
    1500,
    200,
    1000
  );

INSERT INTO public.ships ("name", "cargo", "shield", "weapon")
  VALUES(
    'battlecruiser',
    750,
    400,
    700
  );

INSERT INTO public.ships ("name", "cargo", "shield", "weapon")
  VALUES(
    'bomber',
    500,
    500,
    1000
  );

INSERT INTO public.ships ("name", "cargo", "shield", "weapon")
  VALUES(
    'destroyer',
    2000,
    500,
    2000
  );

INSERT INTO public.ships ("name", "cargo", "shield", "weapon")
  VALUES(
    'deathstar',
    1000000,
    50000,
    200000
  );

INSERT INTO public.ships ("name", "cargo", "shield", "weapon")
  VALUES(
    'recycler',
    20000,
    10,
    1
  );

INSERT INTO public.ships ("name", "cargo", "shield", "weapon")
  VALUES(
    'espionage probe',
    5,
    0.01,
    0.01
  );

INSERT INTO public.ships ("name", "cargo", "shield", "weapon")
  VALUES(
    'solar satellite',
    0,
    1,
    1
  );

INSERT INTO public.ships ("name", "cargo", "shield", "weapon")
  VALUES(
    'colony ship',
    7500,
    100,
    50
  );

-- Seed the ships costs.
INSERT INTO public.ships_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM ships WHERE name='small cargo ship'),
    (SELECT id FROM resources WHERE name='metal'),
    2000
  );
INSERT INTO public.ships_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM ships WHERE name='small cargo ship'),
    (SELECT id FROM resources WHERE name='crystal'),
    2000
  );

INSERT INTO public.ships_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM ships WHERE name='large cargo ship'),
    (SELECT id FROM resources WHERE name='metal'),
    6000
  );
INSERT INTO public.ships_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM ships WHERE name='large cargo ship'),
    (SELECT id FROM resources WHERE name='crystal'),
    6000
  );

INSERT INTO public.ships_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM ships WHERE name='light fighter'),
    (SELECT id FROM resources WHERE name='metal'),
    3000
  );
INSERT INTO public.ships_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM ships WHERE name='light fighter'),
    (SELECT id FROM resources WHERE name='crystal'),
    1000
  );

INSERT INTO public.ships_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM ships WHERE name='heavy fighter'),
    (SELECT id FROM resources WHERE name='metal'),
    6000
  );
INSERT INTO public.ships_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM ships WHERE name='heavy fighter'),
    (SELECT id FROM resources WHERE name='crystal'),
    4000
  );

INSERT INTO public.ships_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM ships WHERE name='cruiser'),
    (SELECT id FROM resources WHERE name='metal'),
    20000
  );
INSERT INTO public.ships_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM ships WHERE name='cruiser'),
    (SELECT id FROM resources WHERE name='crystal'),
    7000
  );
INSERT INTO public.ships_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM ships WHERE name='cruiser'),
    (SELECT id FROM resources WHERE name='deuterium'),
    2000
  );

INSERT INTO public.ships_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM ships WHERE name='battleship'),
    (SELECT id FROM resources WHERE name='metal'),
    45000
  );
INSERT INTO public.ships_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM ships WHERE name='battleship'),
    (SELECT id FROM resources WHERE name='crystal'),
    15000
  );

INSERT INTO public.ships_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM ships WHERE name='battlecruiser'),
    (SELECT id FROM resources WHERE name='metal'),
    30000
  );
INSERT INTO public.ships_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM ships WHERE name='battlecruiser'),
    (SELECT id FROM resources WHERE name='crystal'),
    40000
  );
INSERT INTO public.ships_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM ships WHERE name='battlecruiser'),
    (SELECT id FROM resources WHERE name='deuterium'),
    15000
  );

INSERT INTO public.ships_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM ships WHERE name='bomber'),
    (SELECT id FROM resources WHERE name='metal'),
    50000
  );
INSERT INTO public.ships_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM ships WHERE name='bomber'),
    (SELECT id FROM resources WHERE name='crystal'),
    25000
  );
INSERT INTO public.ships_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM ships WHERE name='bomber'),
    (SELECT id FROM resources WHERE name='deuterium'),
    15000
  );

INSERT INTO public.ships_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM ships WHERE name='destroyer'),
    (SELECT id FROM resources WHERE name='metal'),
    60000
  );
INSERT INTO public.ships_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM ships WHERE name='destroyer'),
    (SELECT id FROM resources WHERE name='crystal'),
    50000
  );
INSERT INTO public.ships_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM ships WHERE name='destroyer'),
    (SELECT id FROM resources WHERE name='deuterium'),
    15000
  );

INSERT INTO public.ships_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM ships WHERE name='deathstar'),
    (SELECT id FROM resources WHERE name='metal'),
    5000000
  );
INSERT INTO public.ships_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM ships WHERE name='deathstar'),
    (SELECT id FROM resources WHERE name='crystal'),
    4000000
  );
INSERT INTO public.ships_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM ships WHERE name='deathstar'),
    (SELECT id FROM resources WHERE name='deuterium'),
    1000000
  );

INSERT INTO public.ships_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM ships WHERE name='recycler'),
    (SELECT id FROM resources WHERE name='metal'),
    10000
  );
INSERT INTO public.ships_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM ships WHERE name='recycler'),
    (SELECT id FROM resources WHERE name='crystal'),
    6000
  );
INSERT INTO public.ships_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM ships WHERE name='recycler'),
    (SELECT id FROM resources WHERE name='deuterium'),
    2000
  );

INSERT INTO public.ships_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM ships WHERE name='espionage probe'),
    (SELECT id FROM resources WHERE name='crystal'),
    1000
  );

INSERT INTO public.ships_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM ships WHERE name='solar satellite'),
    (SELECT id FROM resources WHERE name='crystal'),
    2000
  );
INSERT INTO public.ships_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM ships WHERE name='solar satellite'),
    (SELECT id FROM resources WHERE name='deuterium'),
    500
  );

INSERT INTO public.ships_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM ships WHERE name='colony ship'),
    (SELECT id FROM resources WHERE name='metal'),
    10000
  );
INSERT INTO public.ships_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM ships WHERE name='colony ship'),
    (SELECT id FROM resources WHERE name='crystal'),
    20000
  );
INSERT INTO public.ships_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM ships WHERE name='colony ship'),
    (SELECT id FROM resources WHERE name='deuterium'),
    10000
  );

-- Seed the ships propulsion.
INSERT INTO public.ships_propulsion ("ship", "propulsion", "speed", "min_level", "rank")
  VALUES(
    (SELECT id FROM ships WHERE name='small cargo ship'),
    (SELECT id FROM technologies WHERE name='combustion drive'),
    5000,
    0,
    0
  );
INSERT INTO public.ships_propulsion ("ship", "propulsion", "speed", "min_level", "rank")
  VALUES(
    (SELECT id FROM ships WHERE name='small cargo ship'),
    (SELECT id FROM technologies WHERE name='impulse drive'),
    10000,
    4,
    1
  );

INSERT INTO public.ships_propulsion ("ship", "propulsion", "speed", "min_level", "rank")
  VALUES(
    (SELECT id FROM ships WHERE name='large cargo ship'),
    (SELECT id FROM technologies WHERE name='combustion drive'),
    7500,
    0,
    0
  );

INSERT INTO public.ships_propulsion ("ship", "propulsion", "speed", "min_level", "rank")
  VALUES(
    (SELECT id FROM ships WHERE name='light fighter'),
    (SELECT id FROM technologies WHERE name='combustion drive'),
    12500,
    0,
    0
  );

INSERT INTO public.ships_propulsion ("ship", "propulsion", "speed", "min_level", "rank")
  VALUES(
    (SELECT id FROM ships WHERE name='heavy fighter'),
    (SELECT id FROM technologies WHERE name='impulse drive'),
    10000,
    0,
    0
  );

INSERT INTO public.ships_propulsion ("ship", "propulsion", "speed", "min_level", "rank")
  VALUES(
    (SELECT id FROM ships WHERE name='cruiser'),
    (SELECT id FROM technologies WHERE name='impulse drive'),
    15000,
    0,
    0
  );

INSERT INTO public.ships_propulsion ("ship", "propulsion", "speed", "min_level", "rank")
  VALUES(
    (SELECT id FROM ships WHERE name='battleship'),
    (SELECT id FROM technologies WHERE name='hyperspace drive'),
    10000,
    0,
    0
  );

INSERT INTO public.ships_propulsion ("ship", "propulsion", "speed", "min_level", "rank")
  VALUES(
    (SELECT id FROM ships WHERE name='battlecruiser'),
    (SELECT id FROM technologies WHERE name='hyperspace drive'),
    10000,
    0,
    0
  );

INSERT INTO public.ships_propulsion ("ship", "propulsion", "speed", "min_level", "rank")
  VALUES(
    (SELECT id FROM ships WHERE name='bomber'),
    (SELECT id FROM technologies WHERE name='impulse drive'),
    4000,
    0,
    0
  );
INSERT INTO public.ships_propulsion ("ship", "propulsion", "speed", "min_level", "rank")
  VALUES(
    (SELECT id FROM ships WHERE name='bomber'),
    (SELECT id FROM technologies WHERE name='hyperspace drive'),
    5000,
    7,
    1
  );

INSERT INTO public.ships_propulsion ("ship", "propulsion", "speed", "min_level", "rank")
  VALUES(
    (SELECT id FROM ships WHERE name='destroyer'),
    (SELECT id FROM technologies WHERE name='hyperspace drive'),
    5000,
    0,
    0
  );

INSERT INTO public.ships_propulsion ("ship", "propulsion", "speed", "min_level", "rank")
  VALUES(
    (SELECT id FROM ships WHERE name='deathstar'),
    (SELECT id FROM technologies WHERE name='hyperspace drive'),
    100,
    0,
    0
  );

INSERT INTO public.ships_propulsion ("ship", "propulsion", "speed", "min_level", "rank")
  VALUES(
    (SELECT id FROM ships WHERE name='recycler'),
    (SELECT id FROM technologies WHERE name='combustion drive'),
    2000,
    0,
    0
  );
INSERT INTO public.ships_propulsion ("ship", "propulsion", "speed", "min_level", "rank")
  VALUES(
    (SELECT id FROM ships WHERE name='recycler'),
    (SELECT id FROM technologies WHERE name='impulse drive'),
    4000,
    16,
    1
  );
INSERT INTO public.ships_propulsion ("ship", "propulsion", "speed", "min_level", "rank")
  VALUES(
    (SELECT id FROM ships WHERE name='recycler'),
    (SELECT id FROM technologies WHERE name='hyperspace drive'),
    6000,
    14,
    2
  );

INSERT INTO public.ships_propulsion ("ship", "propulsion", "speed", "min_level", "rank")
  VALUES(
    (SELECT id FROM ships WHERE name='espionage probe'),
    (SELECT id FROM technologies WHERE name='combustion drive'),
    100000000,
    0,
    0
  );

INSERT INTO public.ships_propulsion ("ship", "propulsion", "speed", "min_level", "rank")
  VALUES(
    (SELECT id FROM ships WHERE name='solar satellite'),
    (SELECT id FROM technologies WHERE name='combustion drive'),
    0,
    0,
    0
  );

INSERT INTO public.ships_propulsion ("ship", "propulsion", "speed", "min_level", "rank")
  VALUES(
    (SELECT id FROM ships WHERE name='colony ship'),
    (SELECT id FROM technologies WHERE name='impulse drive'),
    2500,
    0,
    0
  );

-- Seed the ships propulsion increase.
INSERT INTO public.ships_propulsion_increase ("propulsion", "increase")
  VALUES(
    (SELECT id FROM technologies WHERE name='combustion drive'),
    10
  );

INSERT INTO public.ships_propulsion_increase ("propulsion", "increase")
  VALUES(
    (SELECT id FROM technologies WHERE name='impulse drive'),
    20
  );

INSERT INTO public.ships_propulsion_increase ("propulsion", "increase")
  VALUES(
    (SELECT id FROM technologies WHERE name='hyperspace drive'),
    30
  );

-- Seed the ships propulsion cost.
INSERT INTO public.ships_propulsion_cost ("ship", "res", "amount")
  VALUES(
    (SELECT id FROM ships WHERE name='small cargo ship'),
    (SELECT id FROM resources WHERE name='deuterium'),
    10
  );

INSERT INTO public.ships_propulsion_cost ("ship", "res", "amount")
  VALUES(
    (SELECT id FROM ships WHERE name='large cargo ship'),
    (SELECT id FROM resources WHERE name='deuterium'),
    50
  );

INSERT INTO public.ships_propulsion_cost ("ship", "res", "amount")
  VALUES(
    (SELECT id FROM ships WHERE name='light fighter'),
    (SELECT id FROM resources WHERE name='deuterium'),
    20
  );

INSERT INTO public.ships_propulsion_cost ("ship", "res", "amount")
  VALUES(
    (SELECT id FROM ships WHERE name='heavy fighter'),
    (SELECT id FROM resources WHERE name='deuterium'),
    75
  );

INSERT INTO public.ships_propulsion_cost ("ship", "res", "amount")
  VALUES(
    (SELECT id FROM ships WHERE name='cruiser'),
    (SELECT id FROM resources WHERE name='deuterium'),
    300
  );

INSERT INTO public.ships_propulsion_cost ("ship", "res", "amount")
  VALUES(
    (SELECT id FROM ships WHERE name='battleship'),
    (SELECT id FROM resources WHERE name='deuterium'),
    500
  );

INSERT INTO public.ships_propulsion_cost ("ship", "res", "amount")
  VALUES(
    (SELECT id FROM ships WHERE name='battlecruiser'),
    (SELECT id FROM resources WHERE name='deuterium'),
    250
  );

INSERT INTO public.ships_propulsion_cost ("ship", "res", "amount")
  VALUES(
    (SELECT id FROM ships WHERE name='bomber'),
    (SELECT id FROM resources WHERE name='deuterium'),
    700
  );

INSERT INTO public.ships_propulsion_cost ("ship", "res", "amount")
  VALUES(
    (SELECT id FROM ships WHERE name='destroyer'),
    (SELECT id FROM resources WHERE name='deuterium'),
    1000
  );

INSERT INTO public.ships_propulsion_cost ("ship", "res", "amount")
  VALUES(
    (SELECT id FROM ships WHERE name='deathstar'),
    (SELECT id FROM resources WHERE name='deuterium'),
    1
  );

INSERT INTO public.ships_propulsion_cost ("ship", "res", "amount")
  VALUES(
    (SELECT id FROM ships WHERE name='recycler'),
    (SELECT id FROM resources WHERE name='deuterium'),
    300
  );

INSERT INTO public.ships_propulsion_cost ("ship", "res", "amount")
  VALUES(
    (SELECT id FROM ships WHERE name='espionage probe'),
    (SELECT id FROM resources WHERE name='deuterium'),
    1
  );

INSERT INTO public.ships_propulsion_cost ("ship", "res", "amount")
  VALUES(
    (SELECT id FROM ships WHERE name='solar satellite'),
    (SELECT id FROM resources WHERE name='deuterium'),
    0
  );

INSERT INTO public.ships_propulsion_cost ("ship", "res", "amount")
  VALUES(
    (SELECT id FROM ships WHERE name='colony ship'),
    (SELECT id FROM resources WHERE name='deuterium'),
    1000
  );

-- Seed the ships rapid fire against other ships.
-- Small cargo ship.
INSERT INTO public.ships_rapid_fire ("ship", "target", "rapid_fire")
  VALUES(
    (SELECT id FROM ships WHERE name='small cargo ship'),
    (SELECT id FROM ships WHERE name='espionage probe'),
    5
  );

INSERT INTO public.ships_rapid_fire ("ship", "target", "rapid_fire")
  VALUES(
    (SELECT id FROM ships WHERE name='small cargo ship'),
    (SELECT id FROM ships WHERE name='solar satellite'),
    5
  );

-- Large cargo ship.
INSERT INTO public.ships_rapid_fire ("ship", "target", "rapid_fire")
  VALUES(
    (SELECT id FROM ships WHERE name='large cargo ship'),
    (SELECT id FROM ships WHERE name='espionage probe'),
    5
  );

INSERT INTO public.ships_rapid_fire ("ship", "target", "rapid_fire")
  VALUES(
    (SELECT id FROM ships WHERE name='large cargo ship'),
    (SELECT id FROM ships WHERE name='solar satellite'),
    5
  );

-- Light fighter.
INSERT INTO public.ships_rapid_fire ("ship", "target", "rapid_fire")
  VALUES(
    (SELECT id FROM ships WHERE name='light fighter'),
    (SELECT id FROM ships WHERE name='espionage probe'),
    5
  );

INSERT INTO public.ships_rapid_fire ("ship", "target", "rapid_fire")
  VALUES(
    (SELECT id FROM ships WHERE name='light fighter'),
    (SELECT id FROM ships WHERE name='solar satellite'),
    5
  );

-- Heavy fighter.
INSERT INTO public.ships_rapid_fire ("ship", "target", "rapid_fire")
  VALUES(
    (SELECT id FROM ships WHERE name='heavy fighter'),
    (SELECT id FROM ships WHERE name='espionage probe'),
    5
  );

INSERT INTO public.ships_rapid_fire ("ship", "target", "rapid_fire")
  VALUES(
    (SELECT id FROM ships WHERE name='heavy fighter'),
    (SELECT id FROM ships WHERE name='solar satellite'),
    5
  );

INSERT INTO public.ships_rapid_fire ("ship", "target", "rapid_fire")
  VALUES(
    (SELECT id FROM ships WHERE name='heavy fighter'),
    (SELECT id FROM ships WHERE name='small cargo ship'),
    3
  );

-- Cruiser.
INSERT INTO public.ships_rapid_fire ("ship", "target", "rapid_fire")
  VALUES(
    (SELECT id FROM ships WHERE name='cruiser'),
    (SELECT id FROM ships WHERE name='espionage probe'),
    5
  );

INSERT INTO public.ships_rapid_fire ("ship", "target", "rapid_fire")
  VALUES(
    (SELECT id FROM ships WHERE name='cruiser'),
    (SELECT id FROM ships WHERE name='solar satellite'),
    5
  );

INSERT INTO public.ships_rapid_fire ("ship", "target", "rapid_fire")
  VALUES(
    (SELECT id FROM ships WHERE name='cruiser'),
    (SELECT id FROM ships WHERE name='light fighter'),
    6
  );

-- Battleship.
INSERT INTO public.ships_rapid_fire ("ship", "target", "rapid_fire")
  VALUES(
    (SELECT id FROM ships WHERE name='battleship'),
    (SELECT id FROM ships WHERE name='espionage probe'),
    5
  );

INSERT INTO public.ships_rapid_fire ("ship", "target", "rapid_fire")
  VALUES(
    (SELECT id FROM ships WHERE name='battleship'),
    (SELECT id FROM ships WHERE name='solar satellite'),
    5
  );

-- Battlecruiser.
INSERT INTO public.ships_rapid_fire ("ship", "target", "rapid_fire")
  VALUES(
    (SELECT id FROM ships WHERE name='battlecruiser'),
    (SELECT id FROM ships WHERE name='espionage probe'),
    5
  );

INSERT INTO public.ships_rapid_fire ("ship", "target", "rapid_fire")
  VALUES(
    (SELECT id FROM ships WHERE name='battlecruiser'),
    (SELECT id FROM ships WHERE name='solar satellite'),
    5
  );

INSERT INTO public.ships_rapid_fire ("ship", "target", "rapid_fire")
  VALUES(
    (SELECT id FROM ships WHERE name='battlecruiser'),
    (SELECT id FROM ships WHERE name='small cargo ship'),
    3
  );

INSERT INTO public.ships_rapid_fire ("ship", "target", "rapid_fire")
  VALUES(
    (SELECT id FROM ships WHERE name='battlecruiser'),
    (SELECT id FROM ships WHERE name='large cargo ship'),
    3
  );

INSERT INTO public.ships_rapid_fire ("ship", "target", "rapid_fire")
  VALUES(
    (SELECT id FROM ships WHERE name='battlecruiser'),
    (SELECT id FROM ships WHERE name='heavy fighter'),
    4
  );

INSERT INTO public.ships_rapid_fire ("ship", "target", "rapid_fire")
  VALUES(
    (SELECT id FROM ships WHERE name='battlecruiser'),
    (SELECT id FROM ships WHERE name='cruiser'),
    4
  );

INSERT INTO public.ships_rapid_fire ("ship", "target", "rapid_fire")
  VALUES(
    (SELECT id FROM ships WHERE name='battlecruiser'),
    (SELECT id FROM ships WHERE name='battleship'),
    7
  );

-- Bomber.
INSERT INTO public.ships_rapid_fire ("ship", "target", "rapid_fire")
  VALUES(
    (SELECT id FROM ships WHERE name='bomber'),
    (SELECT id FROM ships WHERE name='espionage probe'),
    5
  );

INSERT INTO public.ships_rapid_fire ("ship", "target", "rapid_fire")
  VALUES(
    (SELECT id FROM ships WHERE name='bomber'),
    (SELECT id FROM ships WHERE name='solar satellite'),
    5
  );

-- Destroyer.
INSERT INTO public.ships_rapid_fire ("ship", "target", "rapid_fire")
  VALUES(
    (SELECT id FROM ships WHERE name='destroyer'),
    (SELECT id FROM ships WHERE name='espionage probe'),
    5
  );

INSERT INTO public.ships_rapid_fire ("ship", "target", "rapid_fire")
  VALUES(
    (SELECT id FROM ships WHERE name='destroyer'),
    (SELECT id FROM ships WHERE name='solar satellite'),
    5
  );

INSERT INTO public.ships_rapid_fire ("ship", "target", "rapid_fire")
  VALUES(
    (SELECT id FROM ships WHERE name='destroyer'),
    (SELECT id FROM ships WHERE name='battlecruiser'),
    2
  );

-- Deathstar.
INSERT INTO public.ships_rapid_fire ("ship", "target", "rapid_fire")
  VALUES(
    (SELECT id FROM ships WHERE name='deathstar'),
    (SELECT id FROM ships WHERE name='espionage probe'),
    1250
  );

INSERT INTO public.ships_rapid_fire ("ship", "target", "rapid_fire")
  VALUES(
    (SELECT id FROM ships WHERE name='deathstar'),
    (SELECT id FROM ships WHERE name='solar satellite'),
    1250
  );

INSERT INTO public.ships_rapid_fire ("ship", "target", "rapid_fire")
  VALUES(
    (SELECT id FROM ships WHERE name='deathstar'),
    (SELECT id FROM ships WHERE name='small cargo ship'),
    250
  );

INSERT INTO public.ships_rapid_fire ("ship", "target", "rapid_fire")
  VALUES(
    (SELECT id FROM ships WHERE name='deathstar'),
    (SELECT id FROM ships WHERE name='large cargo ship'),
    250
  );

INSERT INTO public.ships_rapid_fire ("ship", "target", "rapid_fire")
  VALUES(
    (SELECT id FROM ships WHERE name='deathstar'),
    (SELECT id FROM ships WHERE name='light fighter'),
    200
  );

INSERT INTO public.ships_rapid_fire ("ship", "target", "rapid_fire")
  VALUES(
    (SELECT id FROM ships WHERE name='deathstar'),
    (SELECT id FROM ships WHERE name='heavy fighter'),
    100
  );

INSERT INTO public.ships_rapid_fire ("ship", "target", "rapid_fire")
  VALUES(
    (SELECT id FROM ships WHERE name='deathstar'),
    (SELECT id FROM ships WHERE name='cruiser'),
    33
  );

INSERT INTO public.ships_rapid_fire ("ship", "target", "rapid_fire")
  VALUES(
    (SELECT id FROM ships WHERE name='deathstar'),
    (SELECT id FROM ships WHERE name='battleship'),
    30
  );

INSERT INTO public.ships_rapid_fire ("ship", "target", "rapid_fire")
  VALUES(
    (SELECT id FROM ships WHERE name='deathstar'),
    (SELECT id FROM ships WHERE name='battlecruiser'),
    15
  );

INSERT INTO public.ships_rapid_fire ("ship", "target", "rapid_fire")
  VALUES(
    (SELECT id FROM ships WHERE name='deathstar'),
    (SELECT id FROM ships WHERE name='bomber'),
    25
  );

INSERT INTO public.ships_rapid_fire ("ship", "target", "rapid_fire")
  VALUES(
    (SELECT id FROM ships WHERE name='deathstar'),
    (SELECT id FROM ships WHERE name='destroyer'),
    5
  );

INSERT INTO public.ships_rapid_fire ("ship", "target", "rapid_fire")
  VALUES(
    (SELECT id FROM ships WHERE name='deathstar'),
    (SELECT id FROM ships WHERE name='recycler'),
    250
  );

INSERT INTO public.ships_rapid_fire ("ship", "target", "rapid_fire")
  VALUES(
    (SELECT id FROM ships WHERE name='deathstar'),
    (SELECT id FROM ships WHERE name='colony ship'),
    250
  );

-- Recycler.
INSERT INTO public.ships_rapid_fire ("ship", "target", "rapid_fire")
  VALUES(
    (SELECT id FROM ships WHERE name='recycler'),
    (SELECT id FROM ships WHERE name='espionage probe'),
    5
  );

INSERT INTO public.ships_rapid_fire ("ship", "target", "rapid_fire")
  VALUES(
    (SELECT id FROM ships WHERE name='recycler'),
    (SELECT id FROM ships WHERE name='solar satellite'),
    5
  );

-- Colony ship.
INSERT INTO public.ships_rapid_fire ("ship", "target", "rapid_fire")
  VALUES(
    (SELECT id FROM ships WHERE name='colony ship'),
    (SELECT id FROM ships WHERE name='espionage probe'),
    5
  );

INSERT INTO public.ships_rapid_fire ("ship", "target", "rapid_fire")
  VALUES(
    (SELECT id FROM ships WHERE name='colony ship'),
    (SELECT id FROM ships WHERE name='solar satellite'),
    5
  );

-- Seed the ships rapid fire against defenses.
-- Cruiser.
INSERT INTO public.ships_rapid_fire_defenses ("ship", "target", "rapid_fire")
  VALUES(
    (SELECT id FROM ships WHERE name='cruiser'),
    (SELECT id FROM defenses WHERE name='rocket launcher'),
    10
  );

-- Bomber.
INSERT INTO public.ships_rapid_fire_defenses ("ship", "target", "rapid_fire")
  VALUES(
    (SELECT id FROM ships WHERE name='bomber'),
    (SELECT id FROM defenses WHERE name='rocket launcher'),
    20
  );

INSERT INTO public.ships_rapid_fire_defenses ("ship", "target", "rapid_fire")
  VALUES(
    (SELECT id FROM ships WHERE name='bomber'),
    (SELECT id FROM defenses WHERE name='light laser'),
    20
  );

INSERT INTO public.ships_rapid_fire_defenses ("ship", "target", "rapid_fire")
  VALUES(
    (SELECT id FROM ships WHERE name='bomber'),
    (SELECT id FROM defenses WHERE name='heavy laser'),
    10
  );

INSERT INTO public.ships_rapid_fire_defenses ("ship", "target", "rapid_fire")
  VALUES(
    (SELECT id FROM ships WHERE name='bomber'),
    (SELECT id FROM defenses WHERE name='ion cannon'),
    10
  );

-- Destroyer.
INSERT INTO public.ships_rapid_fire_defenses ("ship", "target", "rapid_fire")
  VALUES(
    (SELECT id FROM ships WHERE name='destroyer'),
    (SELECT id FROM defenses WHERE name='light laser'),
    10
  );

-- Deathstar.
INSERT INTO public.ships_rapid_fire_defenses ("ship", "target", "rapid_fire")
  VALUES(
    (SELECT id FROM ships WHERE name='deathstar'),
    (SELECT id FROM defenses WHERE name='rocket launcher'),
    200
  );

INSERT INTO public.ships_rapid_fire_defenses ("ship", "target", "rapid_fire")
  VALUES(
    (SELECT id FROM ships WHERE name='deathstar'),
    (SELECT id FROM defenses WHERE name='light laser'),
    200
  );

INSERT INTO public.ships_rapid_fire_defenses ("ship", "target", "rapid_fire")
  VALUES(
    (SELECT id FROM ships WHERE name='deathstar'),
    (SELECT id FROM defenses WHERE name='heavy laser'),
    100
  );

INSERT INTO public.ships_rapid_fire_defenses ("ship", "target", "rapid_fire")
  VALUES(
    (SELECT id FROM ships WHERE name='deathstar'),
    (SELECT id FROM defenses WHERE name='ion cannon'),
    100
  );

INSERT INTO public.ships_rapid_fire_defenses ("ship", "target", "rapid_fire")
  VALUES(
    (SELECT id FROM ships WHERE name='deathstar'),
    (SELECT id FROM defenses WHERE name='gauss cannon'),
    50
  );
