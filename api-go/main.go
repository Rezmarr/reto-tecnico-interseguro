package main

import (
	"log"

	"example.com/reto-tecnico/api-go/config"
	"example.com/reto-tecnico/api-go/integrations/nodeapi"
	"example.com/reto-tecnico/api-go/routes"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load configuration: %v", err)
	}

	statisticsClient, err := nodeapi.New(cfg.NodeAPIURL, cfg.NodeTimeout, cfg.ServiceAPIKey)
	if err != nil {
		log.Fatalf("configure statistics service: %v", err)
	}

	app := routes.New(cfg, statisticsClient)
	address := ":" + cfg.Port
	log.Printf("Go API listening on %s", address)
	if err := app.Listen(address); err != nil {
		log.Fatalf("start Go API: %v", err)
	}
}
