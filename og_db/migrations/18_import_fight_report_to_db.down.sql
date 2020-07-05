
-- Drop the orcherstration function to generate the fight reports.
DROP FUNCTION fight_report(players json, fleets json, indigenous uuid, planet_id uuid, planet_kind text, moment timestamp with time zone, outcome text, fleet_remains json, ships_remains json, def_remains json, pillage json, debris json, rebuilt integer);

-- Drop the function handling the generation of a fight report footer.
DROP FUNCTION fleet_fight_report_footer(player_id uuid, pillage json, debris json, rebuilt integer);

-- Drop the function handling the generation of the fight status.
DROP FUNCTION fleet_fight_report_status(player_id uuid, outcome text);

-- Drop the function handling the generation of the defender participant of a fight.
DROP FUNCTION fleet_fight_report_indigenous_participant(player_id uuid, planet_id uuid, kind text, ships_remains json, def_remains json);

-- Drop the function handling the generation of the outtsider participants of a fight.
DROP FUNCTION fleet_fight_report_outsider_participant(player_id uuid, fleet_id uuid, remains json);

-- Drop the function to create the header of a fight report.
DROP FUNCTION fleet_fight_report_header(player_id uuid, planet_id uuid, planet_kind text, moment timestamp with time zone);
