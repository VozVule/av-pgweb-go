package connection

import (
	"database/sql"
	"net/http"
)

// ensureDB returns the current DB pool and connection info, or writes an error response.
func (h *ConnectionHandler) ensureDB(w http.ResponseWriter) (*sql.DB, Connection, bool) {
	h.mu.RLock()
	db := h.db
	conn := h.connection
	h.mu.RUnlock()

	if db == nil {
		http.Error(w, "No active connection. Call POST /connect first", http.StatusBadRequest)
		return nil, Connection{}, false
	}

	return db, conn, true
}
