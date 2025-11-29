package main

import (
	"log"
	"net/http"

	"pgweb-service/internal/connection"
	"pgweb-service/internal/http"
)

func main() {
	addr := ":8080"
	mux := http.NewServeMux()

	connectionHandler := connection.NewConnectionHandler()
	connectionHandler.Register(mux)

	log.Printf("REST API listening on %s", addr)
	if err := http.ListenAndServe(addr, httpx.WithCORS(mux)); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
