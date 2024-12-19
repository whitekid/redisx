package redisx

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/whitekid/goxp/errors"
)

func TestIsNotFound(t *testing.T) {
	type args struct {
		err error
	}
	tests := [...]struct {
		name string
		args args
		want bool
	}{
		{`valid`, args{ErrNotFound}, true},
		{`valid`, args{errors.Wrap(ErrNotFound, "XX")}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsNotFound(tt.args.err)
			require.Equal(t, tt.want, got, `IsNotFound() failed: got = %+v, want = %v`, got, tt.want)
		})
	}
}
