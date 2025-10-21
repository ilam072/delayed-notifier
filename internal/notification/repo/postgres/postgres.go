package postgres

import (
	"context"
	"delayed-notifier/internal/notification/types/domain"
	"delayed-notifier/pkg/errutils"
	"github.com/wb-go/wbf/dbpg"
)

type Repo struct {
	db *dbpg.DB
}

func New(db *dbpg.DB) *Repo {
	return &Repo{db: db}
}

func (r *Repo) CreateNotification(ctx context.Context, notification domain.Notification) error {
	const op = "repo.notification.Create"

	query := `
    INSERT INTO notification(id, message, scheduled_at, channel, recipient)
    VALUES ($1, $2, $3, $4, $5)`

	if _, err := r.db.ExecContext(
		ctx,
		query,
		notification.ID,
		notification.Message,
		notification.ScheduledAt,
		notification.Channel,
		notification.Recipient,
	); err != nil {
		return errutils.Wrap(op, err)
	}

	return nil
}
