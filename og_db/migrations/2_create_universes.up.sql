
-- Create the table defining universes.
CREATE TABLE universes (
    id uuid NOT NULL DEFAULT uuid_generate_v4(),
    name text,
    created_at timestamp with time zone default current_timestamp,
    PRIMARY KEY (id)
);

-- Trigger to update the `created_at` field of the table.
CREATE TRIGGER update_universe_creation_time BEFORE INSERT ON universes FOR EACH ROW EXECUTE PROCEDURE update_created_at_column();
