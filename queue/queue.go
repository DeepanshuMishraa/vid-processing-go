package queue

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/DeepanshuMishraa/vid-processing-go.git/types"
	amqp "github.com/rabbitmq/amqp091-go"
)

func ConnectRabbitMQ(connectionUrl string) (*amqp.Connection, error) {
	conn, err := amqp.Dial(connectionUrl)
	if err != nil {
		log.Println("Failed to connect to RabbitMQ")
		return nil, err
	}

	log.Println("Connected to RabbitMQ")
	return conn, nil
}

func Publish(video_id string, conn *amqp.Connection) error {
	ch, err := conn.Channel()
	if err != nil {
		log.Println("Failed to open a channel")
		return err
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"video_queue",
		true,
		false,
		false,
		false,
		amqp.Table{
			amqp.QueueTypeArg: amqp.QueueTypeQuorum,
		},
	)
	if err != nil {
		log.Println("Failed to declare a queue")
		return err
	}

	job := types.VideoJob{
		VideoID: video_id,
	}

	body, err := json.Marshal(job)
	if err != nil {
		log.Println("Failed to marshal the job")
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	err = ch.PublishWithContext(ctx,
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        body,
		})
	if err != nil {
		log.Println("Failed to publish a message")
		return err
	}

	log.Println("Published a message")
	return nil
}

func Consume(conn *amqp.Connection) (<-chan amqp.Delivery, *amqp.Channel, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, nil, err
	}

	q, err := ch.QueueDeclare(
		"video_queue",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		ch.Close()
		return nil, nil, err
	}

	msgs, err := ch.Consume(
		q.Name,
		"",
		false, // manual ack
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		ch.Close()
		return nil, nil, err
	}

	return msgs, ch, nil
}
