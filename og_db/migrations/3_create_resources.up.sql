
-- Create the table defining resources.
CREATE TABLE resources (
    id uuid NOT NULL DEFAULT uuid_generate_v4(),
    name text,
    PRIMARY KEY (id)
);

-- Perform seeding with the base resources.
INSERT INTO public.resources ("name") VALUES('metal');
INSERT INTO public.resources ("name") VALUES('crystal');
INSERT INTO public.resources ("name") VALUES('deuterium');
INSERT INTO public.resources ("name") VALUES('energy');
INSERT INTO public.resources ("name") VALUES('antimatter');
