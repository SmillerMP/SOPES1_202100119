package main

import (
	"RabbitMQ/protofiles/weatherpb"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/grpc"
)

type server struct {
	weatherpb.UnimplementedWeatherServiceServer
}

// Funciones de gRPC
// SendWeatherData es el método que recibe los datos del clima
func (s *server) SendWeatherData(ctx context.Context, req *weatherpb.WeatherListRequest) (*weatherpb.WeatherResponse, error) {

	// Imprimir los datos recibidos
	// for _, weather := range req.GetWeather() {
	// 	fmt.Printf("Recibido: País: %s, Clima: %s, Descripción: %s\n", weather.GetCountry(), weather.GetWeather(), weather.GetDescription())
	// }

	// Iterar sobre los datos recibidos y enviarlos a RabbitMQ
	for _, weather := range req.GetWeather() {
		message := fmt.Sprintf("country: %s, weather: %s, description: %s",
			weather.GetCountry(), weather.GetWeather(), weather.GetDescription())

		publishMessage(message)
	}

	// Imprimir la cantidad de datos enviados
	totalData := len(req.GetWeather())
	actualTime := time.Now().Format("2006-01-02 15:04:05")
	fmt.Printf("Enviados %d datos a RabbitMQ --> [%s]\n", totalData, actualTime)

	return &weatherpb.WeatherResponse{
		Status: "Datos recibidos correctamente",
	}, nil
}

// manejoDeErrores
func handleError(err error, message string) {
	if err != nil {
		log.Panicf("Error: %s: %s", message, err)
	}
}

// Enviar los datos a RabbitMQ
func publishMessage(body string) {

	rabbitmq := os.Getenv("RABBITMQ_SERVER")
	if rabbitmq == "" {
		log.Fatalf("La variable de entorno RABBITMQ_SERVER no está definida")
	}

	// Conectar a RabbitMQ
	conn, err := amqp.Dial("amqp://guest:guest@" + rabbitmq + "/")
	handleError(err, "Error al conectar a RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	handleError(err, "Error al abrir un canal")
	defer ch.Close()

	// Declarar la cola
	q, err := ch.QueueDeclare(
		"weather_data", // name
		false,          // durable
		false,          // delete when unused
		false,          // exclusive
		false,          // no-wait
		nil,            // arguments
	)
	handleError(err, "Error al declarar la cola")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = ch.PublishWithContext(ctx,
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		})

	handleError(err, "Error al publicar el mensaje")
	// log.Printf(" [x] Enviado %s\n", body)
}

func main() {
	// Leer el puerto desde la variable de entorno
	port := os.Getenv("GRPC_PORT")
	if port == "" {
		log.Fatalf("La variable de entorno GRPC_PORT no está definida")
	}

	// Escuchar en el puerto especificado
	listen, err := net.Listen("tcp", ":"+port)
	handleError(err, "Error al escuchar en el puerto")

	// Crear un nuevo servidor gRPC
	gRPCserver := grpc.NewServer()

	weatherpb.RegisterWeatherServiceServer(gRPCserver, &server{})

	fmt.Printf("Servidor gRPC escuchando en el puerto %s...\n", port)

	// Iniciar el servidor gRPC
	if err := gRPCserver.Serve(listen); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}

}
