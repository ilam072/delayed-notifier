package cache

import (
	"context"
	"delayed-notifier/pkg/errutils"
	"github.com/wb-go/wbf/redis"
	"github.com/wb-go/wbf/retry"
)

type Cache struct {
	client *redis.Client
}

func New(client *redis.Client) *Cache {
	return &Cache{client: client}
}

func (c *Cache) SetStatusWithRetry(ctx context.Context, id string, status string, strategy retry.Strategy) error {
	if err := c.client.SetWithRetry(ctx, strategy, id, status); err != nil {
		return errutils.Wrap("failed to cache status", err)
	}
	return nil
}

func (c *Cache) GetStatus(ctx context.Context, id string) (string, error) {
	status, err := c.client.Get(ctx, id)
	if err != nil {
		return "", errutils.Wrap("failed to get status from redis", err)
	}
	return status, nil
}
