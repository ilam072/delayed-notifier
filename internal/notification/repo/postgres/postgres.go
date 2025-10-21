package postgres

import (
	"context"
	"database/sql"
	"delayed-notifier/internal/notification/repo"
	"delayed-notifier/internal/notification/types/domain"
	"delayed-notifier/pkg/errutils"
	"errors"
	"github.com/google/uuid"
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

func (r *Repo) GetStatusByID(ctx context.Context, ID uuid.UUID) (domain.NotificationStatus, error) {
	const op = "repo.notification.GetStatusByID"

	query := `SELECT status FROM notifications WHERE id = $1`

	var status domain.NotificationStatus
	if err := r.db.QueryRowContext(ctx, query, ID).Scan(&status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", errutils.Wrap(op, repo.ErrNotifNotFound)
		}
		return "", errutils.Wrap(op, err)
	}

	return status, nil
}

func (r *Repo) UpdateStatus(ctx context.Context, ID uuid.UUID, status domain.NotificationStatus) error {
	const op = "repo.notification.UpdateStatus"

	query := `UPDATE notifications SET status = $1 WHERE id = $2`

	res, err := r.db.ExecContext(ctx, query, status, ID)
	if err != nil {
		return errutils.Wrap(op, err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return errutils.Wrap(op, err)
	}

	if rows == 0 {
		return errutils.Wrap(op, repo.ErrNotifNotFound)
	}

	return nil
}
