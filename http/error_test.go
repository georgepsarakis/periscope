package http

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewServerError(t *testing.T) {
	type args struct {
		ctx context.Context
		err error
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "nil error",
			args: args{err: nil, ctx: context.Background()},
			want: `{"code":500, "message":"Server Error"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := NewServerError(tt.args.ctx, tt.args.err)
			assert.JSONEq(t, tt.want, string(resp))
		})
	}
}
