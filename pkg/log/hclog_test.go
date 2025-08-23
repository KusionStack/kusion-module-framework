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

func TestGetRotationConfigs(t *testing.T) {
	tests := []struct {
		name             string
		input            map[string]string
		wantedMaxSize    int
		wantedMaxBackups int
		wantedMaxAge     int
	}{
		{
			name:             "empty input",
			input:            nil,
			wantedMaxSize:    10,
			wantedMaxBackups: 10,
			wantedMaxAge:     28,
		},
		{
			name: "empty input values",
			input: map[string]string{
				KusionModuleLogMaxSize:    "",
				KusionModuleLogMaxBackups: "",
				KusionModuleLogMaxAge:     "",
			},
			wantedMaxSize:    10,
			wantedMaxBackups: 10,
			wantedMaxAge:     28,
		},
		{
			name: "invalid input values",
			input: map[string]string{
				KusionModuleLogMaxSize:    "?",
				KusionModuleLogMaxBackups: "?",
				KusionModuleLogMaxAge:     "?",
			},
			wantedMaxSize:    10,
			wantedMaxBackups: 10,
			wantedMaxAge:     28,
		},
		{
			name: "customized valid input values",
			input: map[string]string{
				KusionModuleLogMaxSize:    "100",
				KusionModuleLogMaxBackups: "100",
				KusionModuleLogMaxAge:     "100",
			},
			wantedMaxSize:    100,
			wantedMaxBackups: 100,
			wantedMaxAge:     100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := metadata.NewIncomingContext(context.Background(), metadata.New(tt.input))
			if gotMaxSize, gotMaxBackups, gotMaxAge := getRotationConfigs(ctx); gotMaxSize != tt.wantedMaxSize ||
				gotMaxBackups != tt.wantedMaxBackups || gotMaxAge != tt.wantedMaxAge {
				t.Errorf("getRotationConfigs() = %d, %d, %d want %d, %d, %d", gotMaxSize, gotMaxBackups, gotMaxAge,
					tt.wantedMaxSize, tt.wantedMaxBackups, tt.wantedMaxAge)
			}
		})
	}
}
