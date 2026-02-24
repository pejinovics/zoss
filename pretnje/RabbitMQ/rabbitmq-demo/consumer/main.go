package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	rabbitURL := getEnv("RABBITMQ_URL", "amqp://guest:guest@rabbitmq:5672/")
	conn, err := amqp.Dial(rabbitURL)
	failOnError(err, "failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "failed to open channel")
	defer ch.Close()

	minioEndpoint := getEnv("MINIO_ENDPOINT", "minio:9000")
	minioAccessKey := getEnv("MINIO_ACCESS_KEY", "minioadmin")
	minioSecretKey := getEnv("MINIO_SECRET_KEY", "minioadmin")
	minioClient, err := minio.New(minioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(minioAccessKey, minioSecretKey, ""),
		Secure: false,
	})
	failOnError(err, "failed to connect to MinIO")

	ctx := context.Background()
	bucketName := "snapshots"
	err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
	if err != nil {
		exists, errBucket := minioClient.BucketExists(ctx, bucketName)
		if errBucket != nil || !exists {
			log.Fatalf("bucket error: %s", err)
		}
	}

	// Queue arguments: if MITIGATED=true, set limits
	args := amqp.Table{}
	mitigated, _ := strconv.ParseBool(getEnv("MITIGATED", "false"))
	if mitigated {
		args = amqp.Table{
			"x-max-length-bytes": 500 * 1024 * 1024,
			"x-message-ttl":      300000,
		}
		log.Println("Mitigated mode: queue limits enabled")
	}

	if mitigated {
		_, err = ch.QueueDelete("snapshots", false, false, false)
		if err != nil {
			log.Printf("Note: could not delete queue (might not exist): %s", err)
		}
	}
	q, err := ch.QueueDeclare(
		"snapshots",
		true,  // durable
		false, // autoDelete
		false, // exclusive
		false, // noWait
		args,
	)
	failOnError(err, "failed to declare queue")

	if mitigated {
		err = ch.Qos(1, 0, false)
		failOnError(err, "failed to set QoS")
		log.Println("Mitigated mode: prefetch=1")
	}

	msgs, err := ch.Consume(
		q.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "failed to register consumer")

	log.Println("Consumer started, waiting for messages...")
	processed := 0
	startTime := time.Now()

	for msg := range msgs {
		objectName := fmt.Sprintf("snapshot-%d.bin", processed)
		reader := bytes.NewReader(msg.Body)

		_, err := minioClient.PutObject(ctx, bucketName, objectName, reader, int64(len(msg.Body)), minio.PutObjectOptions{})
		if err != nil {
			log.Printf("failed to save to MinIO: %s", err)
			msg.Nack(false, true)
			continue
		}

		msg.Ack(false)
		processed++

		if processed%100 == 0 {
			elapsed := time.Since(startTime)
			rate := float64(processed) / elapsed.Seconds()
			log.Printf("processed %d messages, rate: %.2f msg/s", processed, rate)
		}
	}
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
