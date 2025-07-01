package pubsub

import (
	"context"
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType int

const (
	durableQueue SimpleQueueType = iota
	transientQueue
)

func PublishJSON[T any](
	ch *amqp.Channel,
	exchange,
	key string,
	val T,
) error {

	rawJSON, err := json.Marshal(val)
	if err != nil {
		return err
	}

	err = ch.PublishWithContext(
		context.Background(),
		exchange,
		key,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        rawJSON,
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
	handler func(T),
) error {

	ch, q, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return err
	}

	delivery, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	go func() {
		for msg := range delivery {
			var body T
			if err := json.Unmarshal(msg.Body, &body); err != nil {
				log.Fatalf("error unmarshaling: %v", err)
			}

			handler(body)

			if err = msg.Ack(false); err != nil {
				log.Fatalf("error ack: %v", err)
			}
		}
	}()

	return nil
}

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
) (*amqp.Channel, amqp.Queue, error) {
	ch, err := conn.Channel()
	if err != nil {
		return &amqp.Channel{}, amqp.Queue{}, err
	}

	var (
		isDurable  bool
		autoDelete bool
		exclusive  bool
	)

	switch queueType {
	case durableQueue:
		isDurable = true
	case transientQueue:
		autoDelete = true
		exclusive = true
	}

	q, err := ch.QueueDeclare(queueName, isDurable, autoDelete, exclusive, false, nil)
	if err != nil {
		return &amqp.Channel{}, amqp.Queue{}, err
	}

	err = ch.QueueBind(queueName, key, exchange, false, nil)
	if err != nil {
		return &amqp.Channel{}, amqp.Queue{}, err
	}

	return ch, q, nil
}
