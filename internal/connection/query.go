package connection

import (
	"context"
	"database/sql"
	"net/http"
	"strings"
	"time"

	"pgweb/internal/util"
)

// ExecuteQuery runs arbitrary SQL provided by the caller and returns rows/metadata.
func (h *ConnectionHandler) ExecuteQuery(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "This endpoint accepts only POST calls", http.StatusMethodNotAllowed)
		return
	}

	db, _, ok := h.ensureDB(w)
	if !ok {
		return
	}

	var payload struct {
		Query string `json:"query"`
	}
	dec := util.DecodeJsonBody(req)
	if err := dec.Decode(&payload); err != nil {
		http.Error(w, "Failed to decode request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(payload.Query) == "" {
		http.Error(w, "query is required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(req.Context(), 15*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, payload.Query)
	if err != nil {
		// fall back to Exec for statements that don't return rows
		if execRes, execErr := db.ExecContext(ctx, payload.Query); execErr == nil {
			rowsAffected, _ := execRes.RowsAffected()
			util.WriteJSON(w, http.StatusOK, map[string]any{
				"rows_affected": rowsAffected,
				"result":        "statement executed",
			})
			return
		}
		http.Error(w, "Failed executing query: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer rows.Close()

	cols, data, err := util.RowsToMaps(rows)
	if err != nil {
		http.Error(w, "Failed reading row data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	util.WriteJSON(w, http.StatusOK, map[string]any{
		"columns": cols,
		"rows":    data,
	})
}
