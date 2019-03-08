package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/crusttech/permit/internal/context"
	"github.com/crusttech/permit/internal/rand"
	"github.com/crusttech/permit/pkg/permit"
)

const minDomainLen = 4
const maxDomainLen = 100

func endpointKeyCreate(storage permitKeeper) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var (
			err error
			p   = permit.Permit{}
			req = struct {
				Domain string `json:"domain"`
				// Expires    *time.Time     `json:"expires,omitempty"`
				Attributes map[string]int `json:"attributes"`
				Contact    string         `json:"contact"`
				Entity     string         `json:"entity"`
				Type       string         `json:"type,omitempty"`
			}{}
			log = context.Log(ctx.Request.Context())
		)

		if err = ctx.BindJSON(&req); err != nil {
			log.With(zap.Error(err)).Error("could not decode request")
			ctx.JSON(http.StatusInternalServerError, newJsonError(errors.Wrap(err, "could not decode request")))
			return
		}

		if !permit.ValidateDomain(req.Domain) {
			ctx.JSON(http.StatusBadRequest, newJsonError("invalid domain"))
			return
		}

		{
			var (
				now      = time.Now().Truncate(time.Second)
				tomorrow = time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
			)

			p.Expires = &time.Time{}

			switch req.Type {
			case "trial":
				// Trial keys, 14 days.
				*p.Expires = tomorrow.AddDate(0, 0, 14)
			// case "unlimited":
			// 	p.Expires = nil
			default:
				// By default, offset expiration date by 1 year
				*p.Expires = tomorrow.AddDate(1, 0, 0)
			}

			log = log.With(zap.Time("expires", *p.Expires))
		}

		p.Domain = req.Domain
		p.Contact = req.Contact
		p.Entity = req.Entity
		p.Key = string(rand.RandBytesMaskImprSrc(permit.KeyLength))
		p.Valid = true
		p.Version = 1
		p.Issued = time.Now().Truncate(time.Second)

		log = log.With(
			zap.String("domain", req.Domain),
			zap.String("key", p.Key),
			zap.String("contact", p.Contact),
			zap.String("entity", p.Entity),
		)

		if req.Attributes == nil {
			// No permissions send with request
			p.Attributes = permit.DefaultAttributes
		} else {
			// Iterate over default attributes and make sure only predefined keys
			// from DefaultAttributes are copied. Ignore the rest and set defaults
			// where keys are missing
			for defKey, defValue := range permit.DefaultAttributes {
				if reqAttribVal, has := req.Attributes[defKey]; !has {
					p.Attributes[defKey] = defValue
				} else {
					p.Attributes[defKey] = reqAttribVal
				}
			}
		}

		if err = storage.Create(p); err != nil {
			log.With(zap.Error(err)).Error("could not store permit")
			ctx.JSON(http.StatusInternalServerError, newJsonError(errors.Wrap(err, "could not store permit")))
		}

		fields := []zap.Field{}
		for k, v := range req.Attributes {
			fields = append(fields, zap.Int("attributes."+k, v))
		}

		log.Info("permit created", fields...)

		ctx.JSON(http.StatusOK, p)
	}
}
