package main

import (
	"RabbitMQ/protofiles/weatherpb"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"google.golang.org/grpc"
)

type server struct {
	weatherpb.UnimplementedWeatherServiceServer
}

// SendWeatherData es el método que recibe los datos del clima
func (s *server) SendWeatherData(ctx context.Context, req *weatherpb.WeatherListRequest) (*weatherpb.WeatherResponse, error) {

	// Imprimir los datos recibidos
	// for _, weather := range req.GetWeather() {
	// 	fmt.Printf("Recibido: País: %s, Clima: %s, Descripción: %s\n", weather.GetCountry(), weather.GetWeather(), weather.GetDescription())
	// }

	actualTime := time.Now().Format("2006-01-02 15:04:05")

	totalData := len(req.GetWeather())
	fmt.Printf("Recibido: %d datos  --> [%s]\n", totalData, actualTime)

	return &weatherpb.WeatherResponse{
		Status: "Datos recibidos correctamente",
	}, nil
}

func main() {
	// Leer el puerto desde la variable de entorno
	port := os.Getenv("GRPC_PORT")
	if port == "" {
		log.Fatalf("La variable de entorno GRPC_PORT no está definida")
	}

	// Escuchar en el puerto especificado
	listen, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// Crear un nuevo servidor gRPC
	gRPCserver := grpc.NewServer()

	weatherpb.RegisterWeatherServiceServer(gRPCserver, &server{})

	fmt.Printf("Servidor gRPC escuchando en el puerto %s...\n", port)

	// Iniciar el servidor gRPC
	if err := gRPCserver.Serve(listen); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}

}
