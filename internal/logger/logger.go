package logger

import (
	"fmt"

	"go.uber.org/zap"
)

var Logger *zap.Logger

type LoggerConfig struct {
	Level string
}

func defaultLoggerConfig() LoggerConfig {
	return LoggerConfig{
		Level: "info",
	}
}

func Init(cf map[string]any) {
	var conf LoggerConfig
	if cf == nil {
		conf = defaultLoggerConfig()
	}
	fmt.Println(conf)

	// core := zapcore.NewCore(
	// 	zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
	// 	// zapcore.AddSync(wr),
	// 	zap.NewAtomicLevelAt(zap.DebugLevel),
	// )

	// logger := zap.New(core)

	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	Logger = logger
}

func Sync() {
	Logger.Sync()
}

func Debug(msg string, fields ...zap.Field) {
	Logger.Info(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	Logger.Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	Logger.Info(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	Logger.Info(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	Logger.Fatal(msg, fields...)
}

// func Infof(format string, args ...interface{}) {
// 	Logger.Sugar().Infof(format, args...)
// }
