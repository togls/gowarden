ARG GO_VERSION=1.18.4
ARG ALPINE_VERSION=3.16

FROM vaultwarden/web-vault:v2022.6.2 as web

FROM --platform=${BUILDPLATFORM} golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS builder

RUN apk update && \
    apk add --no-cache \
    ca-certificates \
    build-base \
    gcc \
    git \
    make \
    tzdata

WORKDIR /app

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    make build

FROM alpine:${ALPINE_VERSION}

# zoneinfo
ENV TZ=Asia/Shanghai
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

WORKDIR /app

COPY --from=web /web-vault /app/web
COPY --from=builder /app/gowarden /app/gowarden
COPY --from=builder /app/static /app/static
COPY --from=builder /app/config.json.example /app/config.json

CMD [ "/app/gowarden", "--config", "/app/config.json" ]
