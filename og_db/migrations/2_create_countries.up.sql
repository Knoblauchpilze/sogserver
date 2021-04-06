
-- Create the table defining countries.
CREATE TABLE countries (
  id uuid NOT NULL DEFAULT uuid_generate_v4(),
  name text NOT NULL,
  PRIMARY KEY (id),
  UNIQUE (name)
);

-- Perform seeding with the base countries.
INSERT INTO public.countries ("name")
  VALUES('france');
INSERT INTO public.countries ("name")
  VALUES('germany');
INSERT INTO public.countries ("name")
  VALUES('spain');
