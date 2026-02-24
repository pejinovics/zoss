package main

import (
	"log"
	"os"
	"strconv"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	rabbitURL := getEnv("RABBITMQ_URL", "amqp://guest:guest@rabbitmq:5672/")
	msgCount, _ := strconv.Atoi(getEnv("MESSAGE_COUNT", "1000"))
	msgSizeMB, _ := strconv.Atoi(getEnv("MESSAGE_SIZE_MB", "1"))
	msgSize := msgSizeMB * 1024 * 1024

	conn, err := amqp.Dial(rabbitURL)
	failOnError(err, "failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "failed to open channel")
	defer ch.Close()

	log.Printf("Sending %d messages of %d MB", msgCount, msgSizeMB)
	start := time.Now()

	for i := 0; i < msgCount; i++ {
		body := make([]byte, msgSize)
		for j := range body {
			body[j] = byte(j % 256)
		}

		err = ch.Publish(
			"",
			"snapshots",
			false,
			false,
			amqp.Publishing{
				DeliveryMode: amqp.Persistent,
				ContentType:  "application/octet-stream",
				Body:         body,
			})
		if err != nil {
			log.Printf("error publishing message %d: %s", i, err)
			break
		}
		if i%100 == 0 {
			log.Printf("sent %d messages", i)
		}
	}

	log.Printf("Finished sending %d messages in %s", msgCount, time.Since(start))
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
