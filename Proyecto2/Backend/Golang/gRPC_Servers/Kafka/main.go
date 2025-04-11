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
	kafkaWriter *kafka.Writer
}

func initKafkaWriter() *kafka.Writer {
	topic := "weather_data"

	kafkaServer := os.Getenv("KAFKA_SERVER")
	if kafkaServer == "" {
		log.Fatalf("La variable de entorno KAFKA_SERVER no está definida")
	}

	return &kafka.Writer{
		Addr:     kafka.TCP(kafkaServer),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}
}

func publishMessage(writer *kafka.Writer, message string) error {
	return writer.WriteMessages(context.Background(),
		kafka.Message{
			Value: []byte(message),
		},
	)
}

// SendWeatherData es el método que recibe los datos del clima
func (s *server) SendWeatherData(ctx context.Context, req *weatherpb.WeatherListRequest) (*weatherpb.WeatherResponse, error) {

	for _, weather := range req.GetWeather() {
		message := fmt.Sprintf("country: %s, weather: %s, description: %s",
			weather.GetCountry(), weather.GetWeather(), weather.GetDescription())

		err := publishMessage(s.kafkaWriter, message)
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

	kafkaWriter := initKafkaWriter()
	defer kafkaWriter.Close()

	// Escuchar en el puerto especificado
	listen, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// Crear un nuevo servidor gRPC
	gRPCserver := grpc.NewServer()

	weatherService := &server{
		kafkaWriter: kafkaWriter,
	}

	weatherpb.RegisterWeatherServiceServer(gRPCserver, weatherService)

	fmt.Printf("Servidor gRPC escuchando en el puerto %s...\n", port)

	// Iniciar el servidor gRPC
	if err := gRPCserver.Serve(listen); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}

}
