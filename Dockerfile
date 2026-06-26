# RoundTable server — multi-stage build (Discord bot + HTTP health server)
# Build: docker compose build
# Run:   docker compose up -d discord
#
# Optional build proxy (ShellCrash mixed-port on host):
#   HTTP_PROXY=http://127.0.0.1:4567 HTTPS_PROXY=http://127.0.0.1:4567 docker compose build discord

FROM golang:1.25-alpine AS builder

ARG HTTP_PROXY
ARG HTTPS_PROXY
ARG NO_PROXY=localhost,127.0.0.1

# Alpine official CDN is often unreachable from CN servers — use Aliyun mirror.
RUN sed -i 's|https://dl-cdn.alpinelinux.org|https://mirrors.aliyun.com|g' /etc/apk/repositories && \
    apk add --no-cache git ca-certificates

WORKDIR /src

ENV CGO_ENABLED=0
ENV GOPROXY=https://goproxy.cn,direct

COPY go.work go.work.sum ./
COPY apps/server/go.mod apps/server/go.sum ./apps/server/

WORKDIR /src/apps/server
RUN go mod download

COPY apps/server/ ./

RUN go build -trimpath -ldflags="-s -w" -o /out/roundtable-discord ./cmd/discord/main.go && \
    go build -trimpath -ldflags="-s -w" -o /out/roundtable-server ./cmd/roundtable/main.go

# --- runtime ---

FROM alpine:3.21

ARG HTTP_PROXY
ARG HTTPS_PROXY

RUN sed -i 's|https://dl-cdn.alpinelinux.org|https://mirrors.aliyun.com|g' /etc/apk/repositories && \
    apk add --no-cache ca-certificates tzdata wget su-exec && \
    adduser -D -u 1000 roundtable

WORKDIR /app

COPY --from=builder /out/roundtable-discord /out/roundtable-server /usr/local/bin/
COPY apps/server/configs ./apps/server/configs
COPY data/_templates ./data/_templates
COPY deploy/docker-entrypoint.sh /app/docker-entrypoint.sh

RUN chmod +x /app/docker-entrypoint.sh && \
    mkdir -p \
      data/workspaces \
      data/profiles/participants \
      data/profiles/principals \
      data/profiles/moderator \
      data/knowledge/participants \
      data/knowledge/principals \
      data/knowledge/shared \
      data/transport && \
    chown -R roundtable:roundtable /app

# entrypoint chowns mounted volumes then exec su-exec roundtable
ENTRYPOINT ["/app/docker-entrypoint.sh"]

ENV ROUND_TABLE_ROOT=/app/apps/server \
    ROUND_TABLE_WORKSPACE_ROOT=/app/data/workspaces \
    ROUND_TABLE_PROFILE_ROOT=/app/data/profiles \
    ROUND_TABLE_PROFILE_TEMPLATES=/app/data/_templates/profiles \
    ROUND_TABLE_KNOWLEDGE_ROOT=/app/data/knowledge \
    ROUND_TABLE_KNOWLEDGE_TEMPLATES=/app/data/_templates/knowledge \
    ROUND_TABLE_DISCORD_BINDINGS_FILE=/app/data/transport/discord-principal.json \
    TZ=Asia/Shanghai

EXPOSE 7777

CMD ["roundtable-discord"]
