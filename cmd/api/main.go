package main

import (
	"log"
	"net/http"
	"pgweb/internal/connection"
)

func main() {
	mux := http.NewServeMux()

	connectionHandler := connection.NewConnectionHandler()
	connectionHandler.Register(mux)

	addr := ":8080"
	log.Printf("REST API listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
