# syntax=docker/dockerfile:1

# Build the web UI into the server's embed directory
FROM node:24-alpine AS web
RUN corepack enable
WORKDIR /src/web
COPY web/package.json web/pnpm-lock.yaml web/pnpm-workspace.yaml ./
RUN --mount=type=cache,target=/root/.local/share/pnpm/store \
    pnpm install --frozen-lockfile
COPY web/ ./
RUN pnpm build

# Build the server binary with the web UI embedded
FROM golang:1.27-alpine AS server
WORKDIR /src/server
COPY server/go.mod server/go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download
COPY server/ ./
COPY --from=web /src/server/internal/webui/dist ./internal/webui/dist
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/simplemavenrepo .

FROM alpine:3.22
ARG VERSION=""
RUN addgroup -S smr && adduser -S -G smr -H smr \
    && mkdir -p /data/repo /data/config \
    && chown -R smr:smr /data
COPY --from=server /out/simplemavenrepo /usr/local/bin/simplemavenrepo

ENV SMR_HOST=0.0.0.0 \
    SMR_PORT=4321 \
    SMR_REPO_PATH=/data/repo \
    SMR_CONFIG_PATH=/data/config \
    SMR_VERSION=${VERSION}

USER smr
WORKDIR /data
VOLUME ["/data/repo", "/data/config"]
EXPOSE 4321

HEALTHCHECK --interval=30s --timeout=3s --start-period=10s \
    CMD wget -q --spider "http://127.0.0.1:${SMR_PORT}/livez" || exit 1

ENTRYPOINT ["simplemavenrepo"]
