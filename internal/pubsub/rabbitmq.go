package pubsub

import (
	"context"
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQClient struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

// Return the RabbitMQClient to do the operations
func New(connStr string) (*RabbitMQClient, error) {

	if connStr == "" {
		log.Fatalf("Unable to Dial, no string to dial to")
	}

	conn, err := amqp.Dial(connStr)
	//maybe have it return instead
	failOnError(err, "Failed to connect to RabbitMQ")

	ch, err := conn.Channel()
	// correct message or no?
	// have it return here too i guess
	failOnError(err, "Failed to get Channel")

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

func (r *RabbitMQClient) Publish2JSON[T any] (exchange, key string, val T, ctx context.Context) error {
	body, err := json.Marshal(val)
	if err != nil {
		return err
	}
	return r.Channel.PublishWithContext(
		ctx, //context
		exchange, //exchange
		key, //routing key
		false, //mandatory
		false, //immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body: body,
		}

	)
}
