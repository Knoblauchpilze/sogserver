
-- Create the table defining resources.
CREATE TABLE resources (
  id uuid NOT NULL DEFAULT uuid_generate_v4(),
  name text,
  base_production integer NOT NULL,
  base_storage integer NOT NULL,
  base_amount integer NOT NULL,
  movable boolean NOT NULL,
  storable boolean NOT NULL,
  dispersable boolean NOT NULL,
  economy_scalable boolean NOT NULL,
  PRIMARY KEY (id),
  UNIQUE (name)
);

-- Perform seeding with the base resources.
-- TODO: Hack to have a lot of resources when starting. Replace the `base_amount` for resources with `500`.
INSERT INTO public.resources ("name", "base_production", "base_storage", "base_amount", "movable", "storable", "dispersable", "economy_scalable")
VALUES('metal', 30, 10000, 100000, 'true', 'true', 'true', 'true');
  -- VALUES('metal', 30, 10000, 500, 'true', 'true', 'true', 'true');
INSERT INTO public.resources ("name", "base_production", "base_storage", "base_amount", "movable", "storable", "dispersable", "economy_scalable")
  VALUES('crystal', 15, 10000, 100000, 'true', 'true', 'true', 'true');
  -- VALUES('crystal', 15, 10000, 500, 'true', 'true', 'true', 'true');
INSERT INTO public.resources ("name", "base_production", "base_storage", "base_amount", "movable", "storable", "dispersable", "economy_scalable")
  VALUES('deuterium', 0, 10000, 100000, 'true', 'true', 'false', 'true');
  -- VALUES('deuterium', 0, 10000, 0, 'true', 'true', 'false', 'true');
INSERT INTO public.resources ("name", "base_production", "base_storage", "base_amount", "movable", "storable", "dispersable", "economy_scalable")
  VALUES('energy', 0, 0, 0, 'false', 'false', 'false', 'false');
INSERT INTO public.resources ("name", "base_production", "base_storage", "base_amount", "movable", "storable", "dispersable", "economy_scalable")
  VALUES('antimatter', 0, 0, 0, 'false', 'false', 'false', 'false');
