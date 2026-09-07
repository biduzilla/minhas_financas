package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisCache(addr, password string, db int, ttl *time.Duration) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	var defaultTTL time.Duration
	if ttl != nil {
		defaultTTL = *ttl
	}

	return &RedisCache{client: client, ttl: defaultTTL}, nil
}

func (r *RedisCache) Get(ctx context.Context, key string, dest any) error {
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

func (r *RedisCache) Set(ctx context.Context, key string, value any, ttl *time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	t := ttl

	if t == nil {
		t = &r.ttl
	}
	return r.client.Set(ctx, key, data, *t).Err()
}

func (r *RedisCache) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	return r.client.Del(ctx, keys...).Err()
}

func (r *RedisCache) DeleteByPrefix(ctx context.Context, prefix string) error {
	var cursor uint64
	pattern := prefix + "*"

	for {
		keys, nextCursor, err := r.client.Scan(
			ctx,
			cursor,
			pattern,
			100,
		).Result()

		if err != nil {
			return err
		}

		if len(keys) > 0 {
			if err := r.client.Del(ctx, keys...).Err(); err != nil {
				return err
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
	return nil
}
