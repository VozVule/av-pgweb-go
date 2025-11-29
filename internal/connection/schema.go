package connection

import (
	"context"
	"log"
	"net/http"
	"time"

	"pgweb/internal/util"
)

// ListSchemas handles GET /schemas to show non-system schemas.
func (h *ConnectionHandler) ListSchemas(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "This endpoint accepts only GET calls", http.StatusMethodNotAllowed)
		return
	}

	db, _, ok := h.ensureDB(w)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(req.Context(), 2*time.Second)
	defer cancel()

	// fetch the schemas from the db
	rows, err := db.QueryContext(ctx, `
        SELECT nspname
        FROM pg_namespace
        WHERE nspname NOT LIKE 'pg_%'
          AND nspname <> 'information_schema'
        ORDER BY nspname
    `)
	if err != nil {
		http.Error(w, "Failed to list schemas: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// append them to the list
	schemas := make([]string, 0)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			http.Error(w, "Failed to scan schema row: "+err.Error(), http.StatusInternalServerError)
			return
		}
		schemas = append(schemas, name)
	}
	if err := rows.Err(); err != nil {
		http.Error(w, "Failed to iterate schemas: "+err.Error(), http.StatusInternalServerError)
		return
	}

	util.WriteJSON(w, http.StatusOK, map[string]any{
		"schemas": schemas,
		"count":   len(schemas),
	})
}

func (h *ConnectionHandler) ListTablesForSchema(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "This call only supports GET methods", http.StatusMethodNotAllowed)
		return
	}

	db, _, ok := h.ensureDB(w)
	if !ok {
		log.Default().Println("Aborting execution due to unconfigured conifguration")
		return
	}
	ctx, cancel := context.WithTimeout(req.Context(), 2*time.Second)
	defer cancel()

	// query the database for actual tables from the requested schema
	schemaName := req.PathValue("schema")
	rows, err := db.QueryContext(ctx, `
		select tablename
		from pg_catalog.pg_tables
		where schemaname = $1
		order by tablename
	`, schemaName,
	)

	if err != nil {
		http.Error(w, "Failed fetching the table names", http.StatusInternalServerError)
		log.Default().Println("Error occured while fetching tables for schema: " + err.Error())
		return
	}

	tables := make([]string, 0)
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			http.Error(w, "Failed to scan table name: "+err.Error(), http.StatusInternalServerError)
			return
		}
		tables = append(tables, table)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "Failed to iterate tables: "+err.Error(), http.StatusInternalServerError)
		return
	}

	util.WriteJSON(w, http.StatusOK, map[string]any{
		"schema": schemaName,
		"tables": tables,
		"count":  len(tables),
	})
}

// ListViewsForSchema enumerates views in a schema.
func (h *ConnectionHandler) ListViewsForSchema(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "This call only supports GET methods", http.StatusMethodNotAllowed)
		return
	}

	db, _, ok := h.ensureDB(w)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(req.Context(), 2*time.Second)
	defer cancel()

	schemaName := req.PathValue("schema")
	rows, err := db.QueryContext(ctx, `
		select viewname
		from pg_catalog.pg_views
		where schemaname = $1
		order by viewname
	`, schemaName,
	)
	if err != nil {
		http.Error(w, "Failed fetching the view names", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	views := make([]string, 0)
	for rows.Next() {
		var view string
		if err := rows.Scan(&view); err != nil {
			http.Error(w, "Failed to scan view name: "+err.Error(), http.StatusInternalServerError)
			return
		}
		views = append(views, view)
	}
	if err := rows.Err(); err != nil {
		http.Error(w, "Failed to iterate views: "+err.Error(), http.StatusInternalServerError)
		return
	}

	util.WriteJSON(w, http.StatusOK, map[string]any{
		"schema": schemaName,
		"views":  views,
		"count":  len(views),
	})
}

// ListIndexesForSchema enumerates indexes in a schema.
func (h *ConnectionHandler) ListIndexesForSchema(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "This call only supports GET methods", http.StatusMethodNotAllowed)
		return
	}

	db, _, ok := h.ensureDB(w)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(req.Context(), 2*time.Second)
	defer cancel()

	schemaName := req.PathValue("schema")
	rows, err := db.QueryContext(ctx, `
		select indexname, tablename
		from pg_catalog.pg_indexes
		where schemaname = $1
		order by indexname
	`, schemaName,
	)
	if err != nil {
		http.Error(w, "Failed fetching the index names", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	indexes := make([]map[string]string, 0)
	for rows.Next() {
		var name, table string
		if err := rows.Scan(&name, &table); err != nil {
			http.Error(w, "Failed to scan index row: "+err.Error(), http.StatusInternalServerError)
			return
		}
		indexes = append(indexes, map[string]string{
			"index": name,
			"table": table,
		})
	}
	if err := rows.Err(); err != nil {
		http.Error(w, "Failed to iterate indexes: "+err.Error(), http.StatusInternalServerError)
		return
	}

	util.WriteJSON(w, http.StatusOK, map[string]any{
		"schema":  schemaName,
		"indexes": indexes,
		"count":   len(indexes),
	})
}
