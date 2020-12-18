package rbmq

import (
	"fmt"

	"github.com/streadway/amqp"
)

// Consumer wraps a rabbitmq consumer
type Consumer struct {
	channel     *amqp.Channel
	consumerTag string
	done        chan error
}

// Message is a wrapper for a rabbitmq message
type Message struct {
	RoutingKey string
	Body       []byte
}

// initConsumer initializes and returns a new rabbitmq consumer instance
func initConsumer(conn *amqp.Connection, config Config, messages chan<- Message) (*Consumer, error) {
	done := make(chan error)

	channel, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	err = channel.ExchangeDeclare(
		config.ExchangeName, // name
		config.ExchangeType, // type
		true,                // durable
		false,               // auto-deleted
		false,               // internal
		false,               // no-wait
		nil,                 // arguments
	)
	if err != nil {
		return nil, err
	}

	queue, err := channel.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return nil, err
	}

	err = channel.QueueBind(
		queue.Name,          // queue name
		config.BindingKey,   // routing key
		config.ExchangeName, // exchange
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	deliveries, err := channel.Consume(
		queue.Name,         // queue
		config.ConsumerTag, // consumer
		true,               // auto ack
		false,              // exclusive
		false,              // no local
		false,              // no wait
		nil,                // args
	)
	if err != nil {
		return nil, err
	}

	go handle(deliveries, messages, done)

	consumer := &Consumer{
		channel:     channel,
		consumerTag: config.ConsumerTag,
		done:        done,
	}

	return consumer, nil
}

// handle is the function that forwards incoming messages to the message channel declared in /cmd/service/main.go
func handle(deliveries <-chan amqp.Delivery, messages chan<- Message, done chan error) {
	for delivery := range deliveries {
		msg := Message{
			RoutingKey: delivery.RoutingKey,
			Body:       delivery.Body,
		}
		messages <- msg
	}
	done <- nil
}

// Close is used to shut down a consumer instance
func (c *Consumer) Close() error {
	if err := c.channel.Cancel(c.consumerTag, true); err != nil {
		return fmt.Errorf("consumer cancel failed: %s", err)
	}

	return <-c.done
}
