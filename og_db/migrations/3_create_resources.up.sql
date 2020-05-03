
-- Create the table defining resources.
CREATE TABLE resources (
  id uuid NOT NULL DEFAULT uuid_generate_v4(),
  name text,
  base_production integer NOT NULL,
  base_storage integer NOT NULL,
  base_amount integer NOT NULL,
  movable boolean NOT NULL,
  PRIMARY KEY (id),
  UNIQUE (name)
);

-- Perform seeding with the base resources.
INSERT INTO public.resources ("name", "base_production", "base_storage", "base_amount", "movable")
  VALUES('metal', 30, 10000, 500, 'true');
INSERT INTO public.resources ("name", "base_production", "base_storage", "base_amount", "movable")
  VALUES('crystal', 15, 10000, 500, 'true');
INSERT INTO public.resources ("name", "base_production", "base_storage", "base_amount", "movable")
  VALUES('deuterium', 0, 10000, 0, 'true');
INSERT INTO public.resources ("name", "base_production", "base_storage", "base_amount", "movable")
  VALUES('energy', 0, 0, 0, 'false');
INSERT INTO public.resources ("name", "base_production", "base_storage", "base_amount", "movable")
  VALUES('antimatter', 0, 0, 0, 'false');
