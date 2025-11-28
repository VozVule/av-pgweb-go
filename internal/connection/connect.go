package connection

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"pgweb/internal/util"

	_ "github.com/lib/pq"
)

// SetConnectionAndConnect handles POST /connect and stores a new pool.
func (h *ConnectionHandler) SetConnectionAndConnect(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, req.Method+" isn't allowed, only POST calls", http.StatusMethodNotAllowed)
		return
	}

	var conn Connection
	dec := util.DecodeJsonBody(req)
	if err := dec.Decode(&conn); err != nil {
		http.Error(w, "Failed to decode request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := validateConnection(conn); err != nil {
		http.Error(w, "Invalid connection parameters: "+err.Error(), http.StatusBadRequest)
		return
	}

	db, err := sql.Open("postgres", conn.ToConnString())
	if err != nil {
		log.Printf("Error opening database %s: %v", conn.Database, err)
		http.Error(w, "Failed to open database connection: "+err.Error(), http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(req.Context(), 2*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		log.Printf("Error validating database %s: %v", conn.Database, err)
		http.Error(w, "Failed to validate database connection: "+err.Error(), http.StatusInternalServerError)
		return
	}

	h.mu.Lock()
	if h.db != nil {
		_ = h.db.Close()
	}
	h.db = db
	h.connection = conn
	h.mu.Unlock()

	util.WriteJSON(w, http.StatusAccepted, map[string]any{
		"message": fmt.Sprintf("Succesful connection to the database %s achived!", conn.Database),
	})
}

// ValidateConnection handles GET /validate to ping the current pool.
func (h *ConnectionHandler) ValidateConnection(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "This endpoint accepts only GET calls", http.StatusBadRequest)
		return
	}

	db, conn, ok := h.ensureDB(w)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(req.Context(), 2*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		http.Error(w, "Database ping failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	util.WriteJSON(w, http.StatusOK, map[string]any{
		"message": fmt.Sprintf("Database %s connection is healthy", conn.Database),
	})
}

// CloseConnection handles POST /close to tear down the pool.
func (h *ConnectionHandler) CloseConnection(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "This endpoint accepts only POST calls", http.StatusBadRequest)
		return
	}

	if _, _, ok := h.ensureDB(w); !ok {
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	if h.db == nil {
		http.Error(w, "No active connection to close", http.StatusBadRequest)
		return
	}

	if err := h.db.Close(); err != nil {
		http.Error(w, "Failed to close database connection: "+err.Error(), http.StatusInternalServerError)
		return
	}

	h.db = nil
	h.connection = Connection{}

	util.WriteJSON(w, http.StatusOK, map[string]any{
		"message": "Database connection closed successfully",
	})
}

func (c Connection) ToConnString() string {
	sslMode := "disable"
	if c.SSLMode {
		sslMode = "require"
	}

	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.Username, c.Password, c.Database, sslMode)
}

func validateConnection(c Connection) error {
	if c.Host == "" {
		return errors.New("host is required")
	}
	if c.Port <= 0 {
		return errors.New("port must be > 0")
	}
	if c.Database == "" {
		return errors.New("database is required")
	}
	return nil
}
