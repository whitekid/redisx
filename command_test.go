package redisx

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/require"
)

var localRedis *Client

func TestMain(m *testing.M) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var err error
	localRedis, err = NewLocalFactory()(ctx)
	if err != nil {
		panic(err)
	}

	m.Run()
}

func TestScan(t *testing.T) {
	type args struct {
		cursor uint64
		match  string
		count  int64
	}
	tests := [...]struct {
		name    string
		args    args
		wantErr bool
		keys    []string
		cursor  uint64
	}{
		{`valid`, args{0, "*", 1000}, false, []string{"key1", "key2"}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			r, m := redismock.NewClientMock()
			m.ExpectScan(tt.args.cursor, tt.args.match, tt.args.count).SetVal(tt.keys, tt.cursor)

			got, cursor, err := Scan(ctx, r, 0, "*", 1000)
			require.Truef(t, (err != nil) == tt.wantErr, `scan() failed: error = %+v, wantErr = %v`, err, tt.wantErr)
			if tt.wantErr {
				return
			}
			require.Equal(t, tt.keys, got)
			require.Equal(t, tt.cursor, cursor)
		})
	}
}

func TestExpire(t *testing.T) {
	type args struct {
		key string
		ttl time.Duration
	}
	tests := [...]struct {
		name    string
		args    args
		wantErr bool
		want    bool
	}{
		{`valid`, args{"key", time.Minute}, false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			r, m := redismock.NewClientMock()
			m.ExpectExpire(tt.args.key, tt.args.ttl).SetVal(true)

			got, err := Expire(ctx, r, tt.args.key, tt.args.ttl)
			require.Truef(t, (err != nil) == tt.wantErr, `Expire() failed: error = %+v, wantErr = %v`, err, tt.wantErr)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestLPopKey(t *testing.T) {
	type args struct {
		key    string
		values []any
	}
	tests := [...]struct {
		name    string
		args    args
		wantErr bool
	}{
		{`valid`, args{"test.lpop", []any{"double", "milk"}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// ensure key exists
			n, err := localRedis.RPush(ctx, tt.args.key, tt.args.values...)
			require.NoError(t, err)
			require.Equal(t, int64(len(tt.args.values)), n)

			values := []string{}
			for {
				got, err := LPop(ctx, localRedis.Client, tt.args.key)
				if err != nil {
					if IsNoValue(err) {
						break
					}
					require.Fail(t, "failed: %+v", err)
				}
				values = append(values, got)
			}
			require.Equal(t, len(tt.args.values), len(values))
		})
	}
}

func TestLPopKeyNotFound(t *testing.T) {
	type args struct {
		key string
	}
	tests := [...]struct {
		name    string
		args    args
		wantErr error
	}{
		{`valid`, args{"test.not-existing-key"}, ErrNoValue},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// ensure key not exists
			localRedis.Del(ctx, tt.args.key)

			_, err := LPop(ctx, localRedis.Client, tt.args.key)
			require.ErrorIs(t, ErrNoValue, err)
		})
	}
}

func TestTTL(t *testing.T) {
	type args struct {
		key    string
		expire time.Duration
	}
	tests := [...]struct {
		name    string
		args    args
		wantErr bool
	}{
		{`valid`, args{"test.ttl", time.Second}, false},
		{`valid`, args{"test.ttl", 0}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			localRedis.Set(ctx, tt.args.key, "hello", tt.args.expire)

			got, err := localRedis.TTL(ctx, tt.args.key)

			require.Truef(t, (err != nil) == tt.wantErr, `TTL() failed: error = %+v, wantErr = %v`, err, tt.wantErr)
			if tt.wantErr {
				return
			}

			if tt.args.expire == 0 {
				require.Equal(t, time.Duration(-1), got)
			} else {
				require.Equal(t, tt.args.expire, got)
			}
		})
	}
}
