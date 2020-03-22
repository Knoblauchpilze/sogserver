
-- Create the table defining researches.
CREATE TABLE researches (
    id uuid NOT NULL DEFAULT uuid_generate_v4(),
    name text,
    PRIMARY KEY (id)
);

-- Create the table defining the cost of a research.
CREATE TABLE researches_costs (
  research uuid NOT NULL references researches,
  res uuid NOT NULL references resources,
  cost integer NOT NULL
);

-- Create the table defining the law of progression of cost of a research.
CREATE TABLE researches_costs_progress (
  research uuid NOT NULL references researches,
  res uuid NOT NULL references resources,
  progress numeric(15, 5) NOT NULL
);

-- Seed the available researches.
INSERT INTO public.researches ("name") VALUES('energy');
INSERT INTO public.researches ("name") VALUES('laser');
INSERT INTO public.researches ("name") VALUES('ions');
INSERT INTO public.researches ("name") VALUES('hyperspace');
INSERT INTO public.researches ("name") VALUES('plasma');
INSERT INTO public.researches ("name") VALUES('combustion_drive');
INSERT INTO public.researches ("name") VALUES('ion_drive');
INSERT INTO public.researches ("name") VALUES('hyperspace_propulsion');
INSERT INTO public.researches ("name") VALUES('spying');
INSERT INTO public.researches ("name") VALUES('computers');
INSERT INTO public.researches ("name") VALUES('astrophysics');
INSERT INTO public.researches ("name") VALUES('intergalactic_research_network');
INSERT INTO public.researches ("name") VALUES('graviton');
INSERT INTO public.researches ("name") VALUES('weapons');
INSERT INTO public.researches ("name") VALUES('shields');
INSERT INTO public.researches ("name") VALUES('hulls');
-- TODO: Perform seeding.
