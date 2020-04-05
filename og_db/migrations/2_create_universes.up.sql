
-- Create the table defining universes.
CREATE TABLE universes (
    id uuid NOT NULL DEFAULT uuid_generate_v4(),
    name text NOT NULL,
    created_at timestamp WITH TIME ZONE DEFAULT current_timestamp,
    economic_speed integer NOT NULL,
    fleet_speed integer NOT NULL,
    research_speed integer NOT NULL,
    fleets_to_ruins_ratio numeric(4,2) NOT NULL,
    defenses_to_ruins_ratio numeric(4,2) NOT NULL DEFAULT 0,
    fleets_consumption_ratio numeric(3, 2),
    galaxies_count integer NOT NULL,
    galaxy_size integer NOT NULL,
    solar_system_size integer NOT NULL,
    PRIMARY KEY (id)
);
