# 统一多阶段 Dockerfile — 构建上下文为 mybilibili-go/
#
# 普通服务:
#   docker build --build-arg SERVICE=core-service -t mybilibili-core .
# work 服务(带 ffmpeg+VAAPI):
#   docker build --build-arg SERVICE=work-service --target runtime-work -t mybilibili-work .
#
# SERVICE 取值:
#   core-service ai-service search-service msg-danmaku-service
#   live-service studio-service work-service bili-proxy-service
# CMD_DIR: 服务 cmd 下的子目录名（默认与服务名同名去 -service 后缀，bili-proxy 特殊）。
#   bili-proxy-service 需 --build-arg CMD_DIR=bili-proxy

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
COPY pkg/go.mod pkg/go.sum* ./pkg/
COPY ${SERVICE}/go.mod ${SERVICE}/go.sum* ./${SERVICE}/
RUN cd ${SERVICE} && go mod download
# 拷全部源码
COPY pkg ./pkg
COPY ${SERVICE} ./${SERVICE}/
# 编译：CMD_DIR 默认取 SERVICE 去掉 -service 后缀
RUN SVC="${SERVICE%-service}" && DIR="${CMD_DIR:-$SVC}" \
    && cd ${SERVICE} && go build -o /out/app ./cmd/${DIR}

# ============ 运行阶段: 普通服务 ============
FROM debian:bookworm-slim AS runtime
COPY --from=build /out/app /app
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && rm -rf /var/lib/apt/lists/*
ENTRYPOINT ["/app"]

# ============ 运行阶段: work 转码 (需 ffmpeg + ffprobe) ============
FROM debian:bookworm-slim AS runtime-work
COPY --from=build /out/app /app
RUN apt-get update && apt-get install -y --no-install-recommends \
        ca-certificates ffmpeg libva2 libva-drm2 \
    && rm -rf /var/lib/apt/lists/*
ENTRYPOINT ["/app"]
