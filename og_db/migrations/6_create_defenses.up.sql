
-- Create the table defining defenses.
CREATE TABLE defenses (
    id uuid NOT NULL DEFAULT uuid_generate_v4(),
    name text,
    construction_time integer NOT NULL,
    shield integer NOT NULL,
    weapon integer NOT NULL,
    PRIMARY KEY (id)
);

-- Create the table defining the cost of a defense.
CREATE TABLE defenses_costs (
  defense uuid NOT NULL references defenses,
  res uuid NOT NULL references resources,
  cost integer NOT NULL
);

-- Seed the available defenses.
-- TODO: Perform seeding.
