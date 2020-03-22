
-- Create the table defining ships.
CREATE TABLE ships (
    id uuid NOT NULL DEFAULT uuid_generate_v4(),
    name text,
    propulsion uuid NOT NULL references technologies,
    speed integer NOT NULL,
    cargo integer NOT NULL,
    shield integer NOT NULL,
    weapon integer NOT NULL,
    PRIMARY KEY (id)
);

-- Create the table defining the cost of a ship.
CREATE TABLE ships_costs (
  ship uuid NOT NULL references ships,
  res uuid NOT NULL references resources,
  cost integer NOT NULL
);

-- Create the table representing the increase in propulsion speed for
-- various propulsion technologies.
CREATE TABLE ships_propulsion_increase (
  propulsion uuid references technologies,
  increase integer NOT NULL
);

-- Create the table defining the consumption of fuel for each ship.
CREATE TABLE ships_propulsion_cost (
  ship uuid NOT NULL references ships,
  res uuid NOT NULL references resources,
  amount integer NOT NULL
);

-- Create the table defining the rapid fire between each ship and any
-- other ship.
CREATE TABLE ships_rapid_fire (
  ship uuid NOT NULL references ships,
  target uuid NOT NULL references ships,
  rapid_fire integer NOT NULL
);

-- Create the table defining the rapid fire between ships and any
-- defense system.
CREATE TABLE ships_rapid_fire_defenses (
  ship uuid NOT NULL references ships,
  target uuid NOT NULL references defenses,
  rapid_fire integer NOT NULL
);

-- Seed the available ships.
INSERT INTO public.ships ("name", "propulsion", "speed", "cargo", "shield", "weapon")
  VALUES(
    'small cargo ship',
    (SELECT id FROM technologies WHERE name='combustion drive'),
    5000,
    5000,
    10,
    5
  );

INSERT INTO public.ships ("name", "propulsion", "speed", "cargo", "shield", "weapon")
  VALUES(
    'large cargo ship',
    (SELECT id FROM technologies WHERE name='combustion drive'),
    25000,
    7500,
    25,
    5
  );

INSERT INTO public.ships ("name", "propulsion", "speed", "cargo", "shield", "weapon")
  VALUES(
    'light fighter',
    (SELECT id FROM technologies WHERE name='combustion drive'),
    50,
    12500,
    10,
    50
  );

INSERT INTO public.ships ("name", "propulsion", "speed", "cargo", "shield", "weapon")
  VALUES(
    'heavy fighter',
    (SELECT id FROM technologies WHERE name='impulse drive'),
    100,
    10000,
    25,
    150
  );

INSERT INTO public.ships ("name", "propulsion", "speed", "cargo", "shield", "weapon")
  VALUES(
    'cruiser',
    (SELECT id FROM technologies WHERE name='impulse drive'),
    800,
    15000,
    50,
    400
  );

INSERT INTO public.ships ("name", "propulsion", "speed", "cargo", "shield", "weapon")
  VALUES(
    'battleship',
    (SELECT id FROM technologies WHERE name='hyperspace drive'),
    1500,
    10000,
    200,
    1000
  );

INSERT INTO public.ships ("name", "propulsion", "speed", "cargo", "shield", "weapon")
  VALUES(
    'battlecruiser',
    (SELECT id FROM technologies WHERE name='hyperspace drive'),
    750,
    10000,
    400,
    700
  );

INSERT INTO public.ships ("name", "propulsion", "speed", "cargo", "shield", "weapon")
  VALUES(
    'bomber',
    (SELECT id FROM technologies WHERE name='hyperspace drive'),
    500,
    4000,
    500,
    1000
  );

INSERT INTO public.ships ("name", "propulsion", "speed", "cargo", "shield", "weapon")
  VALUES(
    'destroyer',
    (SELECT id FROM technologies WHERE name='hyperspace drive'),
    2000,
    5000,
    500,
    2000
  );

INSERT INTO public.ships ("name", "propulsion", "speed", "cargo", "shield", "weapon")
  VALUES(
    'deathstar',
    (SELECT id FROM technologies WHERE name='hyperspace drive'),
    1000000,
    100,
    50000,
    200000
  );

INSERT INTO public.ships ("name", "propulsion", "speed", "cargo", "shield", "weapon")
  VALUES(
    'recycler',
    (SELECT id FROM technologies WHERE name='combustion drive'),
    20000,
    2000,
    10,
    1
  );

INSERT INTO public.ships ("name", "propulsion", "speed", "cargo", "shield", "weapon")
  VALUES(
    'espionage probe',
    (SELECT id FROM technologies WHERE name='combustion drive'),
    5,
    100000000,
    0.01,
    0.01
  );

INSERT INTO public.ships ("name", "propulsion", "speed", "cargo", "shield", "weapon")
  VALUES(
    'solar satellite',
    (SELECT id FROM technologies WHERE name='combustion drive'),
    0,
    0,
    1,
    1
  );

INSERT INTO public.ships ("name", "propulsion", "speed", "cargo", "shield", "weapon")
  VALUES(
    'colony ship',
    (SELECT id FROM technologies WHERE name='impulse drive'),
    7500,
    2500,
    100,
    50
  );

-- Seed the ships costs.
INSERT INTO public.ships_costs ("ship", "res", "cost")
  VALUES(
    (SELECT id FROM ships WHERE name='small cargo ship'),
    (SELECT id FROM resources WHERE name='metal'),
    2000
  );
INSERT INTO public.ships_costs ("ship", "res", "cost")
  VALUES(
    (SELECT id FROM ships WHERE name='small cargo ship'),
    (SELECT id FROM resources WHERE name='crystal'),
    2000
  );

INSERT INTO public.ships_costs ("ship", "res", "cost")
  VALUES(
    (SELECT id FROM ships WHERE name='large cargo ship'),
    (SELECT id FROM resources WHERE name='metal'),
    6000
  );
INSERT INTO public.ships_costs ("ship", "res", "cost")
  VALUES(
    (SELECT id FROM ships WHERE name='large cargo ship'),
    (SELECT id FROM resources WHERE name='crystal'),
    6000
  );

INSERT INTO public.ships_costs ("ship", "res", "cost")
  VALUES(
    (SELECT id FROM ships WHERE name='light fighter'),
    (SELECT id FROM resources WHERE name='metal'),
    3000
  );
INSERT INTO public.ships_costs ("ship", "res", "cost")
  VALUES(
    (SELECT id FROM ships WHERE name='light fighter'),
    (SELECT id FROM resources WHERE name='crystal'),
    1000
  );

INSERT INTO public.ships_costs ("ship", "res", "cost")
  VALUES(
    (SELECT id FROM ships WHERE name='heavy fighter'),
    (SELECT id FROM resources WHERE name='metal'),
    6000
  );
INSERT INTO public.ships_costs ("ship", "res", "cost")
  VALUES(
    (SELECT id FROM ships WHERE name='heavy fighter'),
    (SELECT id FROM resources WHERE name='crystal'),
    4000
  );

INSERT INTO public.ships_costs ("ship", "res", "cost")
  VALUES(
    (SELECT id FROM ships WHERE name='cruiser'),
    (SELECT id FROM resources WHERE name='metal'),
    20000
  );
INSERT INTO public.ships_costs ("ship", "res", "cost")
  VALUES(
    (SELECT id FROM ships WHERE name='cruiser'),
    (SELECT id FROM resources WHERE name='crystal'),
    7000
  );
INSERT INTO public.ships_costs ("ship", "res", "cost")
  VALUES(
    (SELECT id FROM ships WHERE name='cruiser'),
    (SELECT id FROM resources WHERE name='deuterium'),
    2000
  );

INSERT INTO public.ships_costs ("ship", "res", "cost")
  VALUES(
    (SELECT id FROM ships WHERE name='battleship'),
    (SELECT id FROM resources WHERE name='metal'),
    45000
  );
INSERT INTO public.ships_costs ("ship", "res", "cost")
  VALUES(
    (SELECT id FROM ships WHERE name='battleship'),
    (SELECT id FROM resources WHERE name='crystal'),
    15000
  );

INSERT INTO public.ships_costs ("ship", "res", "cost")
  VALUES(
    (SELECT id FROM ships WHERE name='battlecruiser'),
    (SELECT id FROM resources WHERE name='metal'),
    30000
  );
INSERT INTO public.ships_costs ("ship", "res", "cost")
  VALUES(
    (SELECT id FROM ships WHERE name='battlecruiser'),
    (SELECT id FROM resources WHERE name='crystal'),
    40000
  );
INSERT INTO public.ships_costs ("ship", "res", "cost")
  VALUES(
    (SELECT id FROM ships WHERE name='battlecruiser'),
    (SELECT id FROM resources WHERE name='deuterium'),
    15000
  );

INSERT INTO public.ships_costs ("ship", "res", "cost")
  VALUES(
    (SELECT id FROM ships WHERE name='bomber'),
    (SELECT id FROM resources WHERE name='metal'),
    50000
  );
INSERT INTO public.ships_costs ("ship", "res", "cost")
  VALUES(
    (SELECT id FROM ships WHERE name='bomber'),
    (SELECT id FROM resources WHERE name='crystal'),
    25000
  );
INSERT INTO public.ships_costs ("ship", "res", "cost")
  VALUES(
    (SELECT id FROM ships WHERE name='bomber'),
    (SELECT id FROM resources WHERE name='deuterium'),
    15000
  );

INSERT INTO public.ships_costs ("ship", "res", "cost")
  VALUES(
    (SELECT id FROM ships WHERE name='destroyer'),
    (SELECT id FROM resources WHERE name='metal'),
    60000
  );
INSERT INTO public.ships_costs ("ship", "res", "cost")
  VALUES(
    (SELECT id FROM ships WHERE name='destroyer'),
    (SELECT id FROM resources WHERE name='crystal'),
    50000
  );
INSERT INTO public.ships_costs ("ship", "res", "cost")
  VALUES(
    (SELECT id FROM ships WHERE name='destroyer'),
    (SELECT id FROM resources WHERE name='deuterium'),
    15000
  );

INSERT INTO public.ships_costs ("ship", "res", "cost")
  VALUES(
    (SELECT id FROM ships WHERE name='deathstar'),
    (SELECT id FROM resources WHERE name='metal'),
    5000000
  );
INSERT INTO public.ships_costs ("ship", "res", "cost")
  VALUES(
    (SELECT id FROM ships WHERE name='deathstar'),
    (SELECT id FROM resources WHERE name='crystal'),
    4000000
  );
INSERT INTO public.ships_costs ("ship", "res", "cost")
  VALUES(
    (SELECT id FROM ships WHERE name='deathstar'),
    (SELECT id FROM resources WHERE name='deuterium'),
    1000000
  );

INSERT INTO public.ships_costs ("ship", "res", "cost")
  VALUES(
    (SELECT id FROM ships WHERE name='recycler'),
    (SELECT id FROM resources WHERE name='metal'),
    10000
  );
INSERT INTO public.ships_costs ("ship", "res", "cost")
  VALUES(
    (SELECT id FROM ships WHERE name='recycler'),
    (SELECT id FROM resources WHERE name='crystal'),
    6000
  );
INSERT INTO public.ships_costs ("ship", "res", "cost")
  VALUES(
    (SELECT id FROM ships WHERE name='recycler'),
    (SELECT id FROM resources WHERE name='deuterium'),
    2000
  );

INSERT INTO public.ships_costs ("ship", "res", "cost")
  VALUES(
    (SELECT id FROM ships WHERE name='espionage probe'),
    (SELECT id FROM resources WHERE name='crystal'),
    1000
  );

INSERT INTO public.ships_costs ("ship", "res", "cost")
  VALUES(
    (SELECT id FROM ships WHERE name='solar satellite'),
    (SELECT id FROM resources WHERE name='crystal'),
    2000
  );
INSERT INTO public.ships_costs ("ship", "res", "cost")
  VALUES(
    (SELECT id FROM ships WHERE name='solar satellite'),
    (SELECT id FROM resources WHERE name='deuterium'),
    500
  );

INSERT INTO public.ships_costs ("ship", "res", "cost")
  VALUES(
    (SELECT id FROM ships WHERE name='colony ship'),
    (SELECT id FROM resources WHERE name='metal'),
    10000
  );
INSERT INTO public.ships_costs ("ship", "res", "cost")
  VALUES(
    (SELECT id FROM ships WHERE name='colony ship'),
    (SELECT id FROM resources WHERE name='crystal'),
    20000
  );
INSERT INTO public.ships_costs ("ship", "res", "cost")
  VALUES(
    (SELECT id FROM ships WHERE name='colony ship'),
    (SELECT id FROM resources WHERE name='deuterium'),
    10000
  );

-- Seed the ships propulsion increase.
INSERT INTO public.ships_propulsion_increase ("propulsion", "increase")
  VALUES(
    (SELECT id FROM technologies WHERE name='combustion drive'),
    1.1
  );

INSERT INTO public.ships_propulsion_increase ("propulsion", "increase")
  VALUES(
    (SELECT id FROM technologies WHERE name='impulse drive'),
    1.2
  );

INSERT INTO public.ships_propulsion_increase ("propulsion", "increase")
  VALUES(
    (SELECT id FROM technologies WHERE name='hyperspace drive'),
    1.3
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

-- TODO: Perform seeding.
