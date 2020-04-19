
-- Create the table defining defenses.
CREATE TABLE defenses (
  id uuid NOT NULL DEFAULT uuid_generate_v4(),
  name text NOT NULL,
  shield integer NOT NULL,
  weapon integer NOT NULL,
  PRIMARY KEY (id),
  UNIQUE (name)
);

-- Create the table defining the cost of a defense.
CREATE TABLE defenses_costs (
  element uuid NOT NULL,
  res uuid NOT NULL,
  cost integer NOT NULL,
  FOREIGN KEY (element) REFERENCES defenses(id),
  FOREIGN KEY (res) REFERENCES resources(id)
);

-- Seed the available defenses.
INSERT INTO public.defenses ("name", "shield", "weapon") VALUES('rocket launcher', 20, 80);
INSERT INTO public.defenses ("name", "shield", "weapon") VALUES('light laser', 25, 100);
INSERT INTO public.defenses ("name", "shield", "weapon") VALUES('heavy laser', 100, 250);
INSERT INTO public.defenses ("name", "shield", "weapon") VALUES('ion cannon', 500, 150);
INSERT INTO public.defenses ("name", "shield", "weapon") VALUES('gauss cannon', 200, 1100);
INSERT INTO public.defenses ("name", "shield", "weapon") VALUES('plasma turret', 300, 3000);

INSERT INTO public.defenses ("name", "shield", "weapon") VALUES('small shield dome', 2000, 1);
INSERT INTO public.defenses ("name", "shield", "weapon") VALUES('large shield dome', 10000, 1);

INSERT INTO public.defenses ("name", "shield", "weapon") VALUES('anti-ballistic missile', 1, 1);
INSERT INTO public.defenses ("name", "shield", "weapon") VALUES('interplanetary missile', 1, 12000);

-- Seed the defenses costs.
-- Conventional defenses.
INSERT INTO public.defenses_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM defenses WHERE name='rocket launcher'),
    (SELECT id FROM resources WHERE name='metal'),
    2000
  );

INSERT INTO public.defenses_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM defenses WHERE name='light laser'),
    (SELECT id FROM resources WHERE name='metal'),
    1500
  );
INSERT INTO public.defenses_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM defenses WHERE name='light laser'),
    (SELECT id FROM resources WHERE name='crystal'),
    500
  );

INSERT INTO public.defenses_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM defenses WHERE name='heavy laser'),
    (SELECT id FROM resources WHERE name='metal'),
    6000
  );
INSERT INTO public.defenses_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM defenses WHERE name='heavy laser'),
    (SELECT id FROM resources WHERE name='crystal'),
    2000
  );

INSERT INTO public.defenses_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM defenses WHERE name='ion cannon'),
    (SELECT id FROM resources WHERE name='metal'),
    2000
  );
INSERT INTO public.defenses_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM defenses WHERE name='ion cannon'),
    (SELECT id FROM resources WHERE name='crystal'),
    6000
  );

INSERT INTO public.defenses_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM defenses WHERE name='gauss cannon'),
    (SELECT id FROM resources WHERE name='metal'),
    20000
  );
INSERT INTO public.defenses_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM defenses WHERE name='gauss cannon'),
    (SELECT id FROM resources WHERE name='crystal'),
    15000
  );
INSERT INTO public.defenses_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM defenses WHERE name='gauss cannon'),
    (SELECT id FROM resources WHERE name='deuterium'),
    2000
  );

INSERT INTO public.defenses_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM defenses WHERE name='plasma turret'),
    (SELECT id FROM resources WHERE name='metal'),
    50000
  );
INSERT INTO public.defenses_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM defenses WHERE name='plasma turret'),
    (SELECT id FROM resources WHERE name='crystal'),
    50000
  );
INSERT INTO public.defenses_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM defenses WHERE name='plasma turret'),
    (SELECT id FROM resources WHERE name='deuterium'),
    30000
  );

-- Shield domes.
INSERT INTO public.defenses_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM defenses WHERE name='small shield dome'),
    (SELECT id FROM resources WHERE name='metal'),
    10000
  );
INSERT INTO public.defenses_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM defenses WHERE name='small shield dome'),
    (SELECT id FROM resources WHERE name='crystal'),
    10000
  );

INSERT INTO public.defenses_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM defenses WHERE name='large shield dome'),
    (SELECT id FROM resources WHERE name='metal'),
    50000
  );
INSERT INTO public.defenses_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM defenses WHERE name='large shield dome'),
    (SELECT id FROM resources WHERE name='crystal'),
    50000
  );

-- Missiles.
INSERT INTO public.defenses_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM defenses WHERE name='anti-ballistic missile'),
    (SELECT id FROM resources WHERE name='metal'),
    8000
  );
INSERT INTO public.defenses_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM defenses WHERE name='anti-ballistic missile'),
    (SELECT id FROM resources WHERE name='deuterium'),
    2000
  );

INSERT INTO public.defenses_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM defenses WHERE name='interplanetary missile'),
    (SELECT id FROM resources WHERE name='metal'),
    12500
  );
INSERT INTO public.defenses_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM defenses WHERE name='interplanetary missile'),
    (SELECT id FROM resources WHERE name='crystal'),
    2500
  );
INSERT INTO public.defenses_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM defenses WHERE name='interplanetary missile'),
    (SELECT id FROM resources WHERE name='deuterium'),
    10000
  );
