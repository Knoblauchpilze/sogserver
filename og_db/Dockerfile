# This link helped a lot in determining what is needed and what
# is not. It also helped building the rest of the stack to set
# and create the db:
# https://medium.com/@wkrzywiec/database-in-a-docker-container-how-to-start-and-whats-it-about-5e3ceea77e50
FROM postgres

ENV POSTGRES_DB oglike_db
ENV POSTGRES_USER og_administrator
ENV POSTGRES_PASSWORD Ww76hQWbbt7zi2ItM6cNo4YYT

COPY initdb.sql /docker-entrypoint-initdb.d/
