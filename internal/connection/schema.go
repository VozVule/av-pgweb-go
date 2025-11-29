package connection

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"pgweb/internal/util"

	"github.com/lib/pq"
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

// ListTableColumns details columns, types, and constraints for schema.table.
func (h *ConnectionHandler) ListTableColumns(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "This call only supports GET methods", http.StatusMethodNotAllowed)
		return
	}

	db, _, ok := h.ensureDB(w)
	if !ok {
		return
	}

	schemaName := req.PathValue("schema")
	tableName := req.PathValue("table")
	if schemaName == "" || tableName == "" {
		http.Error(w, "schema and table parameters are required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(req.Context(), 2*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, `
		SELECT c.column_name,
		       c.data_type,
		       COALESCE(string_agg(DISTINCT tc.constraint_type, ','), '') AS constraint_types
		FROM information_schema.columns c
		LEFT JOIN information_schema.key_column_usage k
		  ON c.table_schema = k.table_schema
		 AND c.table_name = k.table_name
		 AND c.column_name = k.column_name
		LEFT JOIN information_schema.table_constraints tc
		  ON k.constraint_schema = tc.constraint_schema
		 AND k.constraint_name = tc.constraint_name
		WHERE c.table_schema = $1
		  AND c.table_name = $2
		GROUP BY c.column_name, c.data_type, c.ordinal_position
		ORDER BY c.ordinal_position
	`, schemaName, tableName)
	if err != nil {
		http.Error(w, "Failed fetching column metadata: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	columns := make([]map[string]any, 0)
	for rows.Next() {
		var (
			name          string
			typeName      string
			constraintCSV string
		)
		if err := rows.Scan(&name, &typeName, &constraintCSV); err != nil {
			http.Error(w, "Failed to scan column metadata: "+err.Error(), http.StatusInternalServerError)
			return
		}
		constraints := []string{}
		if constraintCSV != "" {
			constraints = strings.Split(constraintCSV, ",")
		}
		columns = append(columns, map[string]any{
			"name":        name,
			"type":        typeName,
			"constraints": constraints,
		})
	}
	if err := rows.Err(); err != nil {
		http.Error(w, "Failed to iterate columns: "+err.Error(), http.StatusInternalServerError)
		return
	}

	util.WriteJSON(w, http.StatusOK, map[string]any{
		"schema":  schemaName,
		"table":   tableName,
		"columns": columns,
	})
}

// ListTableData returns every row from schema.table, projecting arbitrary columns into JSON.
func (h *ConnectionHandler) ListTableData(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "This call only supports GET methods", http.StatusMethodNotAllowed)
		return
	}

	db, _, ok := h.ensureDB(w)
	if !ok {
		return
	}

	schemaName := req.PathValue("schema")
	tableName := req.PathValue("table")
	if schemaName == "" || tableName == "" {
		http.Error(w, "schema and table parameters are required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second)
	defer cancel()

	// Quote the identifiers to avoid SQL injection via path parameters.
	query := `SELECT * FROM ` + pq.QuoteIdentifier(schemaName) + `.` + pq.QuoteIdentifier(tableName)
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		http.Error(w, "Failed fetching table data: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		http.Error(w, "Failed reading column metadata: "+err.Error(), http.StatusInternalServerError)
		return
	}

	result := make([]map[string]any, 0)
	// Convert each SQL row into a JSON-friendly map by scanning values and pairing them with column names.
	for rows.Next() {
		// values array will hold actual row data
		values := make([]any, len(cols))
		// scanArgs is used for rows.Scan so that it can scan row values into memory via ptrs in scanArgs
		scanArgs := make([]any, len(cols))
		for i := range values {
			scanArgs[i] = &values[i]
		}
		if err := rows.Scan(scanArgs...); err != nil {
			http.Error(w, "Failed scanning row: "+err.Error(), http.StatusInternalServerError)
			return
		}

		rowMap := make(map[string]any, len(cols))
		for i, col := range cols {
			switch v := values[i].(type) {
			case []byte:
				// text/bytea come back as []byte; convert to string for JSON
				rowMap[col] = string(v)
			default:
				rowMap[col] = v
			}
		}
		result = append(result, rowMap)
	}
	if err := rows.Err(); err != nil {
		http.Error(w, "Failed iterating rows: "+err.Error(), http.StatusInternalServerError)
		return
	}

	util.WriteJSON(w, http.StatusOK, map[string]any{
		"schema": schemaName,
		"table":  tableName,
		"rows":   result,
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
