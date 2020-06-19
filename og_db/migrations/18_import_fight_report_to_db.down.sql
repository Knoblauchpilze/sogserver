
-- Drop the function handling the generation of a fight report.
DROP FUNCTION fleet_fight_report(fleet_id uuid, outcome text);

-- Drop the function handling the generation of the fight status.
DROP FUNCTION fleet_fight_report_status(outcome text, player_id uuid);

-- Drop the function handling the generation of the participant of a fight.
DROP FUNCTION fleet_fight_report_attacker_participant(fleet_id uuid, player_id uuid);

-- Drop the function to create the header of a fight report.
DROP FUNCTION fleet_fight_report_header(fleet_id uuid, player_id uuid);
