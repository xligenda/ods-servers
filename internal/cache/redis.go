package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/xligenda/ods-servers/internal/structs"
)

type RedisCacheEntities interface {
	structs.Server
	GetID() string
}

type RedisCache[T RedisCacheEntities] struct {
	client     *redis.Client
	collection string
	ttl        time.Duration //  0 - без TTL
}

func NewRedisCache[T RedisCacheEntities](
	client *redis.Client,
	collection string,
	ttl time.Duration,
) *RedisCache[T] {
	return &RedisCache[T]{
		client:     client,
		collection: collection,
		ttl:        ttl,
	}
}

func (c *RedisCache[T]) Set(value T) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("%s:%s", c.collection, value.GetID())

	if c.ttl == 0 {
		return c.client.Set(context.Background(), key, data, 0).Err()
	}
	return c.client.SetEx(context.Background(), key, data, c.ttl).Err()
}

func (c *RedisCache[T]) GetAll() ([]T, error) {
	var keys []string
	var cursor uint64
	var err error

	for {
		keysBatch, nextCursor, err := c.client.Scan(context.Background(), cursor, fmt.Sprintf("%s:*", c.collection), 100).Result()
		if err != nil {
			return nil, err
		}

		keys = append(keys, keysBatch...)
		cursor = nextCursor

		if cursor == 0 {
			break
		}
	}

	if len(keys) == 0 {
		return nil, nil
	}

	values, err := c.client.MGet(context.Background(), keys...).Result()
	if err != nil {
		return nil, err
	}

	var result []T
	for _, val := range values {
		if val == nil {
			continue
		}

		var item T
		if err := json.Unmarshal([]byte(val.(string)), &item); err != nil {
			return nil, err
		}
		result = append(result, item)
	}

	return result, nil
}

func (c *RedisCache[T]) Get(key string) (*T, error) {
	data, err := c.client.Get(context.Background(), fmt.Sprintf("%s:%s", c.collection, key)).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var value T
	if err := json.Unmarshal([]byte(data), &value); err != nil {
		return nil, err
	}

	return &value, nil
}

func (c *RedisCache[T]) Delete(key string) error {
	return c.client.Del(context.Background(), fmt.Sprintf("%s:%s", c.collection, key)).Err()
}
