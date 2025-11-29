# Pgweb Frontend (HTMX)

Single-page interface built with [HTMX](https://htmx.org/) that talks to the Pgweb Go backend.

## Features

- Connect to a PostgreSQL instance using `/connect`, validate health via `/validate`
- Browse schemas, tables, views, indexes, and table metadata with `/schemas` + related endpoints
- Preview table rows via `/schemas/{schema}/tables/{table}/data`
- Run ad-hoc SQL queries via `/query`

## Development

1. Ensure the Go backend is running and accessible (default `http://localhost:8080`). If serving from another host, update the API base input at the top of the page.
2. Serve the static files from this directory. Examples:
   ```bash
   cd frontend
   python3 -m http.server 5173
   # or
   npx serve .
   ```
3. Open the served URL in a browser (e.g., http://localhost:5173). Use the controls to connect and explore the database.

The UI relies solely on HTMX + vanilla JS for rendering. All network calls are made directly to the backend REST endpoints; ensure CORS is permitted if hosting frontend and backend on different origins.
