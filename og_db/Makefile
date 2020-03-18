# Default variables
DB_NAME ?= oglike_db
DB_HOST ?= localhost
DB_PORT ?= 5500
DB_USER ?= og_administrator

# Target allowing to build the docker image for the db.
docker_db:
	docker build -t oglike_db_image .

# Target allowing to start the docker image for the db.
start_db:
	docker run -d --name oglike_db_container -p ${DB_PORT}:5432 oglike_db_image

# Target allowing to connect to the db in the image.
connect:
	psql -U ${DB_USER} -d ${DB_NAME} -h ${DB_HOST} -p ${DB_PORT}

