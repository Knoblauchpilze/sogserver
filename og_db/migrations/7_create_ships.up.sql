
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
-- TODO: Perform seeding.
