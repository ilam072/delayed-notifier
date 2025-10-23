package dto

type Notification struct {
	Message     string
	ScheduledAt string
	Channel     string
	Recipient   string
}

type SendNotification struct {
	Message     string
	ScheduledAt string
	Channel     string
	Recipient   string
}
