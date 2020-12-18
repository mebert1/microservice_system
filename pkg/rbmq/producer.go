package rbmq

import (
	"github.com/streadway/amqp"
)

// Producer wraps a rabbitmq producer
type Producer struct {
	channel  *amqp.Channel
	exchange string
}

// initProducer initializes and returns a new producer instance
func initProducer(conn *amqp.Connection, exchangeName string, exchangeType string) (*Producer, error) {
	channel, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	err = channel.ExchangeDeclare(
		exchangeName, // name
		exchangeType, // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		return nil, err
	}

	producer := &Producer{
		channel:  channel,
		exchange: exchangeName,
	}

	return producer, nil
}

// Publish sends a message to an exchange
func (p *Producer) Publish(msg []byte, routingKey string) error {
	err := p.channel.Publish(
		p.exchange, // publish to an exchange
		routingKey, // routing to 0 or more queues
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			Headers:         amqp.Table{},
			ContentType:     "application/json",
			ContentEncoding: "",
			Body:            msg,
			DeliveryMode:    amqp.Transient, // 1=non-persistent, 2=persistent
			Priority:        0,              // 0-9
		},
	)
	return err
}

// Close shuts down the producer
func (p *Producer) Close() error {
	return p.channel.Close()
}
