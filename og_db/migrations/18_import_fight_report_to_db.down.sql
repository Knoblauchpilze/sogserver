
-- Drop the function handling the generation of a fight report.
DROP FUNCTION fleet_fight_report(fleet_id uuid, outcome text);

-- Drop the function to create the header of a fight report.
DROP FUNCTION fleet_fight_report_header(fleet_id uuid, player_id uuid);
