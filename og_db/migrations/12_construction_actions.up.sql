
-- Create the table defining building construction actions.
CREATE TABLE construction_actions_buildings (
  id uuid NOT NULL DEFAULT uuid_generate_v4(),
  planet uuid NOT NULL,
  building uuid NOT NULL,
  current_level integer NOT NULL,
  desired_level integer NOT NULL,
  completion_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  FOREIGN KEY (planet) REFERENCES planets(id),
  FOREIGN KEY (building) REFERENCES buildings(id),
  UNIQUE (planet)
);

-- Create the table describing the effects of the building
-- upgrades on the production capacities.
CREATE TABLE construction_actions_buildings_production_effects (
  action uuid NOT NULL,
  res uuid NOT NULL,
  new_production numeric(15, 5) NOT NULL,
  FOREIGN KEY (action) REFERENCES construction_actions_buildings(id),
  FOREIGN KEY (res) REFERENCES resources(id),
  UNIQUE (action, res)
);

-- Similar to the above table but describes the effects of
-- applying an upgrade action on the storage capacities.
CREATE TABLE construction_actions_buildings_storage_effects (
  action uuid NOT NULL,
  res uuid NOT NULL,
  new_storage_capacity numeric(15, 5) NOT NULL,
  FOREIGN KEY (action) REFERENCES construction_actions_buildings(id),
  FOREIGN KEY (res) REFERENCES resources(id),
  UNIQUE (action, res)
);

-- Create the trigger on the table to update the `created_at` field.
CREATE TRIGGER update_buildings_action_creation BEFORE INSERT ON construction_actions_buildings FOR EACH ROW EXECUTE PROCEDURE update_created_at();

-- Create the table defining technologies research actions.
CREATE TABLE construction_actions_technologies (
  id uuid NOT NULL DEFAULT uuid_generate_v4(),
  player uuid NOT NULL,
  technology uuid NOT NULL,
  planet uuid NOT NULL,
  current_level integer NOT NULL,
  desired_level integer NOT NULL,
  completion_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  FOREIGN KEY (player) REFERENCES players(id),
  FOREIGN KEY (technology) REFERENCES technologies(id),
  FOREIGN KEY (planet) REFERENCES planets(id),
  UNIQUE (player)
);

-- Create the trigger on the table to update the `created_at` field.
CREATE TRIGGER update_technologies_action_creation BEFORE INSERT ON construction_actions_technologies FOR EACH ROW EXECUTE PROCEDURE update_created_at();

-- Create the table defining ships construction actions.
CREATE TABLE construction_actions_ships (
  id uuid NOT NULL DEFAULT uuid_generate_v4(),
  planet uuid NOT NULL,
  ship uuid NOT NULL,
  amount integer NOT NULL,
  remaining integer NOT NULL,
  completion_time INTERVAL NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  FOREIGN KEY (planet) REFERENCES planets(id),
  FOREIGN KEY (ship) REFERENCES ships(id)
);

-- Create the trigger on the table to update the `created_at` field.
CREATE TRIGGER update_ships_action_creation BEFORE INSERT ON construction_actions_ships FOR EACH ROW EXECUTE PROCEDURE update_created_at();

-- Create the table defining defenses construction actions.
CREATE TABLE construction_actions_defenses (
  id uuid NOT NULL DEFAULT uuid_generate_v4(),
  planet uuid NOT NULL,
  defense uuid NOT NULL,
  amount integer NOT NULL,
  remaining integer NOT NULL,
  completion_time INTERVAL NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  FOREIGN KEY (planet) REFERENCES planets(id),
  FOREIGN KEY (defense) REFERENCES defenses(id)
);

-- Create the trigger on the table to update the `created_at` field.
CREATE TRIGGER update_defenses_action_creation BEFORE INSERT ON construction_actions_defenses FOR EACH ROW EXECUTE PROCEDURE update_created_at();
