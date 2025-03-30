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
	rabbitConn *amqp.Connection
	rabbitCh   *amqp.Channel
}

// Inicializar la conexión y canal a RabbitMQ
func initRabbitMQ() (*amqp.Connection, *amqp.Channel) {
	rabbitmq := os.Getenv("RABBITMQ_SERVER")
	if rabbitmq == "" {
		log.Fatalf("La variable de entorno RABBITMQ_SERVER no está definida")
	}

	conn, err := amqp.Dial("amqp://guest:guest@" + rabbitmq + "/")
	if err != nil {
		log.Fatalf("Error al conectar a RabbitMQ: %s", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Error al abrir un canal: %s", err)
	}

	// Declarar la cola
	_, err = ch.QueueDeclare(
		"weather_data", // name
		false,          // durable
		false,          // delete when unused
		false,          // exclusive
		false,          // no-wait
		nil,            // arguments
	)
	if err != nil {
		log.Fatalf("Error al declarar la cola: %s", err)
	}

	return conn, ch
}

// Funciones de gRPC
func (s *server) SendWeatherData(ctx context.Context, req *weatherpb.WeatherListRequest) (*weatherpb.WeatherResponse, error) {
	// Iterar sobre los datos recibidos y enviarlos a RabbitMQ
	for _, weather := range req.GetWeather() {
		message := fmt.Sprintf("country: %s, weather: %s, description: %s",
			weather.GetCountry(), weather.GetWeather(), weather.GetDescription())

		// Publicar mensaje en RabbitMQ
		err := publishMessage(s.rabbitCh, message)
		if err != nil {
			return nil, fmt.Errorf("error al enviar datos a RabbitMQ: %v", err)
		}
	}

	// Imprimir la cantidad de datos enviados
	totalData := len(req.GetWeather())
	actualTime := time.Now().Format("2006-01-02 15:04:05")
	fmt.Printf("Enviados %d datos a RabbitMQ --> [%s]\n", totalData, actualTime)

	return &weatherpb.WeatherResponse{
		Status: "Datos recibidos correctamente",
	}, nil
}

// Enviar los datos a RabbitMQ
func publishMessage(ch *amqp.Channel, body string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := ch.PublishWithContext(ctx,
		"",             // exchange
		"weather_data", // routing key
		false,          // mandatory
		false,          // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		})
	return err
}

func main() {
	// Leer el puerto desde la variable de entorno
	port := os.Getenv("GRPC_PORT")
	if port == "" {
		log.Fatalf("La variable de entorno GRPC_PORT no está definida")
	}

	// Inicializar conexión y canal de RabbitMQ
	rabbitConn, rabbitCh := initRabbitMQ()
	defer rabbitConn.Close()
	defer rabbitCh.Close()

	// Escuchar en el puerto especificado
	listen, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Error al escuchar en el puerto: %s", err)
	}

	// Crear un nuevo servidor gRPC
	gRPCserver := grpc.NewServer()

	// Registrar el servicio gRPC con la conexión a RabbitMQ
	weatherService := &server{
		rabbitConn: rabbitConn,
		rabbitCh:   rabbitCh,
	}

	weatherpb.RegisterWeatherServiceServer(gRPCserver, weatherService)

	fmt.Printf("Servidor gRPC escuchando en el puerto %s...\n", port)

	// Iniciar el servidor gRPC
	if err := gRPCserver.Serve(listen); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
