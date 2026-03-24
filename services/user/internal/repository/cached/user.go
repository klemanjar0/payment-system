package cached

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/klemanjar0/payment-system/pkg/hash"
	"github.com/klemanjar0/payment-system/pkg/logger"
	"github.com/klemanjar0/payment-system/services/user/internal/domain"
	"github.com/redis/go-redis/v9"
)

const (
	RedisKeyUserIdCache    = "user:id:%s"
	RedisKeyUserEmailCache = "user:email:%s" // References User ID -> %s has email string : ID
	RedisKeyUserPhoneCache = "user:phone:%s" // References User ID -> %s has phone string : ID
)

var (
	CacheExpirationTime = 5 * time.Minute
)

type CachedUserRepository struct {
	next   domain.UserRepository
	redis  *redis.Client
	hasher *hash.Hasher
}

func NewCachedUserRepository(next domain.UserRepository, redis *redis.Client) *CachedUserRepository {
	hasher := hash.New()
	return &CachedUserRepository{next: next, redis: redis, hasher: hasher}
}

func (r *CachedUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	cacheKey := fmt.Sprintf(RedisKeyUserIdCache, id)
	cachedData, err := r.redis.Get(ctx, cacheKey).Result()

	if err == nil && cachedData != "" {
		var user domain.User
		if err := json.Unmarshal([]byte(cachedData), &user); err == nil {
			logger.Debug("User fetched from cache", "id", id)

			return &user, nil
		}
	}

	user, err := r.next.GetByID(ctx, id)

	if err != nil {
		logger.Error("Failed to get user from the database", "id", id, "err", err)
		return nil, err
	}

	r.saveUserToRedis(ctx, user)

	return user, nil
}

func (r *CachedUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	emailHash := r.hasher.Hash(email)
	cacheKey := fmt.Sprintf(RedisKeyUserEmailCache, emailHash)
	cachedId, err := r.redis.Get(ctx, cacheKey).Result()

	if err != nil && !errors.Is(err, redis.Nil) {
		logger.Warn("Redis GET failed, falling back to DB", "key", cacheKey, "err", err)
	}

	if cachedId != "" && err == nil {
		logger.Debug("User Id fetched from cache by email", "email", email)
		return r.GetByID(ctx, cachedId)
	}

	user, err := r.next.GetByEmail(ctx, email)

	if err != nil {
		logger.Error("Failed to get user from the database", "email", email, "err", err)
		return nil, err
	}

	if err := r.redis.Set(ctx, cacheKey, user.ID, CacheExpirationTime).Err(); err != nil {
		logger.Warn("Failed to cache user data in redis", "user_id", user.ID, "err", err)
		// continue anyway - cache failure is not critical
	}

	r.saveUserToRedis(ctx, user)

	return user, nil
}

func (r *CachedUserRepository) GetByPhone(ctx context.Context, phone string) (*domain.User, error) {
	phoneHash := r.hasher.Hash(phone)
	cacheKey := fmt.Sprintf(RedisKeyUserPhoneCache, phoneHash)
	cachedId, err := r.redis.Get(ctx, cacheKey).Result()

	if err != nil && !errors.Is(err, redis.Nil) {
		logger.Warn("Redis GET failed, falling back to DB", "key", cacheKey, "err", err)
	}

	if cachedId != "" && err == nil {
		logger.Debug("User Id fetched from cache by phone", "phone", phone)
		return r.GetByID(ctx, cachedId)
	}

	user, err := r.next.GetByPhone(ctx, phone)

	if err != nil {
		logger.Error("Failed to get user from the database", "phone", phone, "err", err)
		return nil, err
	}

	if err := r.redis.Set(ctx, cacheKey, user.ID, CacheExpirationTime).Err(); err != nil {
		logger.Warn("Failed to cache user data in redis", "user_id", user.ID, "err", err)
		// continue anyway - cache failure is not critical
	}

	r.saveUserToRedis(ctx, user)

	return user, nil
}

// -- helpers --

func (r *CachedUserRepository) saveUserToRedis(ctx context.Context, user *domain.User) {
	if userData, err := json.Marshal(user); err == nil {
		idKey := fmt.Sprintf(RedisKeyUserIdCache, user.ID)
		if err := r.redis.Set(ctx, idKey, userData, CacheExpirationTime).Err(); err != nil {
			logger.Warn("Failed to cache user by id", "user_id", user.ID, "err", err)
			// continue anyway - cache failure is not critical
		}
	}
}
