package routes

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/recover"

	"example.com/reto-tecnico/api-go/config"
	"example.com/reto-tecnico/api-go/integrations/nodeapi"
	authcontroller "example.com/reto-tecnico/api-go/modules/auth/controllers"
	factorizationcontroller "example.com/reto-tecnico/api-go/modules/factorization/controllers"
	healthcontroller "example.com/reto-tecnico/api-go/modules/health/controllers"
	sharedhttp "example.com/reto-tecnico/api-go/shared/http"
)

func New(cfg config.Config, statisticsClient *nodeapi.Client) *fiber.App {
	app := fiber.New(fiber.Config{
		BodyLimit:    cfg.BodyLimit,
		ErrorHandler: sharedhttp.ErrorHandler,
	})
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: cfg.FrontendOrigins,
		AllowMethods: []string{"GET", "POST", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"},
	}))
	app.Use(sharedhttp.RequestIDMiddleware())

	factorizationController := factorizationcontroller.New(cfg, statisticsClient)
	auth := authcontroller.New(cfg)
	app.Get("/health", healthcontroller.Get)
	app.Post("/api/v1/auth/token", auth.Issue)
	app.Post("/api/v1/factorizations", auth.RequireToken, factorizationController.Create)
	return app
}
