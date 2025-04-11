package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func processMessage(body []byte, clientRedis *redis.Client) {
	log.Printf("Procesando mensaje: %s", body)

	message := string(body)
	var country string
	parts := strings.Split(message, ",")
	for _, part := range parts {
		if strings.Contains(part, "country:") {
			country = strings.Split(part, ":")[1]
			country = strings.TrimSpace(country)
			break
		}
	}

	if country == "" {
		log.Println("No se encontró el país en el mensaje")
		return
	}

	ctx := context.Background()

	err := clientRedis.HSet(ctx, "contador:paises", "Pais", "Valor").Err()
	if err != nil {
		log.Printf("Error al establecer el país %s: %v", country, err)
	}

	err = clientRedis.HIncrBy(ctx, "contador:paises", country, 1).Err()
	if err != nil {
		log.Printf("Error al incrementar el contador para el país %s: %v", country, err)
	}

	err = clientRedis.Incr(ctx, "total").Err()
	if err != nil {
		log.Printf("Error al incrementar el contador total: %v", err)
	}
}

func main() {
	goRoutines := os.Getenv("NO_GOROUTINES")
	if goRoutines == "" {
		log.Fatalf("La variable de entorno NO_GOROUTINES no está definida")
	}

	redisServer := os.Getenv("REDIS_SERVER")
	if redisServer == "" {
		log.Fatalf("La variable de entorno REDIS_SERVER no está definida")
	}

	kafkaServer := os.Getenv("KAFKA_SERVER")
	if kafkaServer == "" {
		log.Fatalf("La variable de entorno KAFKA_SERVER no está definida")
	}

	numGoRoutines, err := strconv.Atoi(goRoutines)
	failOnError(err, "Error al convertir NO_GOROUTINES")

	clientRedis := redis.NewClient(&redis.Options{
		Addr:     redisServer,
		Password: "",
		DB:       0,
		Protocol: 2,
	})
	defer clientRedis.Close()

	log.Printf("Conectando a Redis en %s", redisServer)

	// Crear el lector de Kafka
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{kafkaServer},
		Topic:    "weather_data",
		GroupID:  "weather_consumers",
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})
	defer reader.Close()

	var wg sync.WaitGroup
	messageChannel := make(chan []byte)

	// Crear workers
	for i := 0; i < numGoRoutines; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for msg := range messageChannel {
				processMessage(msg, clientRedis)
			}
		}(i)
	}

	log.Println("Esperando mensajes de Kafka...")

	// Loop infinito para leer mensajes
	for {
		m, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Printf("Error al leer mensaje de Kafka: %v", err)
			break
		}
		messageChannel <- m.Value
	}

	close(messageChannel)
	wg.Wait()
}
