
-- Drop the moon's deletion script.
DROP FUNCTION delete_moon(moon_id uuid);

-- Drop the moon's creation script.
DROP FUNCTION create_moon(moon_data json, resources json);

-- Drop the planet's creation script.
DROP FUNCTION create_planet(planet_data json, resources json, moment TIMESTAMP WITH TIME ZONE);

-- Drop the player's creation script.
DROP FUNCTION create_player(inputs json);

-- Drop the account's creation script.
DROP FUNCTION create_account(inputs json);

-- Drop the universe's creation script.
DROP FUNCTION create_universe(inputs json);
