
-- Create the table defining building construction actions.
CREATE TABLE construction_actions_buildings (
    id uuid NOT NULL DEFAULT uuid_generate_v4(),
    planet uuid NOT NULL,
    building uuid NOT NULL,
    level integer NOT NULL,
    completion_time timestamp WITH TIME ZONE DEFAULT current_timestamp,
    created_at timestamp WITH TIME ZONE DEFAULT current_timestamp,
    FOREIGN KEY (planet) REFERENCES planets(id),
    FOREIGN KEY (building) REFERENCES buildings(id)
);

-- Create the table defining technologies research actions.
CREATE TABLE construction_actions_technologies (
    id uuid NOT NULL DEFAULT uuid_generate_v4(),
    player uuid NOT NULL,
    technology uuid NOT NULL,
    level integer NOT NULL,
    completion_time timestamp WITH TIME ZONE DEFAULT current_timestamp,
    created_at timestamp WITH TIME ZONE DEFAULT current_timestamp,
    FOREIGN KEY (player) REFERENCES players(id),
    FOREIGN KEY (technology) REFERENCES technologies(id)
);

-- Create the table defining ships construction actions.
CREATE TABLE construction_actions_ships (
    id uuid NOT NULL DEFAULT uuid_generate_v4(),
    planet uuid NOT NULL,
    ship uuid NOT NULL,
    completion_time timestamp WITH TIME ZONE DEFAULT current_timestamp,
    created_at timestamp WITH TIME ZONE DEFAULT current_timestamp,
    FOREIGN KEY (planet) REFERENCES planets(id),
    FOREIGN KEY (ship) REFERENCES ships(id)
);

-- Create the table defining defenses construction actions.
CREATE TABLE construction_actions_defenses (
    id uuid NOT NULL DEFAULT uuid_generate_v4(),
    planet uuid NOT NULL,
    defense uuid NOT NULL,
    completion_time timestamp WITH TIME ZONE DEFAULT current_timestamp,
    created_at timestamp WITH TIME ZONE DEFAULT current_timestamp,
    FOREIGN KEY (planet) REFERENCES planets(id),
    FOREIGN KEY (defense) REFERENCES defenses(id)
);
