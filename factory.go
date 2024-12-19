package redisx

import (
	"context"
	"crypto/tls"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/whitekid/goxp/errors"
)

type Factory func(ctx context.Context) (*Client, error)

// NewFactory returns a factory that creates a redis client
func NewFactory(host string, port int, database int, password string, option ...Option) Factory {
	return func(ctx context.Context) (*Client, error) {
		opts := &redis.Options{
			Addr:     fmt.Sprintf("%s:%d", host, port),
			Password: password,
			DB:       database,
		}

		var options Options

		for _, apply := range option {
			apply(&options)
		}

		if options.useTLS {
			opts.TLSConfig = &tls.Config{}
		}

		client := &Client{redis.NewClient(opts)}
		if _, err := client.Ping(ctx); err != nil {
			return nil, errors.Wrapf(err, "ping failed")
		}

		return client, nil
	}
}

// NewLocalFactory returns a factory that creates a redis client for local redis server
func NewLocalFactory() Factory { return NewFactory("127.0.0.1", 6379, 0, "") }

type Options struct {
	useTLS bool
}

type Option func(*Options)

func WithTLS(useTLS bool) Option { return func(o *Options) { o.useTLS = useTLS } }
