
-- Create the table which will reference actions that are outstanding in the server.
CREATE TABLE actions_queue (
  action uuid NOT NULL,
  completion_time TIMESTAMP WITH TIME ZONE NOT NULL,
  type text NOT NULL,
  PRIMARY KEY (action)
);
