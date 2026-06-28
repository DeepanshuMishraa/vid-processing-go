package services

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func ConnectRabbitMQ(connectionUrl string) error {
	conn, err := amqp.Dial(connectionUrl)

	if err != nil {
		log.Println("Failed to connect to RabbitMQ")
		return err
	}

	log.Println("Connected to RabbitMQ")

	defer conn.Close()
	return nil
}
