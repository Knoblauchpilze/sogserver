
-- Create the table defining buildings.
CREATE TABLE buildings (
    id uuid NOT NULL DEFAULT uuid_generate_v4(),
    name text,
    construction_time integer NOT NULL,
    PRIMARY KEY (id)
);

-- Create the table defining the cost of a building.
CREATE TABLE buildings_costs (
  building uuid NOT NULL references buildings,
  res uuid NOT NULL references resources,
  cost integer NOT NULL
);

-- Create the table defining the resources gain of a building.
CREATE TABLE buildings_gains (
  building uuid NOT NULL references buildings,
  res uuid NOT NULL references resources,
  gain integer NOT NULL
);

-- Create the table defining the law of progression of cost of a building.
CREATE TABLE buildings_costs_progress (
  building uuid NOT NULL references buildings,
  res uuid NOT NULL references resources,
  progress numeric(15, 5) NOT NULL
);

-- Create the table defining the law of progression of gains of a building.
CREATE TABLE buildings_gains_progress (
  building uuid NOT NULL references buildings,
  res uuid NOT NULL references resources,
  progress numeric(15, 5) NOT NULL
);

-- Seed the available buildings.
-- TODO: Perform seeding.
