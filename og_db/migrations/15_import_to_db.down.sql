
-- Drop the planet's creation script.
DROP FUNCTION create_planet(planet json, resources json, moment TIMESTAMP WITH TIME ZONE);

-- Drop the player's creation script.
DROP FUNCTION create_player(inputs json);

-- Drop the account's creation script.
DROP FUNCTION create_account(inputs json);

-- Drop the universe's creation script.
DROP FUNCTION create_universe(inputs json);
