package main

import (
	"log"
	"net/http"

	"openhouse-2025-api/internal/config"
	"openhouse-2025-api/internal/http/router"
)

func main() {
	cfg := config.Load()

	r := router.New(cfg)

	addr := ":" + cfg.HTTPPort
	log.Printf("openhouse-2025-api starting on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

