package service

import (
	"context"
	"delayed-notifier/internal/notification/rabbitmq/notifier"
	"delayed-notifier/internal/notification/types/domain"
	"delayed-notifier/internal/notification/types/dto"
	"delayed-notifier/pkg/errutils"
	"github.com/google/uuid"
	"github.com/wb-go/wbf/retry"
	"time"
)

type Notifier interface {
	Publish(notification notifier.Message, strategy retry.Strategy) error
}

type Repo interface {
	CreateNotification(ctx context.Context, notification domain.Notification) error
}

type Cache interface {
	SetStatusWithRetry(ctx context.Context, id uuid.UUID, status string, strategy retry.Strategy) error
}

type Notification struct {
	repo     Repo
	notifier Notifier
	cache    Cache
}

func (s *Notification) Create(ctx context.Context, notification dto.Notification, strategy retry.Strategy) error {
	const op = "service.notification.Create"

	domainNotif, err := dtoToDomain(notification)
	if err != nil {
		return errutils.Wrap(op, err)
	}
	if err := s.repo.CreateNotification(ctx, domainNotif); err != nil {
		return errutils.Wrap(op, err)
	}

	err = s.cache.SetStatusWithRetry(ctx, domainNotif.ID, string(domainNotif.Status), strategy)
	if err != nil {
		return errutils.Wrap(op, err)
	}

	message := domainToMessage(domainNotif)
	if err = s.notifier.Publish(message, strategy); err != nil {
		return errutils.Wrap(op, err)
	}

	return nil
}

func domainToMessage(notification domain.Notification) notifier.Message {
	message := notifier.Message{
		ID:          notification.ID,
		Message:     notification.Message,
		ScheduledAt: notification.ScheduledAt,
		Channel:     string(notification.Channel),
		Recipient:   notification.Recipient,
	}

	return message
}

func dtoToDomain(dto dto.Notification) (domain.Notification, error) {
	var domainCh domain.NotificationChannel
	if dto.Channel == "telegram" {
		domainCh = domain.Telegram
	} else {
		domainCh = domain.Email
	}

	parsedTime, err := time.Parse(time.RFC3339, dto.ScheduledAt)
	if err != nil {
		location, locErr := time.LoadLocation("Europe/Moscow")
		if locErr != nil {
			return domain.Notification{}, locErr
		}
		parsedTime, err = time.ParseInLocation(time.DateTime, dto.ScheduledAt, location)
		if err != nil {
			return domain.Notification{}, err
		}
	}

	parsedTime = parsedTime.UTC()

	return domain.Notification{
		ID:          uuid.New(),
		Message:     dto.Message,
		ScheduledAt: parsedTime,
		Channel:     domainCh,
		Recipient:   dto.Recipient,
	}, nil
}
