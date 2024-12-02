package metrics

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCollectProcessInfo(t *testing.T) {
	tests := []struct {
		name    string
		pid     func() int32
		wantErr bool
	}{
		{
			name: "pid is invalid",
			pid: func() int32 {
				return -1
			},
			wantErr: true,
		},
		{
			name: "pid is valid",
			pid: func() int32 {
				return int32(os.Getpid())
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CollectProcessInfo(tt.pid())
			assert.Equal(t, err != nil, tt.wantErr)
		})
	}
}

func TestRecordProcessRequestMetrics(t *testing.T) {
	tests := []struct {
		name     string
		pid      func() int32
		traceID  string
		duration float64
		wantErr  bool
	}{
		{
			name:     "pid is invalid",
			traceID:  "111-222-333",
			duration: float64(1),
			pid: func() int32 {
				return -1
			},
			wantErr: true,
		},
		{
			name:     "pid is valid",
			traceID:  "111-222-333",
			duration: float64(1),
			pid: func() int32 {
				return int32(os.Getpid())
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RecordProcessRequestMetrics(tt.pid(), tt.traceID, tt.duration)
			assert.Equal(t, err != nil, tt.wantErr)
		})
	}
}

func TestRecordProcessResourceUsageMetrics(t *testing.T) {
	tests := []struct {
		name    string
		pid     func() int32
		wantErr bool
	}{
		{
			name: "pid is invalid",
			pid: func() int32 {
				return -1
			},
			wantErr: true,
		},
		{
			name: "pid is valid",
			pid: func() int32 {
				return int32(os.Getpid())
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RecordProcessResourceUsageMetrics(tt.pid())
			assert.Equal(t, err != nil, tt.wantErr)
		})
	}
}
