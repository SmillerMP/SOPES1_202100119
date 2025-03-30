package main

import (
	"log"
	"os"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

// Función para procesar un mensaje
func processMessage(body []byte) {
	log.Printf("Procesando mensaje: %s", body)
	// Simular procesamiento (puedes reemplazar esto con lógica real)
	// time.Sleep(1 * time.Second)
}

func main() {
	// Leer la variable de entorno para RabbitMQ
	rabbitmq := os.Getenv("RABBITMQ_SERVER")
	if rabbitmq == "" {
		log.Fatalf("La variable de entorno RABBITMQ_SERVER no está definida")
	}

	// Conectar a RabbitMQ
	conn, err := amqp.Dial("amqp://guest:guest@" + rabbitmq + "/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
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

	// Número de workers (goroutines) para procesar mensajes
	numWorkers := 10

	// Canal para enviar mensajes a los workers
	messageChannel := make(chan []byte)

	// Crear workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for body := range messageChannel {
				// log.Printf("Worker %d procesando mensaje", workerID)
				processMessage(body)
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
