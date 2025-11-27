package connection

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"pgweb/internal/util"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

// ConnectionHandler exposes HTTP endpoints for working with the database connection.
type ConnectionHandler struct {
	mu         sync.RWMutex
	connection Connection
	db         *sql.DB
}

type Connection struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Database string `json:"database"`
	SSLMode  bool   `json:"ssl_mode"` // TODO: Implement SSL mode handling
}

func (c Connection) ToConnString() string {
	sslMode := "disable"
	if c.SSLMode {
		sslMode = "require"
	}

	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.Username, c.Password, c.Database, sslMode)
}

// Init new ConnectionHandler
func NewConnectionHandler() *ConnectionHandler {
	return &ConnectionHandler{}
}

func (h *ConnectionHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/connect", h.SetConnectionAndConnect)
	mux.HandleFunc("/validate", h.ValidateConnection)
	mux.HandleFunc("/close", h.CloseConnection)
}

// POST method to set the connection params and connect to the database
func (h *ConnectionHandler) SetConnectionAndConnect(w http.ResponseWriter, req *http.Request) {
	// assert that this is a POST request
	if req.Method != http.MethodPost {
		http.Error(w, req.Method+" isn't allowed, only POST calls", http.StatusMethodNotAllowed)
		return
	}

	// decode the request body into the Connection object
	var conn Connection
	dec := util.DecodeJsonBody(req)
	err := dec.Decode(&conn) // this will decode the body into the conn ptr or return an error
	if err != nil {
		http.Error(w, "Failed to decode request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := validateConnection(conn); err != nil {
		http.Error(w, "Invalid connection parameters: "+err.Error(), http.StatusBadRequest)
		return
	}

	// now we need to actually connect to the database
	db, err := sql.Open("postgres", conn.ToConnString())
	if err != nil {
		log.Default().Panicln("Error opening database " + conn.Database + ": " + err.Error())
		http.Error(w, "Failed to open database connection: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// ping the database to ensure the connection is valid
	if err := db.Ping(); err != nil {
		log.Default().Panicln("Error validating to connected database " + conn.Database + ": " + err.Error())
		http.Error(w, "Failed to validate database connection: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Default().Println("Connection to database " + conn.Database + " successful")
	// replace the existing database
	h.mu.Lock()
	if h.db != nil {
		h.db.Close()
	}
	h.db = db
	h.connection = conn
	h.mu.Unlock()

	// return success response
	util.WriteJSON(w, http.StatusAccepted, map[string]any{
		"message": fmt.Sprintf("Succesful connection to the database %s achived!", conn.Database),
	})
}

func (h *ConnectionHandler) ValidateConnection(w http.ResponseWriter, req *http.Request) {
	// asert this is a GET method
	if req.Method != http.MethodGet {
		http.Error(w, "This endpoint accepts only GET calls", http.StatusBadRequest)
		return
	}

	h.mu.RLock()
	db := h.db
	conn := h.connection
	h.mu.RUnlock()

	if db == nil {
		http.Error(w, "No active connection. Call POST /connect first", http.StatusBadRequest)
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

func (h *ConnectionHandler) CloseConnection(w http.ResponseWriter, req *http.Request) {
	// assert this is a POST method
	if req.Method != http.MethodPost {
		http.Error(w, "This endpoint accepts only POST calsl", http.StatusBadRequest)
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
