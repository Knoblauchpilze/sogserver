
-- Drop the orcherstration function to generate the fight reports.
DROP FUNCTION fight_report(fleet_id uuid, outcome text, fleet_remains json, ships_remains json, def_remains json, pillage json, debris json, rebuilt json);

-- Drop the function handling the generation of a fight report footer.
DROP FUNCTION fight_report_footer(player_id uuid, pillage json, debris json, rebuilt json);

-- Drop the function handling the generation of the fight status.
DROP FUNCTION fleet_fight_report_status(outcome text, player_id uuid);

-- Drop the function handling the generation of the defender participant of a fight.
DROP FUNCTION fleet_fight_report_indigenous_participant(planet_id uuid, kind text, player_id uuid, ships_remains json, def_remains json);

-- Drop the function handling the generation of the outtsider participants of a fight.
DROP FUNCTION fleet_fight_report_outsiders_participant(fleet_id uuid, player_id uuid, remains json);

-- Drop the function to create the header of a fight report.
DROP FUNCTION fleet_fight_report_header(fleet_id uuid, player_id uuid);
