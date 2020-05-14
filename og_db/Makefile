# Default variables
DB_NAME ?= oglike_db
DB_HOST ?= localhost
DB_PORT ?= 5500
DB_USER ?= og_administrator
DB_PASSWORD ?= Ww76hQWbbt7zi2ItM6cNo4YYT
DEMIGRATION_STEPS ?= 7

# Target allowing to build the docker image for the db.
docker_db:
	docker build -t oglike_db_image .

# Target allowing to remove any existing docker image of the db.
remove_db: stop_db
	docker rm oglike_db_container
	docker image rm oglike_db_image

# Target allowing to create the docker image for the db.
create_db:
	docker run -d --name oglike_db_container -p ${DB_PORT}:5432 oglike_db_image

# Target allowing to start the docker image for the db.
start_db:
	docker start oglike_db_container

# Target allowing to stop the docker image fot the db.
stop_db:
	docker stop oglike_db_container

# Target allowing to connect to the db in the image.
connect:
	psql -U ${DB_USER} -d ${DB_NAME} -h ${DB_HOST} -p ${DB_PORT}

# Target allowing to migrate the db to latest defined step.
.PHONY: migrate
migrate:
	migrate -path ${PWD}/migrations -database postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable up

# Target allowing to migrate the db to initial state.
demigrate:
	migrate -path ${PWD}/migrations -database postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable down

# Demigrate only the last step of the DB creation.
demigrateO:
	migrate -path ${PWD}/migrations -database postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable down ${DEMIGRATION_STEPS}
