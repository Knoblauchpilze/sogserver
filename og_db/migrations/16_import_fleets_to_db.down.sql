
-- Drop the function allowing to handle harvesting operation of a fleet.
DROP FUNCTION fleet_harvesting(fleet_id uuid);

-- Drop the function allowing to handle deployment of a fleet.
DROP FUNCTION fleet_deployment(fleet_id uuid);

-- Drop the function allowing to perform the transport action of a fleet.
DROP FUNCTION fleet_transport(fleet_id uuid);

-- Drop convenience script allowing to deposit resources on a target.
DROP FUNCTION fleet_deposit_resources(fleet_id uuid);

-- Drop the fleet import function.
DROP FUNCTION create_fleet(fleet json, ships json, resources json, consumption json);
