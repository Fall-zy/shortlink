package utils

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger

func InitLogger(level string) {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	writeSyncer := zapcore.AddSync(os.Stdout)

	var zapLevel zapcore.Level
	switch level {
	case "debug":
		zapLevel = zap.DebugLevel
	case "warn":
		zapLevel = zap.WarnLevel
	case "error":
		zapLevel = zap.ErrorLevel
	default:
		zapLevel = zap.InfoLevel
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		writeSyncer,
		zapLevel,
	)

	Logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
}

func Sync() {
	if Logger != nil {
		_ = Logger.Sync()
	}
}
