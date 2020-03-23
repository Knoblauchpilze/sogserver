
-- Drop the table referencing dependencies between defenses and technologies.
DROP TABLE tech_tree_defenses_vs_technologies;

-- Drop the table referencing dependencies between defenses and buildings.
DROP TABLE tech_tree_defenses_vs_buildings;

-- Drop the table referencing dependencies between ships and technologies.
DROP TABLE tech_tree_ships_vs_technologies;

-- Drop the table referencing dependencies between ships and buildings.
DROP TABLE tech_tree_ships_vs_buildings;

-- Drop the table referencing dependencies between technologies and buildings.
DROP TABLE tech_tree_technologies_vs_buildings;

-- Drop the table referencing dependencies between buildings and technologies.
DROP TABLE tech_tree_buildings_vs_technologies;

-- Drop the table referencing dependencies between technologies.
DROP TABLE tech_tree_technologies_dependencies;

-- Drop the table referencing dependencies between buildings.
DROP TABLE tech_tree_buildings_dependencies;
