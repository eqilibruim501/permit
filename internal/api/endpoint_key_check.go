package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/crusttech/permit/internal/context"
	"github.com/crusttech/permit/pkg/permit"
)

const minKeyLen = 4
const maxKeyLen = 100

func endpointKeyCheck(storage permitKeeper) gin.HandlerFunc {
	if storage == nil {
		return func(ctx *gin.Context) {
			ctx.AbortWithStatus(http.StatusBadRequest)
		}
	}

	return func(ctx *gin.Context) {
		var (
			err error
			p   *permit.Permit
			req = permit.Permit{}
			log = context.Log(ctx.Request.Context())
		)

		if err = ctx.BindJSON(&req); err != nil {
			log.With(zap.Error(err)).Error("could not decode request")
			ctx.JSON(http.StatusInternalServerError, newJsonError(errors.Wrap(err, "could not decode request")))
			return
		}

		log = log.With(zap.String("key", req.Key), zap.String("domain", req.Domain))

		if !permit.ValidateDomain(req.Domain) {
			ctx.JSON(http.StatusBadRequest, newJsonError("invalid domain"))
			return
		}

		if p, err = storage.Get(req.Key); err != nil || p == nil {
			if err == permit.PermitNotFound {
				ctx.AbortWithStatus(http.StatusNotFound)
			} else {
				log.With(zap.Error(err)).Error("could not fetch permit")
				ctx.JSON(http.StatusInternalServerError, newJsonError(errors.Wrap(err, "could not fetch permit")))
			}
			return

		} else if p.Domain != req.Domain {
			log.Warn("domain mismatch")
			ctx.JSON(http.StatusUnauthorized, newJsonError("domain mismatch"))
			return

		} else if !p.IsValid() {
			log.Warn("permit not valid")
			ctx.JSON(http.StatusUnauthorized, newJsonError("permit not valid"))
			return

		} else if p.Expires != nil {
			ctx.Header("Expires", p.Expires.Format(time.RFC1123))
		}

		fields := []zap.Field{}
		for k, v := range req.Attributes {
			fields = append(fields, zap.Int("attributes."+k, v))
		}

		log.Info("permit check ok", fields...)

		ctx.JSON(http.StatusOK, p)
	}
}
