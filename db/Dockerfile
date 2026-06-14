FROM postgres:16-alpine
COPY internal/db/migrations/001_initial.sql /docker-entrypoint-initdb.d/001_initial.sql
COPY internal/db/migrations/002_events_cli.sql /docker-entrypoint-initdb.d/002_events_cli.sql
