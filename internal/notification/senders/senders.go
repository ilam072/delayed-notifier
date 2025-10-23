package senders

import (
	"delayed-notifier/internal/notification/types/domain"
)

type NotificationSender interface {
	Send(message string, recipient string) error
}

type NotificationSenders struct {
	senders map[domain.NotificationChannel]NotificationSender
}

func New(emailClient NotificationSender) NotificationSenders {
	senders := make(map[domain.NotificationChannel]NotificationSender)

	senders[domain.Email] = emailClient
	//senders[domain.Telegram] = tg.New()

	return NotificationSenders{senders: senders}
}

func (s *NotificationSenders) ForChannel(channel domain.NotificationChannel) NotificationSender {
	return s.senders[channel]
}
