package templates

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_format(t *testing.T) {
	tests := []struct {
		name  string
		input uint64
		want  string
	}{
		{
			name:  "zero",
			input: 0,
			want:  "0",
		},
		{
			name:  "one",
			input: 1,
			want:  "1",
		},
		{
			name:  "2 digits",
			input: 12,
			want:  "12",
		},
		{
			name:  "3 digits",
			input: 123,
			want:  "123",
		},
		{
			name:  "4 digits",
			input: 1234,
			want:  "1.2k",
		},
		{
			name:  "5 digits",
			input: 12345,
			want:  "12.3k",
		},
		{
			name:  "6 digits",
			input: 123456,
			want:  "123.4k",
		},
		{
			name:  "7 digits",
			input: 1234567,
			want:  "1.2M",
		},
		{
			name:  "8 digits",
			input: 12345678,
			want:  "12.3M",
		},
		{
			name:  "9 digits",
			input: 123456789,
			want:  "123.4M",
		},
		{
			name:  "10 digits",
			input: 1234567890,
			want:  "1.2G",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := format(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_formatPercentage(t *testing.T) {
	tests := []struct {
		name  string
		input float64
		want  string
	}{
		{
			name:  "positive",
			input: 0.123,
			want:  "+0.12%",
		},
		{
			name:  "negative",
			input: -0.123,
			want:  "-0.12%",
		},
		{
			name:  "zero",
			input: 0,
			want:  "+0.00%",
		},
		{
			name:  "1%",
			input: 1,
			want:  "+1.00%",
		},
		{
			name:  "100%",
			input: 100,
			want:  "+100.00%",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatPercentage(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}
