package logger

import (
	"fmt"
	"os"
	"skyfox/config"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var once sync.Once

var lgr *zap.Logger

func GetLogger() *zap.Logger {

	if lgr == nil {
		lcfg := config.LoggerConfig{
			Level: "info",
		}
		InitAppLogger(lcfg)
	}
	return lgr
}

func InitAppLogger(lcfg config.LoggerConfig) {
	once.Do(func() {
		lgr = newZapLogger(getLevel(lcfg.Level))
	})
}

func Debug(msg string, args ...interface{}) {
	GetLogger().Debug(format(msg, args...))
}

func Info(msg string, args ...interface{}) {
	GetLogger().Info(format(msg, args...))
}

func Warn(msg string, args ...interface{}) {
	GetLogger().Warn(format(msg, args...))
}

func Error(msg string, args ...interface{}) {
	GetLogger().Error(format(msg, args...))
}

func Fatal(msg string, args ...interface{}) {
	GetLogger().Fatal(format(msg, args...))
}

func format(template string, args ...interface{}) string {
	return fmt.Sprintf(template, args...)
}

func newZapLogger(level zapcore.Level) *zap.Logger {
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig()),
		zapcore.NewMultiWriteSyncer(os.Stdout),
		zap.NewAtomicLevelAt(level),
	)

	return zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
}

func encoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.NanosDurationEncoder,
	}
}

var LevelMap = map[string]zapcore.Level{
	"debug": zapcore.DebugLevel,
	"info":  zapcore.InfoLevel,
	"warn":  zapcore.WarnLevel,
	"error": zapcore.ErrorLevel,
}

func getLevel(level string) zapcore.Level {
	return LevelMap[level]
}
