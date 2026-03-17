package log

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	zapLogger *zap.Logger
)

// InitLogger initializes the logger with console and file output
func InitLogger() error {
	return initLoggerInternal(false)
}

// InitLoggerFileOnly initializes the logger with file output only (for MCP mode)
func InitLoggerFileOnly() error {
	return initLoggerInternal(true)
}

func initLoggerInternal(fileOnly bool) error {
	// Define encoder configuration for human-readable + machine-friendly output
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	// Open the log file
	logFile, err := os.OpenFile("lazy-ai-coder.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	var core zapcore.Core
	if fileOnly {
		// MCP mode: log to file only (stdout/stderr are used for MCP protocol)
		fileEncoder := zapcore.NewJSONEncoder(encoderCfg)
		core = zapcore.NewCore(fileEncoder, zapcore.AddSync(logFile), zap.InfoLevel)
	} else {
		// Normal mode: log to both console and file
		consoleEncoder := zapcore.NewJSONEncoder(encoderCfg)
		fileEncoder := zapcore.NewJSONEncoder(encoderCfg)
		core = zapcore.NewTee(
			zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stderr), zap.InfoLevel),
			zapcore.NewCore(fileEncoder, zapcore.AddSync(logFile), zap.InfoLevel),
		)
	}

	zapLogger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.WarnLevel))
	defer zapLogger.Sync() // flushes buffer, if any

	return nil
}

func GetLogger() *zap.SugaredLogger {
	return zapLogger.Sugar()
}
