package notifier

import (
	"delayed-notifier/pkg/errutils"
	"encoding/json"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/wb-go/wbf/rabbitmq"
	"github.com/wb-go/wbf/retry"
	"time"
)

type Opts struct {
	Exchange   string
	RoutingKey string
	Queue      string
	DLQ        string
}

type Message struct {
	ID          uuid.UUID
	Message     string
	ScheduledAt time.Time
	Channel     string
	Recipient   string
}

type Notifier struct {
	publisher *rabbitmq.Publisher
	consumer  *rabbitmq.Consumer
	opts      Opts
}

func New(channel *rabbitmq.Channel, opts Opts) (*Notifier, error) {
	table := amqp.Table{
		"x-delayed-type": "direct",
	}

	if err := channel.ExchangeDeclare(
		opts.Exchange, // "notification-exchange"
		"x-delayed-message",
		true,
		false,
		false,
		false,
		table,
	); err != nil {
		return nil, errutils.Wrap("failed to declare an exchange", err)
	}

	if _, err := channel.QueueDeclare(
		opts.DLQ, // "notification-dlq"
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return nil, errutils.Wrap("failed to declare DLQ", err)
	}

	table = amqp.Table{
		"x-dead-letter-routing-key": opts.DLQ, // "notification-dlq"
	}

	queue, err := channel.QueueDeclare(
		opts.Queue, // "notification-queue"
		true,
		false,
		false,
		false,
		table,
	)
	if err != nil {
		return nil, errutils.Wrap("failed to declare main notification queue", err)
	}

	if err := channel.QueueBind(
		opts.Queue,
		opts.RoutingKey, // "notification"
		opts.Exchange,
		false,
		nil,
	); err != nil {
		return nil, errutils.Wrap("failed to bind queue", err)
	}

	publisher := rabbitmq.NewPublisher(channel, opts.Exchange)
	consumer := rabbitmq.NewConsumer(channel, &rabbitmq.ConsumerConfig{
		Queue: queue.Name,
	})

	return &Notifier{
		publisher: publisher,
		consumer:  consumer,
		opts:      opts,
	}, nil
}

func (n *Notifier) Publish(notification Message, strategy retry.Strategy) error {
	delay := time.Until(notification.ScheduledAt)
	if delay < 0 {
		delay = 0
	}

	body, err := json.Marshal(notification)
	if err != nil {
		return errutils.Wrap("failed to marshal notification message", err)
	}

	headers := amqp.Table{"x-delay": int(delay.Milliseconds())}

	if err := n.publisher.PublishWithRetry(body, n.opts.RoutingKey, "application/json", strategy, rabbitmq.PublishingOptions{Headers: headers}); err != nil {
		return errutils.Wrap("failed to publish message with retry", err)
	}

	return nil
}
