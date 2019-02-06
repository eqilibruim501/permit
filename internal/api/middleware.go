package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/crusttech/permit/internal/context"
	"github.com/crusttech/permit/internal/rand"
)

const (
	HTTP_HEADER_REQUEST_ID = "request-id"
)

func requestLogMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var requestId string

		ctx := c.Request.Context()

		requestId = c.Request.Header.Get(HTTP_HEADER_REQUEST_ID)
		if len(requestId) == 0 {
			requestId = string(rand.RandBytesMaskImprSrc(32))
		}

		ctx = context.WithRequestId(ctx, requestId)

		// Benchmark the requrst
		start := time.Now()

		// Make sure client gets the request id
		c.Header(HTTP_HEADER_REQUEST_ID, requestId)
		c.Set(HTTP_HEADER_REQUEST_ID, requestId)

		// Append request id to all logs that will emit from the API
		logger := logger.With(zap.String("requestId", requestId))

		// Modified context, with logger
		ctx = context.WithLogger(ctx, logger)

		// Modified request, with new context
		c.Request = c.Request.WithContext(ctx)

		/////////////////

		// Handle this request
		c.Next()

		/////////////////

		// Context is updated, at least with user's ID
		ctx = c.Request.Context()

		end := time.Now()
		latency := end.Sub(start).Seconds()

		status := c.Writer.Status()

		// Use log instance from context
		// (and all fields we've set to it along the way...)
		requestLog := context.Log(ctx).With(
			zap.Int("status", status),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int64("requestContentLength", c.Request.ContentLength),
			zap.Int("responseContentLength", c.Writer.Size()),
			zap.String("ip", c.ClientIP()),
			zap.Float64("latency", latency),
		)

		pfix := "HTTP response status: "

		if status >= http.StatusInternalServerError && len(c.Errors) > 0 {
			requestLog.
				With(zap.Strings("errors", c.Errors.Errors())).
				Info(pfix + http.StatusText(status))
		} else if status >= http.StatusBadRequest {
			requestLog.
				With(zap.Strings("errors", c.Errors.Errors())).
				Info(pfix + http.StatusText(status))
		} else if status >= http.StatusMultipleChoices {
			requestLog.
				With(zap.String("redirectionUrl", c.GetHeader("Location"))).
				Info(pfix + http.StatusText(status))

		} else {
			requestLog.
				Debug(pfix + http.StatusText(status))
		}
	}
}
