package pkg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntToBase62(t *testing.T) {
	tests := []struct {
		name string
		num  int
		want string
	}{
		{
			name: "zero",
			num:  0,
			want: "0",
		},
		{
			name: "negative number",
			num:  -1234,
			want: "-jU",
		},
		{
			name: "positive number",
			num:  89012,
			want: "n9G",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, IntToBase62(tt.num))
		})
	}
}

func TestZerofill(t *testing.T) {
	type args struct {
		s      string
		length int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "empty string",
			args: args{s: "", length: 5},
			want: "00000",
		},
		{
			name: "string length exceeds total length",
			args: args{s: "1234567", length: 6},
			want: "1234567",
		},
		{
			name: "string length is less than total length",
			args: args{s: "123", length: 6},
			want: "000123",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, Zerofill(tt.args.s, tt.args.length), "Zerofill(%v, %v)", tt.args.s, tt.args.length)
		})
	}
}
