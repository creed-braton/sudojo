FROM postgres:18.0-alpine3.22

COPY init.sql /docker-entrypoint-initdb.d/ 

RUN chmod 644 /docker-entrypoint-initdb.d/init.sql
