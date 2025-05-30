package main

import (
	"API_gRPC/protofiles/weatherpb"
	"context"
	"log"
	"os"
	"sync"

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
		type Request struct {
			Country     string `json:"country"`
			Weather     string `json:"weather"`
			Description string `json:"description"`
		}

		// Leer las direcciones de los servidores gRPC desde las variables de entorno
		grpcServerKafka := os.Getenv("GRPC_SERVER_KAFKA")
		grpcServerRabbit := os.Getenv("GRPC_SERVER_RABBITMQ")

		if grpcServerKafka == "" || grpcServerRabbit == "" {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Las variables de entorno GRPC_SERVER_KAFKA o GRPC_SERVER_RABBITMQ no están definidas",
			})
		}
		var req []Request
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Error parsing request body",
			})
		}

		// Crear la lista de datos a enviar
		var weatherList []*weatherpb.WeatherRequest
		for _, data := range req {
			weatherList = append(weatherList, &weatherpb.WeatherRequest{
				Country:     data.Country,
				Weather:     data.Weather,
				Description: data.Description,
			})
		}

		// Conectar al primer servidor gRPC
		connKafka, err := grpc.Dial(grpcServerKafka, grpc.WithInsecure())
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Error connecting to gRPC server Kafka",
			})
		}
		defer connKafka.Close()

		client1 := weatherpb.NewWeatherServiceClient(connKafka)

		// Conectar al segundo servidor gRPC
		connRabbit, err := grpc.Dial(grpcServerRabbit, grpc.WithInsecure())
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Error connecting to gRPC server RabbitMQ",
			})
		}
		defer connRabbit.Close()

		client2 := weatherpb.NewWeatherServiceClient(connRabbit)

		// Enviar los datos a ambos servidores en paralelo
		var wg sync.WaitGroup
		var kafkaErr, rabbitErr error

		chunkSize := (len(weatherList) + 9) / 10

		wg.Add(10)
		for i := 0; i < 10; i++ {
			start := i * chunkSize
			end := start + chunkSize
			if end > len(weatherList) {
				end = len(weatherList)
			}

			go func(chunk []*weatherpb.WeatherRequest) {
				defer wg.Done()
				_, kafkaErr = client1.SendWeatherData(context.Background(), &weatherpb.WeatherListRequest{
					Weather: chunk,
				})
			}(weatherList[start:end])
		}

		wg.Add(10)
		for i := 0; i < 10; i++ {
			start := i * chunkSize
			end := start + chunkSize
			if end > len(weatherList) {
				end = len(weatherList)
			}

			go func(chunk []*weatherpb.WeatherRequest) {
				defer wg.Done()
				_, rabbitErr = client2.SendWeatherData(context.Background(), &weatherpb.WeatherListRequest{
					Weather: chunk,
				})
			}(weatherList[start:end])
		}

		wg.Wait()

		// Manejar errores y respuestas
		if kafkaErr != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Error al enviar datos al servidor gRPC Kafka",
				"details": kafkaErr.Error(),
			})
		}

		if rabbitErr != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Error al enviar datos al servidor gRPC RabbitMQ",
				"details": rabbitErr.Error(),
			})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Datos enviados a ambos servidores gRPC",
		})
	})

	// Escuchar en el puerto especificado
	log.Printf("API escuchando en el puerto %s...\n", apiPort)
	app.Listen("0.0.0.0:" + apiPort)
}
