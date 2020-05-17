-- Drop the function performing the update of debris fields after harvesting.
DROP FUNCTION fleet_harvesting_success(fleet_id uuid, debris_id uuid, resources json, dispersed text, gathered text);

-- Drop the function allowing to notify a colonization failure.
DROP FUNCTION fleet_colonization_failed(fleet_id uuid);

-- Drop the function allowing to notify a colonization success.
DROP FUNCTION fleet_colonization_success(fleet_id uuid, planet json, resources json);

-- Drop the function allowing to handle deployment of a fleet.
DROP FUNCTION fleet_deployment(fleet_id uuid);

-- Drop the function allowing to perform the transport action of a fleet.
DROP FUNCTION fleet_transport(fleet_id uuid);

-- Drop the function allowing to make a fleet return to its base.
DROP FUNCTION fleet_return_to_base(fleet_id uuid);

-- Drop the function allowing to update the fleet's position in the actions queue.
DROP FUNCTION fleet_update_to_return_time(fleet_id uuid);

-- Drop the function allowing to delete a fleet from the DB.
DROP FUNCTION fleet_deletion(fleet_id uuid);

-- Drop the function allowing to deploy ships on a location.
DROP FUNCTION fleet_ships_deployment(fleet_id uuid, target_id uuid, target_kind text);

-- Drop the function allowing to deposit resources on a location.
DROP FUNCTION fleet_deposit_resources(fleet_id uuid, target_id uuid, target_kind text);

-- Drop the fleet import function.
DROP FUNCTION create_fleet(fleet json, ships json, resources json, consumption json);

-- Drop the message creation function.
DROP FUNCTION create_message_for(player_id uuid, message_name text, VARIADIC args text[]);