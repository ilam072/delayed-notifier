package service

import (
	"context"
	"delayed-notifier/internal/notification/rabbitmq/notifier"
	"delayed-notifier/internal/notification/repo"
	"delayed-notifier/internal/notification/types/domain"
	"delayed-notifier/internal/notification/types/dto"
	"delayed-notifier/pkg/errutils"
	"errors"
	"github.com/google/uuid"
	"github.com/wb-go/wbf/redis"
	"github.com/wb-go/wbf/retry"
	"github.com/wb-go/wbf/zlog"
	"time"
)

type Notifier interface {
	Publish(notification notifier.Message, strategy retry.Strategy) error
}

type Repo interface {
	CreateNotification(ctx context.Context, notification domain.Notification) error
	GetStatusByID(ctx context.Context, ID uuid.UUID) (domain.NotificationStatus, error)
}

type Cache interface {
	SetStatusWithRetry(ctx context.Context, id string, status string, strategy retry.Strategy) error
	GetStatus(ctx context.Context, id string) (string, error)
}

type Notification struct {
	notifRepo Repo
	notifier  Notifier
	cache     Cache
}

var (
	ErrNotifNotFound = errors.New("notification not found")
)

func (s *Notification) Create(ctx context.Context, notification dto.Notification, strategy retry.Strategy) error {
	const op = "service.notification.Create"

	domainNotif, err := dtoToDomain(notification)
	if err != nil {
		return errutils.Wrap(op, err)
	}
	if err := s.notifRepo.CreateNotification(ctx, domainNotif); err != nil {
		return errutils.Wrap(op, err)
	}

	err = s.cache.SetStatusWithRetry(ctx, domainNotif.ID.String(), string(domainNotif.Status), strategy)
	if err != nil {
		return errutils.Wrap(op, err)
	}

	message := domainToMessage(domainNotif)
	if err = s.notifier.Publish(message, strategy); err != nil {
		return errutils.Wrap(op, err)
	}

	return nil
}

func (n *Notification) GetStatusByID(ctx context.Context, ID string) (string, error) {
	const op = "service.notification.GetStatusByID"

	status, err := n.cache.GetStatus(ctx, ID)
	if err == nil {
		return status, nil
	}
	if !errors.Is(err, redis.NoMatches) {
		zlog.Logger.Error().Err(err).Str("id", ID).Msg("failed to get notification status from cache")
	}

	parsedID, err := uuid.Parse(ID)
	if err != nil {
		return "", errutils.Wrap(op, err)
	}

	domainStatus, err := n.notifRepo.GetStatusByID(ctx, parsedID)
	if err != nil {
		if errors.Is(err, repo.ErrNotifNotFound) {
			return "", errutils.Wrap(op, ErrNotifNotFound)
		}
		return "", errutils.Wrap(op, err)
	}

	return string(domainStatus), nil
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
