package log

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/hashicorp/go-hclog"
	"google.golang.org/grpc/metadata"
	"gopkg.in/natefinch/lumberjack.v2"

	"kusionstack.io/kusion-module-framework/pkg/util/kfile"
)

const (
	KusionModuleName          = "kusion_module_name"
	KusionTraceID             = "kusion_trace_id"
	KusionModuleLogMaxSize    = "kusion_module_log_max_size"
	KusionModuleLogMaxBackups = "kusion_module_log_max_backups"
	KusionModuleLogMaxAge     = "kusion_module_log_max_age"
)

func GetModuleLogger(ctx context.Context) hclog.Logger {
	traceID := getTraceID(ctx)
	moduleName := getModuleName(ctx)
	if moduleName == "" {
		moduleName = "kusionstack-default-module"
	}

	kusionDataDir, _ := kfile.KusionDataFolder()
	logFile := filepath.Join(kusionDataDir, Folder, "modules", moduleName, fmt.Sprintf("%s.log", moduleName))
	maxSize, maxBackups, maxAge := getRotationConfigs(ctx)

	lumberjackLogger := &lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    maxSize,
		MaxBackups: maxBackups,
		MaxAge:     maxAge,
		Compress:   false,
		LocalTime:  true,
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

func getRotationConfigs(ctx context.Context) (logMaxSize, logMaxBackups, logMaxAge int) {
	var err error

	logMaxSizeStrs := metadata.ValueFromIncomingContext(ctx, KusionModuleLogMaxSize)
	if len(logMaxSizeStrs) == 0 {
		// Set the default max size of kusion module logger to 10MB.
		logMaxSize = 10
	} else {
		logMaxSize, err = strconv.Atoi(logMaxSizeStrs[0])
		if err != nil {
			logMaxSize = 10
		}
	}

	logMaxBackupStrs := metadata.ValueFromIncomingContext(ctx, KusionModuleLogMaxBackups)
	if len(logMaxBackupStrs) == 0 {
		// Set the default max backups of kusion module logger to 10.
		logMaxBackups = 10
	} else {
		logMaxBackups, err = strconv.Atoi(logMaxBackupStrs[0])
		if err != nil {
			logMaxBackups = 10
		}
	}

	logMaxAgeStrs := metadata.ValueFromIncomingContext(ctx, KusionModuleLogMaxAge)
	if len(logMaxAgeStrs) == 0 {
		// Set the default max age of kusion module logger to 28 days.
		logMaxAge = 28
	} else {
		logMaxAge, err = strconv.Atoi(logMaxAgeStrs[0])
		if err != nil {
			logMaxAge = 28
		}
	}

	return
}
