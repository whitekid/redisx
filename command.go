package redisx

import (
	"context"
	"iter"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/whitekid/goxp/errors"
)

//
// Helpers
//

func (c *Client) Get(ctx context.Context, key string) (string, error) {
	v, err := c.Client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return v, ErrNotFound
		}

		return v, errors.Wrapf(err, "get failed: key=%s", key)
	}

	return v, err
}

func (c *Client) Set(ctx context.Context, key string, value any, expiration ...time.Duration) (string, error) {
	return Set(ctx, c.Client, key, value, expiration...)
}

func Set(ctx context.Context, c redis.Cmdable, key string, value any, expiration ...time.Duration) (string, error) {
	var expire time.Duration
	for _, exp := range expiration {
		expire = exp
		break
	}

	logger.Debugf("set: key=%s, value=%v, expire=%v", key, value, expire)
	v, err := c.Set(ctx, key, value, expire).Result()
	if err != nil {
		return v, errors.Wrapf(err, "set failed: key=%s, value=%v", key, value)
	}

	return v, err
}

func (c *Client) Del(ctx context.Context, key string) (int64, error) {
	v, err := c.Client.Del(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return v, ErrNotFound
		}

		return v, errors.Wrapf(err, "del failed: key=%s", key)
	}

	return v, err
}

func (c *Client) Ping(ctx context.Context) (string, error) {
	v, err := c.Client.Ping(ctx).Result()
	if err != nil {
		return v, errors.Wrapf(err, "ping failed: error=%+v", err)
	}

	return v, nil
}

func (c *Client) HGet(ctx context.Context, key string, field string) (string, error) {
	v, err := c.Client.HGet(ctx, key, field).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return v, ErrNotFound
		}

		return v, errors.Wrapf(err, "hget failed: key=%s, field=%s", key, field)
	}
	return v, nil
}

func (c *Client) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	v, err := c.Client.HGetAll(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return v, ErrNotFound
		}

		return nil, errors.Wrapf(err, "hgetall failed: key=%s", key)
	}
	return v, nil
}

func (c *Client) HSet(ctx context.Context, key string, values ...any) (int64, error) {
	v, err := c.Client.HSet(ctx, key, values...).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return v, ErrNotFound
		}

		return v, errors.Wrapf(err, "hset failed: key=%s, values=%v", key, values)
	}
	return v, nil
}

func (c *Client) HDel(ctx context.Context, key string, fields ...string) (int64, error) {
	return HDel(ctx, c.Client, key, fields...)
}

func HDel(ctx context.Context, c redis.Cmdable, key string, fields ...string) (int64, error) {
	v, err := c.HDel(ctx, key, fields...).Result()
	if err != nil {
		return v, errors.Wrapf(err, "hdel failed: key=%s, fields=%v", key, fields)
	}
	return v, nil
}

func (c *Client) RPush(ctx context.Context, key string, values ...any) (int64, error) {
	return RPush(ctx, c.Client, key, values...)
}

func RPush(ctx context.Context, c redis.Cmdable, key string, values ...any) (int64, error) {
	v, err := c.RPush(ctx, key, values...).Result()
	if err != nil {
		return v, errors.Wrapf(err, "rpush failed: key=%s, values=%v", key, values)
	}

	return v, nil
}

// LPop lpop from list
// return ErrNoValue if key not found or list is empty
func (c *Client) LPop(ctx context.Context, key string) (string, error) {
	return LPop(ctx, c.Client, key)
}

func LPop(ctx context.Context, c redis.Cmdable, key string) (string, error) {
	v, err := c.LPop(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", ErrNoValue
		}

		return v, errors.Wrapf(err, "lpop failed: key=%s", key)
	}
	return v, nil
}

func (c *Client) BPop(ctx context.Context, timeout time.Duration, keys ...string) ([]string, error) {
	return BLPop(ctx, c.Client, timeout, keys...)
}

func BLPop(ctx context.Context, c redis.Cmdable, timeout time.Duration, keys ...string) ([]string, error) {
	v, err := c.BLPop(ctx, timeout, keys...).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrNoValue
		}

		return v, errors.Wrapf(err, "lpop failed: key=%v", keys)
	}
	return v, nil
}

// Scan return keys, cursor, error
func (c *Client) Scan(ctx context.Context, cursor uint64, match string, count int64) ([]string, uint64, error) {
	return Scan(ctx, c.Client, cursor, match, count)
}

func Scan(ctx context.Context, c redis.Cmdable, cursor uint64, match string, count int64) ([]string, uint64, error) {
	keys, cursor, err := c.Scan(ctx, cursor, match, count).Result()
	if err != nil {
		return keys, cursor, errors.Wrapf(err, "scan failed: match=%s, cursor=%d, error=%+v", match, cursor, err)
	}

	return keys, cursor, nil
}

func (c *Client) Publilsh(ctx context.Context, channel string, message any) (int64, error) {
	logger.Debugf("publish(): channels=%+v, message=%+v", channel, message)

	r, err := c.Client.Publish(ctx, channel, message).Result()
	if err != nil {
		return r, errors.Wrapf(err, "publlish failed: channel=%s, message=%v", channel, message)
	}

	return r, nil
}

// PSubscribe return iterator for pubsub channel, payload
func (c *Client) PSubscribe(ctx context.Context, channels ...string) iter.Seq2[string, string] {
	return func(yield func(string, string) bool) {
		logger.Debugf("psubscribe(): channels=%+v", channels)

		pubsub := c.Client.PSubscribe(ctx, channels...)
		defer pubsub.Close()

		for payload := range pubsub.Channel() {
			if !yield(payload.Channel, payload.Payload) {
				return
			}
		}
	}
}

func (c *Client) TTL(ctx context.Context, key string) (time.Duration, error) {
	return TTL(ctx, c.Client, key)
}

func TTL(ctx context.Context, c redis.Cmdable, key string) (time.Duration, error) {
	v, err := c.TTL(ctx, key).Result()
	if err != nil {
		return v, errors.Wrapf(err, "ttl failed: key=%s", key)
	}

	if v == -2 {
		return v, ErrNotFound
	}

	return v, nil
}

func (c *Client) Expire(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	return Expire(ctx, c.Client, key, expiration)
}

func Expire(ctx context.Context, c redis.Cmdable, key string, expiration time.Duration) (bool, error) {
	v, err := c.Expire(ctx, key, expiration).Result()
	if err != nil {
		return v, errors.Wrapf(err, "expire failed: key=%s, expiration=%v", key, expiration)
	}

	return v, nil
}
