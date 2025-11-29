package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"HobitsService/internal/logger"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

// NotificationMessage представляет сообщение для отправки в Telegram бота
type NotificationMessage struct {
	UserID        int    `json:"user_id"`
	TelegramID    int64  `json:"telegram_id"`
	Message       string `json:"message"`
	UnconfirmedCount int `json:"unconfirmed_count"`
}

// RabbitMQClient клиент для работы с RabbitMQ
type RabbitMQClient struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   amqp.Queue
}

// NewRabbitMQClient создает новый RabbitMQ клиент
func NewRabbitMQClient(url string) (*RabbitMQClient, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Объявляем очередь для уведомлений
	q, err := ch.QueueDeclare(
		"habit_notifications", // name
		true,                  // durable
		false,                 // delete when unused
		false,                 // exclusive
		false,                 // no-wait
		nil,                   // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	logger.Info("RabbitMQ client created successfully", zap.String("queue", q.Name))

	return &RabbitMQClient{
		conn:    conn,
		channel: ch,
		queue:   q,
	}, nil
}

// PublishNotification отправляет уведомление в очередь
func (r *RabbitMQClient) PublishNotification(ctx context.Context, msg NotificationMessage) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	err = r.channel.PublishWithContext(
		ctx,
		"",           // exchange
		r.queue.Name, // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
			Timestamp:    time.Now(),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	logger.Debug("Notification published to RabbitMQ",
		zap.Int("user_id", msg.UserID),
		zap.Int("unconfirmed_count", msg.UnconfirmedCount),
	)

	return nil
}

// Close закрывает соединение с RabbitMQ
func (r *RabbitMQClient) Close() error {
	if r.channel != nil {
		if err := r.channel.Close(); err != nil {
			logger.Error("Failed to close channel", zap.Error(err))
		}
	}
	if r.conn != nil {
		if err := r.conn.Close(); err != nil {
			logger.Error("Failed to close connection", zap.Error(err))
			return err
		}
	}
	logger.Info("RabbitMQ client closed")
	return nil
}
