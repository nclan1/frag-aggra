package pubsub

import (
	"context"
	"encoding/json"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQClient struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
}

func New(connStr string) (*RabbitMQClient, error) {
	if connStr == "" {
		log.Fatalf("Unable to Dial, no string to dial to")
	}

	var conn *amqp.Connection
	var err error
	for i := range 10 {
		conn, err = amqp.Dial(connStr)
		if err == nil {
			break
		}
		log.Printf("RabbitMQ not ready, retrying in 3s... (%d/10)", i+1)
		time.Sleep(3 * time.Second)
	}
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	return &RabbitMQClient{
		Conn:    conn,
		Channel: ch,
	}, nil
}

// Closes the connection, defer this func when creating a new RabbitMQClient
func (r *RabbitMQClient) Close() {
	if r != nil && r.Channel != nil && r.Conn != nil {
		r.Conn.Close()
		r.Channel.Close()
	}
}

func (r *RabbitMQClient) ConsumeFromClient(q string) (<-chan amqp.Delivery, error) {

	msgs, err := r.Channel.Consume(
		q,     // queue
		"",    // consumer tag (empty for auto-generation)
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // arguments
	)
	return msgs, err
}

// Calling publish on this for the scraper service
func (r *RabbitMQClient) Publish2JSON(exchange, key string, val any, ctx context.Context) error {
	body, err := json.Marshal(val)
	if err != nil {
		return err
	}
	return r.Channel.PublishWithContext(
		ctx,      //context
		exchange, //exchange
		key,      //routing key
		false,    //mandatory
		false,    //immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
		},
	)
}
