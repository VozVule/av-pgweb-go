# Pgweb Backend

Go REST API demo with PostgreSQL via Docker Compose.

## Frontend (HTMX)

A lightweight HTMX UI lives in `frontend/`. Serve `frontend/index.html` (e.g., `cd frontend && python3 -m http.server 5173`) and point it at your Pgweb backend via the "API Base URL" field. The frontend consumes the same REST endpoints documented below for browsing schemas, inspecting tables, and running ad-hoc SQL.

## Migrations with Atlas

1. Install the [Atlas CLI](https://atlasgo.io/getting-started) (`brew install ariga/tap/atlas` on macOS).
2. Export `PGWEB_DATABASE_URL` pointing at the database you want to migrate, for example:
   ```bash
   export PGWEB_DATABASE_URL=postgres://pgweb:pgweb@localhost:5432/pgweb?sslmode=disable
   ```
3. (Optional) set `PGWEB_ATLAS_BIN` or `PGWEB_MIGRATIONS_DIR` if the binary/directory lives elsewhere. Defaults are `atlas` and `./migrations`.
4. Run `./scripts/migrate.sh` to source `.env` and invoke `go run ./cmd/migrate`, or run `atlas migrate apply --env local` manually before starting the API. The `atlas.hcl` file in the repo defines the `local` environment (change it if your migration directory or DSN needs to differ by environment).

The SQL files in `migrations/` are executed in order (`000` → `003`) to create the `company` schema, tables, indexes, and seed data. The API no longer runs migrations automatically—apply them first, then boot the server via `go run ./cmd/api`.

## Running the Application

1. **Apply migrations** (once per database):
   ```bash
   ./scripts/migrate.sh
   # or: PGWEB_DATABASE_URL=... go run ./cmd/migrate
   ```
2. **Start the backend API** (default on `:8080`):
   ```bash
   go run ./cmd/api
   ```
3. **Start the HTMX frontend** (served separately):
   ```bash
   cd frontend
   python3 -m http.server 5173   # or any static file server
   ```
4. Visit the frontend (e.g., http://localhost:5173) and set the “API Base URL” input to your backend (default `http://localhost:8080`). CORS is enabled on the API to allow the browser to call it from another origin.


### Database connection management

Connect to one PostgreSQL database using a DSN/env config.

Create and manage a connection pool.

Pool hides details from the rest of the app.

Current endpoints:

- `POST /connect` — open a connection pool using provided JSON credentials.
- `GET /validate` — ping the active pool to ensure it is still healthy.
- `POST /close` — close the pool and discard stored credentials.
- `GET /schemas` — list all non-system schemas in the connected database.
- `GET /schemas/{schema}/tables` — list tables for a schema.
- `GET /schemas/{schema}/tables/{table}/columns` — list a table's columns plus constraint metadata.
- `GET /schemas/{schema}/tables/{table}/data` — dump table rows (limited to current DB size).
- `GET /schemas/{schema}/views` — list views for a schema.
- `GET /schemas/{schema}/indexes` — list indexes for a schema.
- `POST /query` — execute arbitrary SQL (use with caution!).

### API for metadata + SQL execution

Read-only metadata endpoints:
