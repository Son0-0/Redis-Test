FROM postgres:14-alpine

ENV POSTGRES_USER root

ENV POSTGRES_PASSWORD test

COPY *.sql /docker-entrypoint-initdb.d/