package api

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func setupLogger(debug bool, level string) (logger *zap.Logger, err error) {
	if debug {
		return zap.NewDevelopment()
	}

	var l = zapcore.DebugLevel

	if err := l.Set(level); err != nil {
		return nil, err
	}

	conf := zap.Config{
		Level:            zap.NewAtomicLevelAt(l),
		Development:      false,
		Encoding:         "json",
		EncoderConfig:    zap.NewProductionEncoderConfig(),
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}

	// Make logstash's life a bit easier
	conf.EncoderConfig.LevelKey = "appLogLevel"
	conf.EncoderConfig.MessageKey = "message"
	conf.EncoderConfig.TimeKey = "@timestamp"
	conf.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	return conf.Build()
}
