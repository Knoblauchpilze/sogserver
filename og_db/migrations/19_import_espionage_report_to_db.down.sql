
-- Drop the function handling the generation of the espionage report.
DROP FUNCTION espionage_report(fleet_id uuid, counter_espionage integer, info_level integer);

-- Drop the function generating the technologies report.
DROP FUNCTION generate_technologies_report(player_id uuid, fleet_id uuid, pOffset integer, report_id uuid);

-- Drop the function generating the buildings report.
DROP FUNCTION generate_buildings_report(player_id uuid, fleet_id uuid, pOffset integer, report_id uuid);

-- Drop the function generating the defenses report.
DROP FUNCTION generate_defenses_report(player_id uuid, fleet_id uuid, pOffset integer, report_id uuid);

-- Drop the function generating the ships report.
DROP FUNCTION generate_ships_report(player_id uuid, fleet_id uuid, pOffset integer, report_id uuid);

-- Drop the function generating the activity report.
DROP FUNCTION generate_activity_report(player_id uuid, fleet_id uuid, pOffset integer, report_id uuid);

-- Drop the function generating the resources report.
DROP FUNCTION generate_resources_report(player_id uuid, fleet_id uuid, pOffset integer, report_id uuid);

-- Drop the function generating the header of the espionage report.
DROP FUNCTION generate_header_report(player_id uuid, fleet_id uuid, report_id uuid);

-- Drop the function generating the counter espionage report.
DROP FUNCTION generate_counter_espionage_report(fleet_id uuid, prob integer);
