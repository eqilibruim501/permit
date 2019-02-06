package context

import (
	"context"
	"time"

	"go.uber.org/zap"
)

type (
	loggerKey    struct{}
	requestIdKey struct{}

	Context = context.Context
)

func Background() context.Context {
	return context.Background()
}

func WithLogger(ctx Context, logger *zap.Logger) Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

func WithValue(ctx Context, key interface{}, val interface{}) Context {
	return context.WithValue(ctx, key, val)
}

func WithRequestId(ctx Context, requestId string) Context {
	return context.WithValue(ctx, requestIdKey{}, requestId)
}

func WithTimeout(parent Context, timeout time.Duration) (Context, context.CancelFunc) {
	return context.WithTimeout(parent, timeout)
}

func Log(ctx Context) *zap.Logger {
	return MustLog(ctx, zap.NewNop())
}

func MustLog(ctx Context, fallback *zap.Logger) *zap.Logger {
	if log, ok := ctx.Value(loggerKey{}).(*zap.Logger); ok {
		return log
	}

	return fallback
}

func RequestId(ctx Context) string {
	if requestId, ok := ctx.Value(requestIdKey{}).(string); ok {
		return requestId
	}

	return ""
}
