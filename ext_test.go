package redisx

import (
	"context"
	"strings"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestScanValues(t *testing.T) {
	type args struct {
		prefix string
	}
	tests := [...]struct {
		name    string
		args    args
		wantErr bool
	}{
		{`valid`, args{"test."}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			r := localRedis
			r.Set(ctx, tt.args.prefix+"scan", "hello")

			keys := []string{}
			for key, err := range localRedis.ScanValues(ctx, ScanOpts{Match: tt.args.prefix + "*"}) {
				require.NoError(t, err)

				keys = append(keys, key)
				require.Truef(t, strings.HasPrefix(key, tt.args.prefix), "prefix=%s, key=%s", tt.args.prefix, key)
				break
			}
		})
	}
}

func TestScanValuesWithPipelined(t *testing.T) {
	type args struct {
		prefix string
	}
	tests := [...]struct {
		name    string
		args    args
		wantErr bool
	}{
		{`valid`, args{""}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			r := localRedis

			r.Pipelined(ctx, func(p redis.Pipeliner) error {
				for key, err := range ScanValues(ctx, p, ScanOpts{Match: tt.args.prefix + "*"}) {
					require.NoError(t, err)

					t.Logf("key=%s", key)
				}

				return nil
			})
		})
	}
}
