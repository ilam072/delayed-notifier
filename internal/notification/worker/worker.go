package worker

import (
	"context"
	"delayed-notifier/internal/notification/rabbitmq/notifier"
	"github.com/wb-go/wbf/retry"
	"github.com/wb-go/wbf/zlog"
	"sync"
)

type NotifConsumer interface {
	Consume(ctx context.Context, notifications chan notifier.Message, strategy retry.Strategy) error
}

type NotifHandler interface {
	HandleNotif(ctx context.Context, notification notifier.Message, strategy retry.Strategy)
}

type Notification interface {
	GetStatusByID(ctx context.Context, id string) (string, error)
}

type WorkerPool struct {
	consumer NotifConsumer
	handler  NotifHandler
	notif    Notification
	workers  int
}

func NewWorkerPool(consumer NotifConsumer, handler NotifHandler, notif Notification, workers int) *WorkerPool {
	return &WorkerPool{
		consumer: consumer,
		handler:  handler,
		notif:    notif,
		workers:  workers,
	}
}

func (w *WorkerPool) Start(ctx context.Context, strategy retry.Strategy) {
	notifications := make(chan notifier.Message, w.workers*15)

	go func() {
		if err := w.consumer.Consume(ctx, notifications, strategy); err != nil {
			zlog.Logger.Error().Err(err).Msg("failed to consume notifications")
		}
	}()

	wg := sync.WaitGroup{}
	wg.Add(w.workers)
	for i := 0; i < w.workers; i++ {
		go func() {
			defer wg.Done()

			for {
				select {
				case <-ctx.Done():
					zlog.Logger.Info().Msg("worker shutting down due to canceled context")
					return
				case notification, ok := <-notifications:
					if !ok {
						zlog.Logger.Info().Msg("worker shutting down due to closed channel")
						return
					}

					status, err := w.notif.GetStatusByID(ctx, notification.ID.String())
					if err != nil {
						zlog.Logger.Error().Err(err).Str("id", notification.ID.String()).Msg("failed to get notification status by id")
						continue
					}

					if status != "canceled" {
						w.handler.HandleNotif(ctx, notification, strategy)
					}
				}

			}
		}()
	}
	<-ctx.Done()
	wg.Wait()
}
