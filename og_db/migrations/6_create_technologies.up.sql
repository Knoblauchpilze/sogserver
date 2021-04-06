
-- Create the table defining technologies.
CREATE TABLE technologies (
  id uuid NOT NULL DEFAULT uuid_generate_v4(),
  name text NOT NULL,
  PRIMARY KEY (id),
  UNIQUE (name)
);

-- Create the table defining the cost of a technology.
CREATE TABLE technologies_costs (
  element uuid NOT NULL,
  res uuid NOT NULL,
  cost integer NOT NULL,
  FOREIGN KEY (element) REFERENCES technologies(id),
  FOREIGN KEY (res) REFERENCES resources(id)
);

-- Create the table defining the law of progression of cost of a technology.
CREATE TABLE technologies_costs_progress (
  element uuid NOT NULL,
  progress numeric(15, 5) NOT NULL,
  FOREIGN KEY (element) REFERENCES technologies(id)
);

-- Seed the available technologies.
INSERT INTO public.technologies ("name") VALUES('energy');
INSERT INTO public.technologies ("name") VALUES('laser');
INSERT INTO public.technologies ("name") VALUES('ions');
INSERT INTO public.technologies ("name") VALUES('hyperspace');
INSERT INTO public.technologies ("name") VALUES('plasma');

INSERT INTO public.technologies ("name") VALUES('combustion drive');
INSERT INTO public.technologies ("name") VALUES('impulse drive');
INSERT INTO public.technologies ("name") VALUES('hyperspace drive');

INSERT INTO public.technologies ("name") VALUES('espionage');
INSERT INTO public.technologies ("name") VALUES('computers');
INSERT INTO public.technologies ("name") VALUES('astrophysics');
INSERT INTO public.technologies ("name") VALUES('intergalactic research network');
INSERT INTO public.technologies ("name") VALUES('graviton');

INSERT INTO public.technologies ("name") VALUES('weapons');
INSERT INTO public.technologies ("name") VALUES('shielding');
INSERT INTO public.technologies ("name") VALUES('armour');

-- Seed the technologies costs.
-- Fundamental techonologies.
INSERT INTO public.technologies_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM technologies WHERE name='energy'),
    (SELECT id FROM resources WHERE name='crystal'),
    800
  );
INSERT INTO public.technologies_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM technologies WHERE name='energy'),
    (SELECT id FROM resources WHERE name='deuterium'),
    400
  );

INSERT INTO public.technologies_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM technologies WHERE name='laser'),
    (SELECT id FROM resources WHERE name='metal'),
    200
  );
INSERT INTO public.technologies_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM technologies WHERE name='laser'),
    (SELECT id FROM resources WHERE name='crystal'),
    100
  );

INSERT INTO public.technologies_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM technologies WHERE name='ions'),
    (SELECT id FROM resources WHERE name='metal'),
    1000
  );
INSERT INTO public.technologies_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM technologies WHERE name='ions'),
    (SELECT id FROM resources WHERE name='crystal'),
    300
  );
INSERT INTO public.technologies_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM technologies WHERE name='ions'),
    (SELECT id FROM resources WHERE name='deuterium'),
    100
  );

INSERT INTO public.technologies_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM technologies WHERE name='hyperspace'),
    (SELECT id FROM resources WHERE name='crystal'),
    4000
  );
INSERT INTO public.technologies_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM technologies WHERE name='hyperspace'),
    (SELECT id FROM resources WHERE name='deuterium'),
    2000
  );

INSERT INTO public.technologies_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM technologies WHERE name='plasma'),
    (SELECT id FROM resources WHERE name='metal'),
    2000
  );
INSERT INTO public.technologies_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM technologies WHERE name='plasma'),
    (SELECT id FROM resources WHERE name='crystal'),
    4000
  );
INSERT INTO public.technologies_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM technologies WHERE name='plasma'),
    (SELECT id FROM resources WHERE name='deuterium'),
    1000
  );

-- Propulsion techonologies.
INSERT INTO public.technologies_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM technologies WHERE name='combustion drive'),
    (SELECT id FROM resources WHERE name='metal'),
    400
  );
INSERT INTO public.technologies_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM technologies WHERE name='combustion drive'),
    (SELECT id FROM resources WHERE name='deuterium'),
    600
  );

INSERT INTO public.technologies_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM technologies WHERE name='impulse drive'),
    (SELECT id FROM resources WHERE name='metal'),
    2000
  );
INSERT INTO public.technologies_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM technologies WHERE name='impulse drive'),
    (SELECT id FROM resources WHERE name='crystal'),
    4000
  );
INSERT INTO public.technologies_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM technologies WHERE name='impulse drive'),
    (SELECT id FROM resources WHERE name='deuterium'),
    600
  );

INSERT INTO public.technologies_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM technologies WHERE name='hyperspace drive'),
    (SELECT id FROM resources WHERE name='metal'),
    10000
  );
INSERT INTO public.technologies_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM technologies WHERE name='hyperspace drive'),
    (SELECT id FROM resources WHERE name='crystal'),
    20000
  );
INSERT INTO public.technologies_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM technologies WHERE name='hyperspace drive'),
    (SELECT id FROM resources WHERE name='deuterium'),
    6000
  );

-- Advanced technologies.
INSERT INTO public.technologies_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM technologies WHERE name='espionage'),
    (SELECT id FROM resources WHERE name='metal'),
    200
  );
INSERT INTO public.technologies_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM technologies WHERE name='espionage'),
    (SELECT id FROM resources WHERE name='crystal'),
    1000
  );
INSERT INTO public.technologies_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM technologies WHERE name='espionage'),
    (SELECT id FROM resources WHERE name='deuterium'),
    200
  );

INSERT INTO public.technologies_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM technologies WHERE name='computers'),
    (SELECT id FROM resources WHERE name='crystal'),
    400
  );
INSERT INTO public.technologies_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM technologies WHERE name='computers'),
    (SELECT id FROM resources WHERE name='deuterium'),
    600
  );

INSERT INTO public.technologies_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM technologies WHERE name='astrophysics'),
    (SELECT id FROM resources WHERE name='metal'),
    4000
  );
INSERT INTO public.technologies_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM technologies WHERE name='astrophysics'),
    (SELECT id FROM resources WHERE name='crystal'),
    8000
  );
INSERT INTO public.technologies_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM technologies WHERE name='astrophysics'),
    (SELECT id FROM resources WHERE name='deuterium'),
    4000
  );

INSERT INTO public.technologies_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM technologies WHERE name='intergalactic research network'),
    (SELECT id FROM resources WHERE name='metal'),
    240000
  );
INSERT INTO public.technologies_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM technologies WHERE name='intergalactic research network'),
    (SELECT id FROM resources WHERE name='crystal'),
    400000
  );
INSERT INTO public.technologies_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM technologies WHERE name='intergalactic research network'),
    (SELECT id FROM resources WHERE name='deuterium'),
    160000
  );

INSERT INTO public.technologies_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM technologies WHERE name='graviton'),
    (SELECT id FROM resources WHERE name='energy'),
    300000
  );

-- Combat technologies.
INSERT INTO public.technologies_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM technologies WHERE name='weapons'),
    (SELECT id FROM resources WHERE name='metal'),
    800
  );
INSERT INTO public.technologies_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM technologies WHERE name='weapons'),
    (SELECT id FROM resources WHERE name='crystal'),
    200
  );

INSERT INTO public.technologies_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM technologies WHERE name='shielding'),
    (SELECT id FROM resources WHERE name='metal'),
    200
  );
INSERT INTO public.technologies_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM technologies WHERE name='shielding'),
    (SELECT id FROM resources WHERE name='crystal'),
    600
  );

INSERT INTO public.technologies_costs ("element", "res", "cost")
  VALUES(
    (SELECT id FROM technologies WHERE name='armour'),
    (SELECT id FROM resources WHERE name='metal'),
    1000
  );

-- Seed the researches costs progress.
-- Fundamental technologies.
INSERT INTO public.technologies_costs_progress ("element", "progress")
  VALUES(
    (SELECT id FROM technologies WHERE name='energy'),
    2
  );
INSERT INTO public.technologies_costs_progress ("element", "progress")
  VALUES(
    (SELECT id FROM technologies WHERE name='laser'),
    2
  );
INSERT INTO public.technologies_costs_progress ("element", "progress")
  VALUES(
    (SELECT id FROM technologies WHERE name='ions'),
    2
  );
INSERT INTO public.technologies_costs_progress ("element", "progress")
  VALUES(
    (SELECT id FROM technologies WHERE name='hyperspace'),
    2
  );
INSERT INTO public.technologies_costs_progress ("element", "progress")
  VALUES(
    (SELECT id FROM technologies WHERE name='plasma'),
    2
  );

-- Propulstion technologies.
INSERT INTO public.technologies_costs_progress ("element", "progress")
  VALUES(
    (SELECT id FROM technologies WHERE name='combustion drive'),
    2
  );
INSERT INTO public.technologies_costs_progress ("element", "progress")
  VALUES(
    (SELECT id FROM technologies WHERE name='impulse drive'),
    2
  );
INSERT INTO public.technologies_costs_progress ("element", "progress")
  VALUES(
    (SELECT id FROM technologies WHERE name='hyperspace drive'),
    2
  );

-- Advanced technologies.
INSERT INTO public.technologies_costs_progress ("element", "progress")
  VALUES(
    (SELECT id FROM technologies WHERE name='espionage'),
    2
  );
INSERT INTO public.technologies_costs_progress ("element", "progress")
  VALUES(
    (SELECT id FROM technologies WHERE name='computers'),
    2
  );
INSERT INTO public.technologies_costs_progress ("element", "progress")
  VALUES(
    (SELECT id FROM technologies WHERE name='astrophysics'),
    1.75
  );
INSERT INTO public.technologies_costs_progress ("element", "progress")
  VALUES(
    (SELECT id FROM technologies WHERE name='intergalactic research network'),
    2
  );
INSERT INTO public.technologies_costs_progress ("element", "progress")
  VALUES(
    (SELECT id FROM technologies WHERE name='graviton'),
    3
  );

-- Combat technologies.
INSERT INTO public.technologies_costs_progress ("element", "progress")
  VALUES(
    (SELECT id FROM technologies WHERE name='weapons'),
    2
  );
INSERT INTO public.technologies_costs_progress ("element", "progress")
  VALUES(
    (SELECT id FROM technologies WHERE name='shielding'),
    2
  );
INSERT INTO public.technologies_costs_progress ("element", "progress")
  VALUES(
    (SELECT id FROM technologies WHERE name='armour'),
    2
  );
