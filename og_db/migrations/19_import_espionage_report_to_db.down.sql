
-- Drop the function handling the generation of the espionage report.
DROP FUNCTION espionage_report(fleet_id uuid, counter_espionage integer, info_level integer);

-- Drop the function generating the counter espionage report.
DROP FUNCTION generate_counter_espionage_report(fleet_id uuid, prob integer);
