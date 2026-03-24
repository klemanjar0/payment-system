package tokenblacklist

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const keyPrefix = "blacklist:"

// Blacklist manages a set of revoked JWT token identifiers.
type Blacklist interface {
	Blacklist(ctx context.Context, jti string, ttl time.Duration) error
	IsBlacklisted(ctx context.Context, jti string) (bool, error)
}

type redisBlacklist struct {
	client *redis.Client
}

func NewRedisBlacklist(client *redis.Client) Blacklist {
	return &redisBlacklist{client: client}
}

func (b *redisBlacklist) Blacklist(ctx context.Context, jti string, ttl time.Duration) error {
	return b.client.Set(ctx, key(jti), 1, ttl).Err()
}

func (b *redisBlacklist) IsBlacklisted(ctx context.Context, jti string) (bool, error) {
	n, err := b.client.Exists(ctx, key(jti)).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func key(jti string) string {
	return fmt.Sprintf("%s%s", keyPrefix, jti)
}
