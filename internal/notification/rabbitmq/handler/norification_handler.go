package handler

import (
	"context"
	"delayed-notifier/internal/notification/rabbitmq/notifier"
	"delayed-notifier/internal/notification/service"
	"delayed-notifier/internal/notification/types/dto"
	"errors"
	"github.com/wb-go/wbf/retry"
	"github.com/wb-go/wbf/zlog"
	"time"
)

type Notification interface {
	Send(notification dto.SendNotification) error
	SetStatus(ctx context.Context, ID string, status string, strategy retry.Strategy) error
}

type Handler struct {
	notification Notification
}

func (h *Handler) HandleNotif(ctx context.Context, notification notifier.Message, strategy retry.Strategy) {
	id := notification.ID.String()

	dtoNotif := dto.SendNotification{
		Message:     notification.Message,
		ScheduledAt: notification.ScheduledAt.Format(time.RFC3339),
		Channel:     notification.Channel,
		Recipient:   notification.Recipient,
	}

	handleFunc := func() error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			return h.notification.Send(dtoNotif)
		}
	}

	sendErr := retry.Do(handleFunc, strategy)
	status := "sent"
	if sendErr != nil {
		status = "failed"
	}

	if err := h.notification.SetStatus(ctx, id, status, strategy); err != nil {
		if errors.Is(err, service.ErrNotifNotFound) {
			zlog.Logger.Warn().Err(err).Str("id", id).Msg("notification not found")
			return
		}
		zlog.Logger.Error().Err(err).Str("id", id).Msgf("failed to set notification status (%s)", status)
		return
	}

	if sendErr != nil {
		zlog.Logger.Error().Err(sendErr).Str("id", id).Msg("failed to send notification")
		return
	}

	zlog.Logger.Info().Str("id", id).Msg("notification successfully sent")
}
