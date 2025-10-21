package rest

import (
	"delayed-notifier/internal/notification/service"
	"delayed-notifier/internal/notification/types/dto"
	"delayed-notifier/internal/response"
	"encoding/json"
	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/retry"
	"github.com/wb-go/wbf/zlog"
	"net/http"
)

type Handler struct {
	notification service.Notification
	strategy     retry.Strategy
}

func New(notification service.Notification, strategy retry.Strategy) *Handler {
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

	if err := h.notification.Create(c.Request.Context(), dtoNotif, h.strategy); err != nil {
		zlog.Logger.Error().Err(err).Msg("failed to create notification")
		c.JSON(http.StatusInternalServerError, response.Error("failed to create notification"))
	}

	c.Status(http.StatusCreated)
}
