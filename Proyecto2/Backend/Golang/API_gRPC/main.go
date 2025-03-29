package main

import (
	"API_gRPC/protofiles/weatherpb"
	"context"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"google.golang.org/grpc"
)

func main() {
	// Leer el puerto de la API desde el archivo .env
	apiPort := os.Getenv("API_PORT")
	if apiPort == "" {
		log.Fatalf("La variable de entorno API_PORT no está definida")
	}

	// Leer la dirección del servidor gRPC desde el archivo .env
	grpcServer := os.Getenv("GRPC_SERVER")
	if grpcServer == "" {
		log.Fatalf("La variable de entorno GRPC_SERVER no está definida")
	}

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

		// conectar al servidor gRPC
		conn, err := grpc.Dial(grpcServer, grpc.WithInsecure())
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Error connecting to gRPC server",
			})
		}
		defer conn.Close()

		// crear cliente gRPC
		client := weatherpb.NewWeatherServiceClient(conn)

		// convertir los datos a un objeto Weather
		var weatherList []*weatherpb.WeatherRequest
		for _, data := range req {
			weatherList = append(weatherList, &weatherpb.WeatherRequest{
				Country:     data.Country,
				Weather:     data.Weather,
				Description: data.Description,
			})
		}

		// Enviar la solicitud al servidor gRPC
		response, err := client.SendWeatherData(context.Background(), &weatherpb.WeatherListRequest{
			Weather: weatherList, // Asegúrate de usar el campo correcto definido en el .proto
		})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Error al enviar datos al servidor gRPC",
				"details": err.Error(),
			})
		}

		// Manejar la respuesta del servidor gRPC
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "OK API Golang",
			"status":  response.Status, // Incluye el estado de la respuesta del servidor
		})
	})

	// Escuchar en el puerto especificado
	log.Printf("API escuchando en el puerto %s...\n", apiPort)
	app.Listen("0.0.0.0:" + apiPort)
}
