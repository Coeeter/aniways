package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	redisRetryInterval = 2 * time.Second
)

type RedisClient struct {
	r        redis.Cmdable
	appEnv   string
	log      *slog.Logger
	useCache bool // If false, cache will be bypassed even if available
}

func NewRedisClient(
	ctx context.Context,
	appEnv, addr, pass string,
	log *slog.Logger,
	useCache bool,
) (*RedisClient, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     pass,
		DialTimeout:  3 * time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
	})

	isDev := appEnv == "development"

	for attempt := 1; ; attempt++ {
		pingCtx, pingCancel := context.WithTimeout(ctx, 3*time.Second)
		pingErr := rdb.Ping(pingCtx).Err()
		pingCancel()
		if pingErr == nil {
			break
		}
		if !isDev {
			return nil, fmt.Errorf("redis ping: %w", pingErr)
		}
		log.Warn("redis not ready, retrying...",
			"attempt", attempt,
			"error", pingErr,
			"retry_in", redisRetryInterval,
		)
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("redis ping cancelled: %w", ctx.Err())
		case <-time.After(redisRetryInterval):
		}
	}

	log.Info("connected to redis", "addr", addr, "useCache", useCache)

	return &RedisClient{r: rdb, appEnv: appEnv, log: log, useCache: useCache}, nil
}

func (c *RedisClient) Close() error {
	if client, ok := c.r.(*redis.Client); ok {
		return client.Close()
	}
	return fmt.Errorf("unsupported Redis client type")
}

func (c *RedisClient) Set(ctx context.Context, key string, value any, expiresIn time.Duration) error {
	valueBytes, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.r.Set(ctx, key, valueBytes, expiresIn).Err()
}

func (c *RedisClient) Get(ctx context.Context, key string, dest any) (bool, error) {
	raw, err := c.r.Get(ctx, key).Bytes()
	switch err {
	case nil:
	case redis.Nil:
		return false, nil // cache miss
	default:
		return false, err
	}

	return true, json.Unmarshal(raw, dest)
}

func (c *RedisClient) Del(ctx context.Context, key string) error {
	return c.r.Del(ctx, key).Err()
}

func (c *RedisClient) Pipeline() *RedisClient {
	return &RedisClient{r: c.r.Pipeline(), appEnv: c.appEnv, log: c.log, useCache: c.useCache}
}

func (c *RedisClient) Exec(ctx context.Context) ([]redis.Cmder, error) {
	if pipe, ok := c.r.(redis.Pipeliner); ok {
		return pipe.Exec(ctx)
	}
	return nil, fmt.Errorf("unsupported Redis client type for pipeline execution")
}

func GetOrFill[T any](
	ctx context.Context,
	rc *RedisClient,
	key string,
	ttl time.Duration,
	fetch func(context.Context) (T, error),
) (val T, err error) {
	var zero T

	shouldUseCache := rc.appEnv != "development" || rc.useCache

	var tmp T
	if shouldUseCache {
		if ok, err := rc.Get(ctx, key, &tmp); err == nil && ok {
			rc.log.Debug("cache hit", "key", key)
			return tmp, nil
		} else if err != nil {
			rc.log.Warn("cache get failed, fetching", "key", key, "err", err)
		}
	} else {
		rc.log.Debug("cache bypassed", "key", key, "useCache", rc.useCache, "appEnv", rc.appEnv)
	}

	tmp, err = fetch(ctx)
	if err != nil {
		return zero, err
	}

	// Only set cache if we're using it
	if shouldUseCache {
		if err = rc.Set(ctx, key, tmp, ttl); err != nil {
			rc.log.Warn("cache set failed", "key", key, "err", err)
		}
	}

	return tmp, nil
}
