package main

import (
	"flownebula/server/internal/api"
	"log"
)

func main() {
	app, cfg := api.New()

	log.Printf("Nebula server running on %s (metrics: %s)", cfg.ServerListenAddr, cfg.ServerMetricsAddr)
	if err := app.Listen(cfg.ServerListenAddr); err != nil {
		log.Fatal(err)
	}
}
