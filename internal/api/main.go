package api

import (
	"net/http"
	"time"

	"github.com/SentimensRG/sigctx"
	"github.com/cnjack/throttle"
	"github.com/gin-gonic/contrib/jwt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/crusttech/permit/internal/env"
	"github.com/crusttech/permit/pkg/permit"
)

type (
	permitKeeper interface {
		Get(key string) (*permit.Permit, error)
		Create(p permit.Permit) error
	}

	jsonError struct {
		Error string `json:"error"`
	}
)

const maxRequestsPerHour = 60

func Serve(storage permitKeeper) {
	var jwtSecret string
	var g *gin.RouterGroup
	log, err := setupLogger(env.GetBoolEnv("LOG_PRETTY"), "debug")
	if err != nil {
		panic("Unable to setup logging")
	}

	if storage == nil {
		panic("Missing storage")
	}

	if jwtSecret = env.GetStringEnv("JWT_SECRET", ""); jwtSecret == "" {
		panic("JWT_SECRET missing")
	}

	ctx := sigctx.New()

	gin.SetMode(env.GetStringEnv("GIN_MODE", gin.DebugMode))
	router := gin.New()

	router.Use(requestLogMiddleware(log))

	g = router.Group("/key")
	g.Use(jwt.Auth(jwtSecret))
	g.POST("", endpointKeyCreate(storage))
	// router.GET("/key/:key", endpointKeyRead(storage))

	// Key check with throttling
	g = router.Group("/check")
	g.Use(throttle.Policy(&throttle.Quota{
		Limit:  maxRequestsPerHour,
		Within: time.Hour,
	}))
	g.POST("", endpointKeyCheck(storage))

	// Catch all path
	router.Any("/", func(ctx *gin.Context) {
		ctx.Header("Content-Type", "text/html; charset=utf-8")
		ctx.String(
			http.StatusOK,
			`<html><body style="text-align: center; margin: 50px;"><h1>Crust subscription server</h1></body></html>`,
		)
	})

	srv := &http.Server{
		Addr:    env.GetStringEnv("API_LISTEN", "localhost:80"),
		Handler: router,
	}

	defer func() {
		log.Info("Shutting down HTTP server")
		srv.Shutdown(ctx)
	}()

	go func() {
		log.Info("Starting HTTP API server", zap.String("addr", srv.Addr))

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.With(zap.Error(err)).
				Fatal("Could not start HTTP API server")
		}

	}()

	select {
	case <-ctx.Done():
		break
	}
}

func newJsonError(err interface{}) jsonError {
	switch val := err.(type) {
	case string:
		return jsonError{val}
	case error:
		return jsonError{val.Error()}
	default:
		return jsonError{"unexpected error type"}
	}
}
