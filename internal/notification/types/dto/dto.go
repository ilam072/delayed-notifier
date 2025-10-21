package dto

type Notification struct {
	Message     string
	ScheduledAt string
	Retries     int
	Channel     string
	Recipient   string
}
