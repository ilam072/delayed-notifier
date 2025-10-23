package rest

import (
	"context"
	"delayed-notifier/internal/notification/service"
	"delayed-notifier/internal/notification/types/dto"
	"delayed-notifier/internal/response"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/retry"
	"github.com/wb-go/wbf/zlog"
	"net/http"
)

type Notification interface {
	Create(ctx context.Context, notification dto.Notification, strategy retry.Strategy) error
	GetStatusByID(ctx context.Context, ID string) (string, error)
	SetStatus(ctx context.Context, ID string, status string, strategy retry.Strategy) error
}

type Validator interface {
	Validate(i interface{}) error
}

type Handler struct {
	notification Notification
	validator    Validator
	strategy     retry.Strategy
}

func New(notification Notification, strategy retry.Strategy) *Handler {
	return &Handler{
		notification: notification,
		strategy:     strategy,
	}
}

func (h *Handler) CreateNotification(c *ginext.Context) {
	var dtoNotif dto.Notification

	if err := json.NewDecoder(c.Request.Body).Decode(&dtoNotif); err != nil {
		zlog.Logger.Error().Err(err).Msg("failed to decode request body")
		c.JSON(http.StatusBadRequest, response.Error("invalid request body"))
		return
	}

	if err := h.validator.Validate(dtoNotif); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(fmt.Sprintf("validation error: %s", err.Error())))
		return
	}

	if err := h.notification.Create(c.Request.Context(), dtoNotif, h.strategy); err != nil {
		zlog.Logger.Error().Err(err).Msg("failed to create notification")
		c.JSON(http.StatusInternalServerError, response.Error("failed to create notification"))
	}

	c.Status(http.StatusCreated)
}

func (h *Handler) GetNotificationStatus(c *ginext.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("id must be UUID format"))
		return
	}

	status, err := h.notification.GetStatusByID(c.Request.Context(), id.String())
	if err != nil {
		if errors.Is(err, service.ErrNotifNotFound) {
			c.JSON(http.StatusNotFound, response.Error("notification with such id not found"))
			return
		}
		zlog.Logger.Error().Err(err).Str("id", id.String()).Msg("failed to cancel notification")
		c.JSON(http.StatusInternalServerError, response.Error("failed to get notification status"))
		return
	}

	c.JSON(http.StatusOK, response.Success(status))
}

func (h *Handler) CancelNotification(c *ginext.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("id must be UUID format"))
		return
	}

	if err := h.notification.SetStatus(c.Request.Context(), id.String(), "canceled", h.strategy); err != nil {
		if errors.Is(err, service.ErrNotifNotFound) {
			c.JSON(http.StatusNotFound, response.Error("notification with such id not found"))
			return
		}
		zlog.Logger.Error().Err(err).Str("id", id.String()).Msg("failed to cancel notification")
		c.JSON(http.StatusInternalServerError, response.Error("failed to cancel notification"))
		return
	}

	c.Status(http.StatusOK)
}
