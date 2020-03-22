
-- Create the table defining universes.
CREATE TABLE universes (
    id uuid NOT NULL DEFAULT uuid_generate_v4(),
    name text NOT NULL,
    created_at timestamp with time zone default current_timestamp,
    economic_speed integer NOT NULL,
    fleet_speed integer NOT NULL,
    research_speed integer NOT NULL,
    fleet_to_ruins_ratio numeric(4,2) NOT NULL,
    defense_to_ruins_ratio numeric(4,2) NOT NULL default 0,
    consumption_ratio numeric(3, 2),
    galaxy_count integer NOT NULL,
    solar_system_size integer NOT NULL,
    PRIMARY KEY (id)
);

-- Trigger to update the `created_at` field of the table.
CREATE TRIGGER update_universe_creation_time BEFORE INSERT ON universes FOR EACH ROW EXECUTE PROCEDURE update_created_at_column();
