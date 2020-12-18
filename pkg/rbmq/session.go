package rbmq

import (
	"fmt"
	"time"

	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

// Session wraps a amqp connection and its config
type Session struct {
	conn   *amqp.Connection
	config Config
}

// Config contains the data required to establish a connection to a rabbitmq server
type Config struct {
	User         string
	Password     string
	URL          string
	ExchangeName string
	ExchangeType string
	BindingKey   string
	ConsumerTag  string
}

// NewSession connects to the rabbitmq server and returns an active session
func NewSession(config Config, logger *zap.SugaredLogger) (*Session, error) {
	amqpURL := fmt.Sprintf("amqp://%s:%s@%s", config.User, config.Password, config.URL)

	for {
		conn, err := amqp.Dial(amqpURL)
		if err != nil {
			logger.Infow("Failed to connect to RabbitMQ, retrying...")
			time.Sleep(5 * time.Second)
			continue
		}

		session := &Session{
			conn:   conn,
			config: config,
		}
		return session, nil
	}
}

// NewConsumer uses the sessions connection to connect a new consumer to the rabbitmq server
func (s *Session) NewConsumer(messages chan<- Message) (*Consumer, error) {
	return initConsumer(s.conn, s.config, messages)
}

// NewProducer uses the sessions connection to connect a new producer to the rabbitmq server
func (s *Session) NewProducer(exchangeName string, exchangeType string) (*Producer, error) {
	return initProducer(s.conn, exchangeName, exchangeType)
}

// Close shuts down the session and closes the connection
func (s *Session) Close() error {
	return s.conn.Close()
}
