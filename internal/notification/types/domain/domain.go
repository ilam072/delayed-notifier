package domain

import (
	"github.com/google/uuid"
	"time"
)

// NotificationChannel - enum для каналов уведомлений
type NotificationChannel string

const (
	Email    NotificationChannel = "email"
	Telegram NotificationChannel = "telegram"
)

// NotificationStatus - enum для статусов уведомлений
type NotificationStatus string

const (
	Scheduled NotificationStatus = "scheduled"
	Sent      NotificationStatus = "sent"
	Canceled  NotificationStatus = "canceled"
	Failed    NotificationStatus = "failed"
)

// Notification - структура уведомления
type Notification struct {
	ID          uuid.UUID
	Message     string
	ScheduledAt time.Time
	Retries     int
	Channel     NotificationChannel
	Recipient   string
	Status      NotificationStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
