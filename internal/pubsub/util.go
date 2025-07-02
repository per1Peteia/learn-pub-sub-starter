package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/per1Peteia/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Acktype int

const (
	Ack Acktype = iota
	NackRequeue
	NackDiscard
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
	handler func(T) Acktype,
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

			msgHandlingStatus := handler(body)

			switch msgHandlingStatus {
			case Ack:
				if err := msg.Ack(false); err != nil {
					log.Fatalf("error ack: %v", err)
				}
			case NackRequeue:
				if err := msg.Nack(false, true); err != nil {
					log.Fatalf("error nack: %v", err)
				}
			case NackDiscard:
				if err := msg.Nack(false, false); err != nil {
					log.Fatalf("error nack: %v", err)
				}
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

	q, err := ch.QueueDeclare(queueName, isDurable, autoDelete, exclusive, false, amqp.Table{"x-dead-letter-exchange": routing.ExchangeDLX})
	if err != nil {
		return &amqp.Channel{}, amqp.Queue{}, err
	}

	err = ch.QueueBind(queueName, key, exchange, false, nil)
	if err != nil {
		return &amqp.Channel{}, amqp.Queue{}, err
	}

	return ch, q, nil
}
