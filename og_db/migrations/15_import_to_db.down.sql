
-- Drop the moon's deletion script.
DROP FUNCTION delete_moon(moon_id uuid);

-- Drop the moon's creation script.
DROP FUNCTION create_moon(moon_data json, resources json);

-- Drop the planet's update script.
DROP FUNCTION update_planet(planet_id uuid, inputs json);

-- Drop the planet's creation script.
DROP FUNCTION create_planet(planet_data json, resources json, moment TIMESTAMP WITH TIME ZONE);

-- Drop the player's update script.
DROP FUNCTION update_player(player_id uuid, inputs json);

-- Drop the player's creation script.
DROP FUNCTION create_player(inputs json);

-- Drop the account's update script.
DROP FUNCTION update_account(account_id uuid, inputs json);

-- Drop the account's creation script.
DROP FUNCTION create_account(inputs json);

-- Drop the universe's creation script.
DROP FUNCTION create_universe(inputs json);
