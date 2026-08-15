package logger

import (
	"fmt"
	"os"
	"strings"

	"github.com/Shadowmaple/logflow/internal/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger

func getLogger() *zap.Logger {
	if Logger != nil {
		return Logger
	}
	return zap.NewExample()
}

func defaultLoggerConfig() *config.LoggerConfig {
	return &config.LoggerConfig{
		Level:  "debug",
		Format: "json",
		Output: "stdout",
	}
}

func parseLevel(level string) zapcore.Level {
	switch strings.ToLower(level) {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn", "warning":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

func buildEncoder(format string) zapcore.Encoder {
	var encoderCfg zapcore.EncoderConfig
	switch strings.ToLower(format) {
	case "console":
		encoderCfg = zap.NewDevelopmentEncoderConfig()
		encoderCfg.EncodeLevel = zapcore.CapitalLevelEncoder
		encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
		return zapcore.NewConsoleEncoder(encoderCfg)
	case "json":
		fallthrough
	default:
		encoderCfg = zap.NewProductionEncoderConfig()
		encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
		return zapcore.NewJSONEncoder(encoderCfg)
	}
}

func buildWriter(output, filePath string) (zapcore.WriteSyncer, error) {
	switch strings.ToLower(output) {
	case "stderr":
		return zapcore.AddSync(os.Stderr), nil
	case "file":
		if filePath == "" {
			filePath = "logflow.log"
		}
		f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, fmt.Errorf("open log file failed: %w", err)
		}
		return zapcore.AddSync(f), nil
	case "stdout":
		fallthrough
	default:
		return zapcore.AddSync(os.Stdout), nil
	}
}

func Init(conf *config.LoggerConfig) {
	if conf == nil {
		conf = defaultLoggerConfig()
	}

	level := parseLevel(conf.Level)
	encoder := buildEncoder(conf.Format)
	writer, err := buildWriter(conf.Output, conf.FilePath)
	if err != nil {
		// fallback to stdout if file open fails
		fmt.Fprintf(os.Stderr, "logger init warning: %v, fallback to stdout\n", err)
		writer = zapcore.AddSync(os.Stdout)
	}

	core := zapcore.NewCore(encoder, writer, level)
	Logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
}

func Sync() {
	getLogger().Sync()
}

func Debug(msg string, fields ...zap.Field) {
	getLogger().Debug(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	getLogger().Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	getLogger().Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	getLogger().Error(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	getLogger().Fatal(msg, fields...)
}
