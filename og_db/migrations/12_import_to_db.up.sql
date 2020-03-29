
-- Import the universes into the corresponding table.
CREATE OR REPLACE FUNCTION create_universe(inputs json) RETURNS VOID AS $$
BEGIN
  INSERT INTO universes SELECT * FROM json_populate_record(null::universes, inputs);
END
$$ LANGUAGE plpgsql;

-- Import the accounts into the corresponding table.
CREATE OR REPLACE FUNCTION create_account(inputs json) RETURNS VOID AS $$
BEGIN
  INSERT INTO accounts SELECT * FROM json_populate_record(null::accounts, inputs);
END
$$ LANGUAGE plpgsql;

-- Create players from the account and universe data.
CREATE OR REPLACE FUNCTION create_player(inputs json) RETURNS VOID AS $$
BEGIN
  INSERT INTO players
  SELECT *
  FROM json_populate_record(null::players, inputs)
  -- Only insert elements which have a corresponding element in the universes
  -- and in the accounts tables. This guarantee that we're not creating players
  -- in unknown universes or attached to unknown accounts.
  WHERE EXISTS (
          SELECT u.id
          FROM universes u
          WHERE u.id = uni
        )
        and EXISTS (
          SELECT a.id
          FROM accounts a
          WHERE a.id = account
        );
END
$$ LANGUAGE plpgsql;
