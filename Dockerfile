# 统一多阶段 Dockerfile — 构建上下文为仓库根
#
# 用法:
#   docker build --build-arg SERVICE=core -t mybilibili-core .
#   docker build --build-arg SERVICE=bili-proxy --build-arg CMD_DIR=bili-proxy -t mybilibili-bili .
#
# SERVICE 取值（与 services/ 下目录名一致）:
#   core ai search msg-danmaku live studio work bili-proxy
# CMD_DIR: 服务 cmd 下的子目录名（默认与 SERVICE 相同）。
#   bili-proxy 服务需 --build-arg CMD_DIR=bili-proxy

ARG GO_IMAGE=golang:1.26-bookworm

# ============ 构建阶段 ============
FROM ${GO_IMAGE} AS build
ARG SERVICE
ARG CMD_DIR
ENV CGO_ENABLED=0 GOOS=linux GOFLAGS=-mod=mod GOWORK=off
# 国内网络: Go 模块走七牛代理(主) + 直连(备)，关闭校验加速
ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB=off

WORKDIR /build
# 先拷依赖清单利用缓存
COPY shared/pkg/go.mod shared/pkg/go.sum* ./shared/pkg/
COPY services/${SERVICE}/go.mod services/${SERVICE}/go.sum* ./services/${SERVICE}/
RUN cd services/${SERVICE} && go mod download
# 拷全部源码
COPY shared/pkg ./shared/pkg
COPY services/${SERVICE} ./services/${SERVICE}/
# 编译：CMD_DIR 默认与 SERVICE 相同
RUN DIR="${CMD_DIR:-${SERVICE}}" \
    && cd services/${SERVICE} && go build -o /out/app ./cmd/${DIR}

# ============ 开发阶段（热重载） ============
FROM ${GO_IMAGE} AS dev
ARG SERVICE
ENV CGO_ENABLED=0 GOOS=linux GOFLAGS=-mod=mod GOWORK=off
ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB=off
ENV SERVICE=${SERVICE}

RUN go install github.com/air-verse/air@latest

WORKDIR /app
COPY shared/pkg ./shared/pkg
COPY services/${SERVICE} ./services/${SERVICE}/
COPY entrypoint-dev.sh /usr/local/bin/entrypoint-dev.sh
RUN chmod +x /usr/local/bin/entrypoint-dev.sh

EXPOSE 8080
CMD ["/usr/local/bin/entrypoint-dev.sh"]

# ============ 运行阶段 ============
# 极简基底：Go 服务为 CGO_ENABLED=0 纯静态二进制，Alpine 只需 ca-certificates。
# 不装 ffmpeg/chrome——图片转 WebP 已前移到前端，转码由裸跑 transcoder 承担。
# 镜像体积对比：debian-bookworm-slim(~70MB) → alpine(~3.5MB + certs)。
#
# 启动模型（参考 postgres/redis/nginx 官方镜像）:
#   entrypoint.sh 以 root 身份负责初始化(数据目录 chown 等)，
#   然后 su-exec drop 到非 root 用户 (app) 跑业务二进制。
#   这样既保证安全 (二进制非 root 跑)，又解决 external volume
#   任意所有权导致的写权限问题。
FROM alpine:3.20 AS runtime
RUN apk add --no-cache ca-certificates tzdata su-exec \
    && addgroup -S app && adduser -S -G app app
COPY --from=build /out/app /app
COPY entrypoint.sh /usr/local/bin/entrypoint.sh
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
CMD ["/app"]
