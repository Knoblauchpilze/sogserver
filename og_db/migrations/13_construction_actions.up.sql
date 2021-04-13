
-- Create the table defining building construction actions.
CREATE TABLE construction_actions_buildings (
  id uuid NOT NULL DEFAULT uuid_generate_v4(),
  planet uuid NOT NULL,
  element uuid NOT NULL,
  current_level integer NOT NULL,
  desired_level integer NOT NULL,
  points numeric(15, 5) NOT NULL,
  completion_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  FOREIGN KEY (planet) REFERENCES planets(id),
  FOREIGN KEY (element) REFERENCES buildings(id),
  UNIQUE (planet)
);

-- Create the table describing the effects of the building
-- upgrades on the production capacities.
CREATE TABLE construction_actions_buildings_production_effects (
  action uuid NOT NULL,
  resource uuid NOT NULL,
  production_change numeric(15, 5) NOT NULL,
  FOREIGN KEY (action) REFERENCES construction_actions_buildings(id),
  FOREIGN KEY (resource) REFERENCES resources(id),
  UNIQUE (action, resource)
);

-- Similar to the above table but describes the effects of
-- applying an upgrade action on the storage capacities.
CREATE TABLE construction_actions_buildings_storage_effects (
  action uuid NOT NULL,
  resource uuid NOT NULL,
  storage_capacity_change numeric(15, 5) NOT NULL,
  FOREIGN KEY (action) REFERENCES construction_actions_buildings(id),
  FOREIGN KEY (resource) REFERENCES resources(id),
  UNIQUE (action, resource)
);

-- Similar to the above tables, used to describe the table
-- for fields increase for buildings.
CREATE TABLE construction_actions_buildings_fields_effects (
  action uuid NOT NULL,
  additional_fields integer NOT NULL,
  FOREIGN KEY (action) REFERENCES construction_actions_buildings(id),
  UNIQUE (action)
);

-- Create the table defining technologies research actions.
CREATE TABLE construction_actions_technologies (
  id uuid NOT NULL DEFAULT uuid_generate_v4(),
  planet uuid NOT NULL,
  player uuid NOT NULL,
  element uuid NOT NULL,
  current_level integer NOT NULL,
  desired_level integer NOT NULL,
  points numeric(15, 5) NOT NULL,
  completion_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  FOREIGN KEY (player) REFERENCES players(id),
  FOREIGN KEY (element) REFERENCES technologies(id),
  FOREIGN KEY (planet) REFERENCES planets(id),
  UNIQUE (player)
);

-- Create the table defining ships construction actions.
CREATE TABLE construction_actions_ships (
  id uuid NOT NULL DEFAULT uuid_generate_v4(),
  planet uuid NOT NULL,
  element uuid NOT NULL,
  amount integer NOT NULL,
  remaining integer NOT NULL,
  completion_time INTERVAL NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  FOREIGN KEY (planet) REFERENCES planets(id),
  FOREIGN KEY (element) REFERENCES ships(id)
);

-- Create the table defining defenses construction actions.
CREATE TABLE construction_actions_defenses (
  id uuid NOT NULL DEFAULT uuid_generate_v4(),
  planet uuid NOT NULL,
  element uuid NOT NULL,
  amount integer NOT NULL,
  remaining integer NOT NULL,
  completion_time INTERVAL NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  FOREIGN KEY (planet) REFERENCES planets(id),
  FOREIGN KEY (element) REFERENCES defenses(id)
);

-- Create the table defining building construction actions for moons.
CREATE TABLE construction_actions_buildings_moon (
  id uuid NOT NULL DEFAULT uuid_generate_v4(),
  moon uuid NOT NULL,
  element uuid NOT NULL,
  current_level integer NOT NULL,
  desired_level integer NOT NULL,
  completion_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  FOREIGN KEY (moon) REFERENCES moons(id),
  FOREIGN KEY (element) REFERENCES buildings(id),
  UNIQUE (moon)
);

-- Similar table to the one defined for buildings on a planet but in
-- the case of a moon.
CREATE TABLE construction_actions_buildings_fields_effects_moon (
  action uuid NOT NULL,
  additional_fields integer NOT NULL,
  FOREIGN KEY (action) REFERENCES construction_actions_buildings_moon(id),
  UNIQUE (action)
);

-- Create the table defining ships construction actions for moons.
CREATE TABLE construction_actions_ships_moon (
  id uuid NOT NULL DEFAULT uuid_generate_v4(),
  moon uuid NOT NULL,
  element uuid NOT NULL,
  amount integer NOT NULL,
  remaining integer NOT NULL,
  completion_time INTERVAL NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  FOREIGN KEY (moon) REFERENCES moons(id),
  FOREIGN KEY (element) REFERENCES ships(id)
);

-- Create the table defining defenses construction actions for moons.
CREATE TABLE construction_actions_defenses_moon (
  id uuid NOT NULL DEFAULT uuid_generate_v4(),
  moon uuid NOT NULL,
  element uuid NOT NULL,
  amount integer NOT NULL,
  remaining integer NOT NULL,
  completion_time INTERVAL NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  FOREIGN KEY (moon) REFERENCES moons(id),
  FOREIGN KEY (element) REFERENCES defenses(id)
);
