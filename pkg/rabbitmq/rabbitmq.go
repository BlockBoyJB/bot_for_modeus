package rabbitmq

import (
	"context"
	"encoding/json"
	amqp "github.com/rabbitmq/amqp091-go"
)

type broker interface {
	PublishMessage(ctx context.Context, queue string, message interface{}) error
	InitQueue(queue string) (<-chan amqp.Delivery, error)
	Close()
}

type RabbitMQ struct {
	conn *amqp.Connection
	broker
}

func New(url string) (*RabbitMQ, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}
	return &RabbitMQ{conn: conn}, nil
}

func (r *RabbitMQ) PublishMessage(ctx context.Context, queue string, message interface{}) error {
	m, err := json.Marshal(message)
	if err != nil {
		return err
	}
	ch, err := r.conn.Channel()
	if err != nil {
		return err
	}
	defer func() { _ = ch.Close() }()
	err = ch.PublishWithContext(
		ctx,
		"",
		queue,
		false,
		false,
		amqp.Publishing{Body: m},
	)
	return err
}

func (r *RabbitMQ) InitQueue(queue string) (<-chan amqp.Delivery, error) {
	ch, err := r.conn.Channel()
	if err != nil {
		return nil, err
	}
	q, err := ch.QueueDeclare(
		queue,
		false,
		false,
		true,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}
	m, err := ch.Consume(
		q.Name,
		"Consumer_"+q.Name,
		false,
		true,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (r *RabbitMQ) Close() {
	_ = r.conn.Close()
}
