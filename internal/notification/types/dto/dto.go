package dto

type Notification struct {
	Message     string `json:"message" validate:"required"`
	ScheduledAt string `json:"scheduled_at" validate:"required"`
	Channel     string `json:"channel" validate:"required"`
	Recipient   string `json:"recipient" validate:"required"`
}

type SendNotification struct {
	Message     string
	ScheduledAt string
	Channel     string
	Recipient   string
}
