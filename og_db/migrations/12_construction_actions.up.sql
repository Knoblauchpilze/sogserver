
-- Create the table defining building construction actions.
CREATE TABLE construction_actions_buildings (
    planet uuid NOT NULL,
    building uuid NOT NULL,
    level integer NOT NULL,
    completion_time timestamp WITH TIME ZONE DEFAULT current_timestamp,
    created_at timestamp WITH TIME ZONE DEFAULT current_timestamp,
    FOREIGN KEY (planet) REFERENCES planets(id),
    FOREIGN KEY (building) REFERENCES buildings(id)
);

-- Trigger to update the `created_at` field of the table.
CREATE TRIGGER update_building_action_creation_time BEFORE INSERT ON construction_actions_buildings FOR EACH ROW EXECUTE PROCEDURE update_created_at_column();

-- Create the table defining technologies research actions.
CREATE TABLE construction_actions_technologies (
    player uuid NOT NULL,
    technology uuid NOT NULL,
    level integer NOT NULL,
    completion_time timestamp WITH TIME ZONE DEFAULT current_timestamp,
    created_at timestamp WITH TIME ZONE DEFAULT current_timestamp,
    FOREIGN KEY (player) REFERENCES players(id),
    FOREIGN KEY (technology) REFERENCES technologies(id)
);

-- Trigger to update the `created_at` field of the table.
CREATE TRIGGER update_technology_action_creation_time BEFORE INSERT ON construction_actions_technologies FOR EACH ROW EXECUTE PROCEDURE update_created_at_column();

-- Create the table defining ships construction actions.
CREATE TABLE construction_actions_ships (
    planet uuid NOT NULL,
    ship uuid NOT NULL,
    completion_time timestamp WITH TIME ZONE DEFAULT current_timestamp,
    created_at timestamp WITH TIME ZONE DEFAULT current_timestamp,
    FOREIGN KEY (planet) REFERENCES planets(id),
    FOREIGN KEY (ship) REFERENCES ships(id)
);

-- Trigger to update the `created_at` field of the table.
CREATE TRIGGER update_ship_action_creation_time BEFORE INSERT ON construction_actions_ships FOR EACH ROW EXECUTE PROCEDURE update_created_at_column();

-- Create the table defining defenses construction actions.
CREATE TABLE construction_actions_defenses (
    planet uuid NOT NULL,
    defense uuid NOT NULL,
    completion_time timestamp WITH TIME ZONE DEFAULT current_timestamp,
    created_at timestamp WITH TIME ZONE DEFAULT current_timestamp,
    FOREIGN KEY (planet) REFERENCES planets(id),
    FOREIGN KEY (defense) REFERENCES defenses(id)
);

-- Trigger to update the `created_at` field of the table.
CREATE TRIGGER update_defense_action_creation_time BEFORE INSERT ON construction_actions_defenses FOR EACH ROW EXECUTE PROCEDURE update_created_at_column();
