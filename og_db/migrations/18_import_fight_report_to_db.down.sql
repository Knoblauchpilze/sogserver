
-- Drop the function handling the generation of the fight status.
DROP FUNCTION fleet_fight_report_status(outcome text, player_id uuid);

-- Drop the function handling the generation of the defender participant of a fight.
DROP FUNCTION fleet_fight_report_indigenous_participant(fleet_id uuid, def_remains json, ships_remains json);

-- Drop the function handling the generation of the outtsider participants of a fight.
DROP FUNCTION fleet_fight_report_outsiders_participant(fleet_id uuid, player_id uuid, remains json);

-- Drop the function to create the header of a fight report.
DROP FUNCTION fleet_fight_report_header(fleet_id uuid, player_id uuid);
