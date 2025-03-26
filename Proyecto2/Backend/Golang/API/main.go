package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	app := fiber.New()

	app.Use(cors.New())

	// logger del middleware
	app.Use(logger.New())

	app.Get("/", func(c *fiber.Ctx) error {
		// retornar mensaje de bienvenida json
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Hello, World!",
		})

	})

	app.Post("/weather", func(c *fiber.Ctx) error {
		// retornar mensaje de bienvenida json
		type Request struct {
			Country     string `json:"country"` // path del disco
			Weather     string `json:"weather"` // ruta de la carpeta
			Description string `json:"description"`
		}

		var req []Request
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Error parsing request body",
			})
		}

		// retornar mensaje de ok
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "OK API Golang"})
	})

	app.Listen("0.0.0.0:8010")
}
