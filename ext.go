package redisx

import (
	"context"
	"iter"
	"slices"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/whitekid/goxp/errors"
)

type ScanOpts struct {
	Match string
}

// Scanned return scan iterator
func (c *Client) ScanValues(ctx context.Context, opts ScanOpts) iter.Seq2[string, error] {
	return ScanValues(ctx, c.Client, opts)
}

// ScanValues return scan iterator
func ScanValues(ctx context.Context, c redis.Cmdable, opts ScanOpts) iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {
		var cursor uint64

		for {
			keys, cursor, err := Scan(ctx, c, cursor, opts.Match, 1000)
			if err != nil {
				yield("", errors.Wrapf(err, "scan failed"))
				return
			}

			for _, key := range keys {
				if !yield(key, nil) {
					return
				}
			}

			if cursor == 0 {
				break
			}
		}
	}
}

type NameAndValue[T1 any, T2 any] struct {
	Name  T1
	Value T2
}

func (c *Client) BLPopValues(ctx context.Context, timeout time.Duration, keys ...string) iter.Seq2[*NameAndValue[string, string], error] {
	return BLPopValues(ctx, c.Client, timeout, keys...)
}

// BLPopValues return BLPop iterator,
func BLPopValues(ctx context.Context, c redis.Cmdable, timeout time.Duration, keys ...string) iter.Seq2[*NameAndValue[string, string], error] {
	return func(yield func(*NameAndValue[string, string], error) bool) {
		for {
			items, err := BLPop(ctx, c, timeout, keys...)
			if err != nil {
				if errors.Is(err, context.DeadlineExceeded) {
					yield(nil, err)
					return
				}

				if errors.Is(err, redis.Nil) {
					logger.Debug("blpop timeout, retry..")
					continue
				}

				yield(nil, err)
				return
			}

			for items := range slices.Chunk(items, 2) {
				if !yield(&NameAndValue[string, string]{Name: items[0], Value: items[1]}, nil) {
					return
				}
			}
		}
	}
}

func (c *Client) LPopValues(ctx context.Context, key string) iter.Seq2[string, error] {
	return LPopValues(ctx, c.Client, key)
}

func LPopValues(ctx context.Context, c redis.Cmdable, key string) iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {
		for {
			v, err := LPop(ctx, c, key)
			if err != nil {
				if IsNoValue(err) {
					yield("", err)
					return
				}

				yield("", err)
				return
			}

			if !yield(v, nil) {
				return
			}
		}
	}
}
