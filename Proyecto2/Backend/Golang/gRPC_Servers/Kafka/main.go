package main

import (
	"Kafka/protofiles/weatherpb"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/segmentio/kafka-go"
	"google.golang.org/grpc"
)

type server struct {
	weatherpb.UnimplementedWeatherServiceServer
	kafkaConn *kafka.Conn
}

func initKafka() *kafka.Conn {
	topic := "weather_data"
	partition := 0

	kafkaServer := os.Getenv("KAFKA_SERVER")
	if kafkaServer == "" {
		log.Fatalf("La variable de entorno RABBITMQ_SERVER no está definida")
	}

	// localhost:9092
	conn, err := kafka.DialLeader(context.Background(), "tcp", kafkaServer, topic, partition)
	if err != nil {
		log.Fatal("failed to dial leader:", err)
	}

	return conn
}

func publishMessage(conn *kafka.Conn, message string) error {
	_, err := conn.WriteMessages(
		kafka.Message{
			Value: []byte(message),
		},
	)
	return err
}

// SendWeatherData es el método que recibe los datos del clima
func (s *server) SendWeatherData(ctx context.Context, req *weatherpb.WeatherListRequest) (*weatherpb.WeatherResponse, error) {

	for _, weather := range req.GetWeather() {
		message := fmt.Sprintf("country: %s, weather: %s, description: %s",
			weather.GetCountry(), weather.GetWeather(), weather.GetDescription())

		err := publishMessage(s.kafkaConn, message)
		if err != nil {
			return nil, fmt.Errorf("error al enviar el mensaje a Kafka: %v", err)
		}
	}

	totalData := len(req.GetWeather())
	actualTime := time.Now().Format("2006-01-02 15:04:05")
	fmt.Printf("Enviado: %d datos a Kafka --> [%s]\n", totalData, actualTime)

	return &weatherpb.WeatherResponse{
		Status: "Datos enviados correctamente",
	}, nil
}

func main() {
	// Leer el puerto desde la variable de entorno
	port := os.Getenv("GRPC_PORT")
	if port == "" {
		log.Fatalf("La variable de entorno GRPC_PORT no está definida")
	}

	kafkaConn := initKafka()
	defer kafkaConn.Close()

	// Escuchar en el puerto especificado
	listen, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// Crear un nuevo servidor gRPC
	gRPCserver := grpc.NewServer()

	weatherService := &server{
		kafkaConn: kafkaConn,
	}

	weatherpb.RegisterWeatherServiceServer(gRPCserver, weatherService)

	fmt.Printf("Servidor gRPC escuchando en el puerto %s...\n", port)

	// Iniciar el servidor gRPC
	if err := gRPCserver.Serve(listen); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}

}
