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
		log.Fatalf("Error al cargar la configuración: %v", err)
	}

	statisticsClient, err := nodeapi.New(cfg.NodeAPIURL, cfg.NodeTimeout, cfg.ServiceAPIKey)
	if err != nil {
		log.Fatalf("Error al configurar el servicio de estadísticas: %v", err)
	}

	app := routes.New(cfg, statisticsClient)
	address := ":" + cfg.Port
	log.Printf("La API de Go está escuchando en %s", address)
	if err := app.Listen(address); err != nil {
		log.Fatalf("Error al iniciar la API de Go: %v", err)
	}
}
