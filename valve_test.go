package valve_test

import (
	"testing"

	"github.com/ardnew/valve"
	"github.com/stretchr/testify/assert"
)

func TestIO_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		want string
		io   valve.IO
	}{
		{
			name: "Read",
			want: "read",
			io:   valve.Read,
		},
		{
			name: "Write",
			want: "write",
			io:   valve.Write,
		},
		{
			name: "Close",
			want: "close",
			io:   valve.Close,
		},
		{
			name: "ReadWrite",
			want: "read/write",
			io:   valve.ReadWrite,
		},
		{
			name: "NOP",
			want: "nop",
			io:   valve.NOP,
		},
		{
			name: "DEADBEEF",
			want: "invalid",
			io:   valve.DEADBEEF,
		},
		{
			name: "Unknown",
			want: "unknown",
			io:   valve.IO(10),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.io.String()
			assert.Equal(t, tt.want, got)
		})
	}
}
