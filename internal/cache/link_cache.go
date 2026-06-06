package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/winchester/shorturls/internal/models"
)

type LinkCache struct {
	client *redis.Client
}

func NewLinkCache(client *redis.Client) *LinkCache {
	return &LinkCache{client: client}
}

func (c *LinkCache) Set(ctx context.Context, link *models.Link) error {
	data, err := json.Marshal(link)
	if err != nil {
		return fmt.Errorf("marshal link: %w", err)
	}
	ttl := 10 * time.Minute
	return c.client.Set(ctx, linkKey(link.Alias), data, ttl).Err()
}

func (c *LinkCache) Get(ctx context.Context, alias string) (*models.Link, error) {
	data, err := c.client.Get(ctx, linkKey(alias)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("get cache: %w", err)
	}
	var link models.Link
	if err := json.Unmarshal(data, &link); err != nil {
		return nil, fmt.Errorf("unmarshal link: %w", err)
	}
	return &link, nil
}

func (c *LinkCache) Delete(ctx context.Context, alias string) error {
	return c.client.Del(ctx, linkKey(alias)).Err()
}

func linkKey(alias string) string {
	return fmt.Sprintf("link:%s", alias)
}
