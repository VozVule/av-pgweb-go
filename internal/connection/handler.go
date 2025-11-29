package connection

import (
	"database/sql"
	"net/http"
	"sync"
)

// ConnectionHandler stores configuration and state for DB-related endpoints.
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
	SSLMode  bool   `json:"ssl_mode"`
}

// NewConnectionHandler creates a handler with no active DB pool.
func NewConnectionHandler() *ConnectionHandler {
	return &ConnectionHandler{}
}

// Register wires HTTP endpoints to the mux.
func (h *ConnectionHandler) Register(mux *http.ServeMux) {
    mux.HandleFunc("/connect", h.SetConnectionAndConnect)
    mux.HandleFunc("/validate", h.ValidateConnection)
    mux.HandleFunc("/close", h.CloseConnection)
	mux.HandleFunc("/schemas", h.ListSchemas)
	mux.HandleFunc("/schemas/{schema}/tables", h.ListTablesForSchema)
	mux.HandleFunc("/schemas/{schema}/tables/{table}/columns", h.ListTableColumns)
	mux.HandleFunc("/schemas/{schema}/tables/{table}/data", h.ListTableData)
	mux.HandleFunc("/schemas/{schema}/views", h.ListViewsForSchema)
	mux.HandleFunc("/schemas/{schema}/indexes", h.ListIndexesForSchema)
}
