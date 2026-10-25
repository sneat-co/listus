// Command listusd is the listus backend service.
//
// This is a scaffold: it currently exposes only a health endpoint. Listus
// domain endpoints are intentionally not implemented yet.
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/sneat-co/listus/backend/internal/health"
)

var (
	listenAndServe = http.ListenAndServe
	logFatalf      = log.Fatalf
)

func main() {
	addr := os.Getenv("LISTUS_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	mux := http.NewServeMux()
	mux.Handle("GET /health", health.Handler())

	log.Printf("listusd listening on %s", addr)
	if err := listenAndServe(addr, mux); err != nil {
		logFatalf("listusd failed: %v", err)
	}
}
