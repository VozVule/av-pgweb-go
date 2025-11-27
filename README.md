# Pgweb Backend

Go REST API demo with PostgreSQL via Docker Compose.


### Database connection management

Connect to one PostgreSQL database using a DSN/env config.

Create and manage a connection pool.

Pool hides details from the rest of the app.

Current endpoints:

- `POST /connect` — open a connection pool using provided JSON credentials.
- `GET /validate` — ping the active pool to ensure it is still healthy.
- `POST /close` — close the pool and discard stored credentials.

### API for metadata + SQL execution

Read-only metadata endpoints:

GET

/schemas – list schemas

/schemas/{schema}/tables – list tables in a schema

/schemas/{schema}/views – list views in a schema

/schemas/{schema}/indexes – list indexes (maybe per table)

/tables/{table}/columns – list columns + constraints

POST
/query – run a sanitized SQL query and return rows.
