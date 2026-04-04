# syntax=docker/dockerfile:1.7

# 构建阶段：使用 Go 官方镜像编译静态二进制。
FROM --platform=$BUILDPLATFORM golang:1.24-alpine AS builder

WORKDIR /src

# 先复制依赖描述文件，尽量利用 Docker 构建缓存。
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

# 支持 buildx 透传目标平台参数，默认构建 Linux 二进制。
ARG TARGETOS=linux
ARG TARGETARCH=amd64
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags="-s -w" -o /out/dnsherene ./cmd/dnsherene

# 运行阶段：保留 cron 所需的 Alpine 用户态工具。
FROM alpine:3.21

# ca-certificates 用于 HTTPS 请求，su-exec 用于降权执行，tzdata 用于时区支持。
RUN apk add --no-cache ca-certificates su-exec tzdata

# 创建非 root 用户，实际业务命令以该用户身份运行。
RUN adduser -D -h /home/dnshe dnshe

COPY --from=builder /out/dnsherene /usr/local/bin/dnsherene
COPY docker/entrypoint.sh /usr/local/bin/docker-entrypoint.sh

RUN chmod +x /usr/local/bin/docker-entrypoint.sh

# 默认进入 cron 模式；如需立即执行一次，可改用 `run` 子命令。
ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]
CMD ["cron"]
