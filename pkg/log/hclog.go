package log

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/hashicorp/go-hclog"
	"google.golang.org/grpc/metadata"
	"gopkg.in/natefinch/lumberjack.v2"

	"kusionstack.io/kusion-module-framework/pkg/util/kfile"
)

const (
	KusionModuleName = "kusion_module_name"
	KusionTraceID    = "kusion_trace_id"
)

func GetModuleLogger(ctx context.Context) hclog.Logger {
	traceID := getTraceID(ctx)
	moduleName := getModuleName(ctx)
	if moduleName == "" {
		moduleName = "kusionstack-default-module"
	}

	kusionDataDir, _ := kfile.KusionDataFolder()
	logFile := filepath.Join(kusionDataDir, Folder, "modules", moduleName, fmt.Sprintf("%s.log", moduleName))
	lumberjackLogger := &lumberjack.Logger{
		Filename:  logFile,
		MaxSize:   10,
		Compress:  false,
		LocalTime: true,
		MaxAge:    28,
	}

	return hclog.New(&hclog.LoggerOptions{
		Name:   moduleName,
		Output: lumberjackLogger,
		Level:  hclog.Debug,
	}).With("trace_id", traceID)
}

func getTraceID(ctx context.Context) string {
	ids := metadata.ValueFromIncomingContext(ctx, KusionTraceID)
	if len(ids) == 0 {
		return ""
	}
	return ids[0]
}

func getModuleName(ctx context.Context) string {
	names := metadata.ValueFromIncomingContext(ctx, KusionModuleName)
	if len(names) == 0 {
		return ""
	}
	return names[0]
}
