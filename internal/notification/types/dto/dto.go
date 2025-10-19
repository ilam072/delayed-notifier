package dto

import (
	"time"
)

type Notification struct {
	Message     string
	ScheduledAt time.Time
	Retries     int
	Channel     string
	Recipient   string
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
