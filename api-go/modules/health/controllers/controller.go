package controllers

import (
	"example.com/reto-tecnico/api-go/modules/health/contracts"
	"github.com/gofiber/fiber/v3"
)

func Get(c fiber.Ctx) error { return c.JSON(contracts.Response{Status: "ok", Service: "api-go"}) }
