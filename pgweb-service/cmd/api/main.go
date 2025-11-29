package main

import (
	"log"
	"net/http"

	"pgweb/internal/connection"
)

func main() {
	addr := ":8080"
	mux := http.NewServeMux()

	connectionHandler := connection.NewConnectionHandler()
	connectionHandler.Register(mux)

	log.Printf("REST API listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
