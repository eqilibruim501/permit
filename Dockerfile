## builder image
FROM golang:1.10-alpine AS builder

WORKDIR /go/src/github.com/crusttech/permit

ENV CGO_ENABLED=0

COPY . .

RUN apk update && apk upgrade && apk add --no-cache git
RUN mkdir /build; \
    go build \
        -o /build/cli cmd/cli/*.go

## target image
FROM alpine:3.7

RUN mkdir -p /permit /storage

ENV PATH="/permit:{$PATH}"
WORKDIR /permit
VOLUME /storage

COPY --from=builder /build/* /permit/
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

EXPOSE 80

ENTRYPOINT [ "/permit/cli" ]
CMD [ "api" ]
