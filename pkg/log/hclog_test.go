package log

import (
	"context"
	"testing"

	"google.golang.org/grpc/metadata"
)

func TestGetTraceID(t *testing.T) {
	tests := []struct {
		name  string
		input map[string]string
		want  string
	}{
		{
			name:  "input is empty",
			input: nil,
			want:  "",
		},
		{
			name:  "input trace id is empty",
			input: map[string]string{},
			want:  "",
		},
		{
			name: "input trace id is not empty",
			input: map[string]string{
				KusionTraceID: "f3fa93ae-ac9a-11ef-aca1-acde48001122",
			},
			want: "f3fa93ae-ac9a-11ef-aca1-acde48001122",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			ctx = metadata.NewIncomingContext(ctx, metadata.New(tt.input))
			if got := getTraceID(ctx); got != tt.want {
				t.Errorf("getTraceID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetModuleName(t *testing.T) {
	tests := []struct {
		name  string
		input map[string]string
		want  string
	}{
		{
			name:  "input is empty",
			input: nil,
			want:  "",
		},
		{
			name:  "input module name is empty",
			input: map[string]string{},
			want:  "",
		},
		{
			name: "input module name is not empty",
			input: map[string]string{
				KusionModuleName: "service",
			},
			want: "service",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			ctx = metadata.NewIncomingContext(ctx, metadata.New(tt.input))
			if got := getModuleName(ctx); got != tt.want {
				t.Errorf("getModuleName() = %v, want %v", got, tt.want)
			}
		})
	}
}
