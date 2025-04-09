package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/valkey-io/valkey-go"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

// Función para procesar un mensaje
func processMessage(body []byte, clientValkey valkey.Client) {
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

	err := clientValkey.Do(ctx, clientValkey.B().Hset().Key("contador:paises").FieldValue().FieldValue("Pais", "Valor").Build()).Error()
	if err != nil {
		log.Printf("Error al inicializar el hash en Valkey: %v", err)
	}

	// Incrementar el contador de la variable country en Valkey
	err = clientValkey.Do(ctx, clientValkey.B().Hincrby().Key("contador:paises").Field(country).Increment(1).Build()).Error()
	if err != nil {
		log.Printf("Error al incrementar el contador para el país %s: %v", country, err)
	}

	err = clientValkey.Do(ctx, clientValkey.B().Incr().Key("total").Build()).Error()
	if err != nil {
		log.Printf("Error al incrementar el contador total: %v", err)
	}
}

// Función para crear el cliente de RabbitMQ y el canal
func createRabbitMQConnection(rabbitmq string) (*amqp.Connection, *amqp.Channel, error) {
	conn, err := amqp.Dial("amqp://guest:guest@" + rabbitmq + "/")
	if err != nil {
		return nil, nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, nil, err
	}

	return conn, ch, nil
}

func main() {
	// Leer la variable de entorno para RabbitMQ
	rabbitmq := os.Getenv("RABBITMQ_SERVER")
	if rabbitmq == "" {
		log.Fatalf("La variable de entorno RABBITMQ_SERVER no está definida")
	}

	// Leer la variable de entorno para el número de goroutines
	goRutines := os.Getenv("NO_GOROUTINES")
	if goRutines == "" {
		log.Fatalf("La variable de entorno NO_GOROUTINES no está definida")
	}

	valkeyServer := os.Getenv("VALKEY_SERVER")
	if valkeyServer == "" {
		log.Fatalf("La variable de entorno NO_GOROUTINES no está definida")
	}

	// Convertir el valor de la variable de entorno a int
	numGoRutines, err := strconv.Atoi(goRutines)
	if err != nil {
		log.Fatalf("Error al convertir la variable de entorno NO_GOROUTINES a int: %v", err)
	}

	// configurar valkey
	clientValkey, err := valkey.NewClient(valkey.ClientOption{InitAddress: []string{valkeyServer}})
	if err != nil {
		log.Fatalf("Error al crear el cliente de valkey: %v", err)
	}
	defer clientValkey.Close()
	log.Printf("Conectando a Valkey en %s", "valkey:6379")

	// Conectar a RabbitMQ
	conn, ch, err := createRabbitMQConnection(rabbitmq)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()
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
	failOnError(err, "Failed to declare a queue")

	// Consumir mensajes de la cola
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	// Crear un grupo de espera para sincronizar las goroutines
	var wg sync.WaitGroup

	// Canal para enviar mensajes a los workers
	messageChannel := make(chan []byte)

	// Crear workers
	for i := 0; i < numGoRutines; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for body := range messageChannel {
				// log.Printf("Worker %d procesando mensaje", workerID)
				processMessage(body, clientValkey)
			}
		}(i)
	}

	// Enviar mensajes al canal para que los procesen los workers
	go func() {
		for d := range msgs {
			messageChannel <- d.Body
		}
		close(messageChannel) // Cerrar el canal cuando no haya más mensajes
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	wg.Wait() // Esperar a que todos los workers terminen
}
