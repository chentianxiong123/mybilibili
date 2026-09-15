# 编译产物输出目录。默认 ./bin (Go 生态约定), 可被命令行覆盖:
#   BIN=/custom/path make build
#   BIN=$HOME/bin   make build-transcoder
# 任何带 -o 的 in-source build 都用 clean-source 兜底清理.
BIN      ?= ./bin
GO       := go
LDFLAGS  := -ldflags="-s -w"

# 准备 BIN 目录
$(shell mkdir -p $(BIN))

.PHONY: run build clean build-core build-ai build-search build-msg-danmaku build-work build-live build-studio build-bili build-transcoder build-transcoder-nvenc build-transcoder-vaapi

run:
	$(GO) build $(LDFLAGS) -o $(BIN)/mybilibili-core ./core-service/cmd/core
	$(BIN)/mybilibili-core

build: build-core build-ai build-search build-msg-danmaku build-work build-live build-studio build-bili build-transcoder

build-core:
	$(GO) build $(LDFLAGS) -o $(BIN)/mybilibili-core ./core-service/cmd/core

build-ai:
	$(GO) build $(LDFLAGS) -o $(BIN)/mybilibili-ai ./ai-service/cmd/ai

build-search:
	$(GO) build $(LDFLAGS) -o $(BIN)/mybilibili-search ./search-service/cmd/search

build-msg-danmaku:
	$(GO) build $(LDFLAGS) -o $(BIN)/mybilibili-msg-danmaku ./msg-danmaku-service/cmd/msg-danmaku

build-work:
	$(GO) build $(LDFLAGS) -o $(BIN)/mybilibili-work ./work-service/cmd/work

build-studio:
	$(GO) build $(LDFLAGS) -o $(BIN)/mybilibili-studio ./studio-service/cmd/studio

build-live:
	$(GO) build $(LDFLAGS) -o $(BIN)/mybilibili-live ./live-service/cmd/live

build-bili:
	$(GO) build $(LDFLAGS) -o $(BIN)/mybilibili-bili ./bili-proxy-service/cmd/bili-proxy

# transcoder 一次编译三个版本（软编/VAAPI/NVENC），哪台机器用哪个
build-transcoder: build-transcoder-soft build-transcoder-vaapi build-transcoder-nvenc

build-transcoder-soft:
	$(GO) build $(LDFLAGS) -o $(BIN)/mybilibili-transcoder ./transcoder-service/cmd/transcoder

build-transcoder-vaapi:
	$(GO) build $(LDFLAGS) -tags vaapi -o $(BIN)/mybilibili-transcoder-vaapi ./transcoder-service/cmd/transcoder

build-transcoder-nvenc:
	$(GO) build $(LDFLAGS) -tags nvenc -o $(BIN)/mybilibili-transcoder-nvenc ./transcoder-service/cmd/transcoder

clean:
	rm -f $(BIN)/mybilibili-core $(BIN)/mybilibili-ai $(BIN)/mybilibili-search $(BIN)/mybilibili-msg-danmaku $(BIN)/mybilibili-work $(BIN)/mybilibili-studio $(BIN)/mybilibili-live $(BIN)/mybilibili-bili $(BIN)/mybilibili-transcoder

# 兜底清理: 扫描源码目录里所有产物, 防止某天 in-source build 污染.
# 用通配后缀, 不依赖服务名, 任何 *.exe / core / a.out 都会被识别.
clean-source:
	find . -type f \( -name 'core' -o -name 'a.out' -o -name '*.exe' -o -name '*.test' -o -name 'build-errors.log' \) -path './core-service/*' -delete 2>/dev/null || true
	find . -type f \( -name 'core' -o -name 'a.out' -o -name '*.exe' -o -name '*.test' -o -name 'build-errors.log' \) -path './ai-service/*' -delete 2>/dev/null || true
	find . -type f \( -name 'core' -o -name 'a.out' -o -name '*.exe' -o -name '*.test' -o -name 'build-errors.log' \) -path './search-service/*' -delete 2>/dev/null || true
	find . -type f \( -name 'core' -o -name 'a.out' -o -name '*.exe' -o -name '*.test' -o -name 'build-errors.log' \) -path './msg-danmaku-service/*' -delete 2>/dev/null || true
	find . -type f \( -name 'core' -o -name 'a.out' -o -name '*.exe' -o -name '*.test' -o -name 'build-errors.log' \) -path './live-service/*' -delete 2>/dev/null || true
	find . -type f \( -name 'core' -o -name 'a.out' -o -name '*.exe' -o -name '*.test' -o -name 'build-errors.log' \) -path './studio-service/*' -delete 2>/dev/null || true
	find . -type f \( -name 'core' -o -name 'a.out' -o -name '*.exe' -o -name '*.test' -o -name 'build-errors.log' \) -path './work-service/*' -delete 2>/dev/null || true
	find . -type f \( -name 'core' -o -name 'a.out' -o -name '*.exe' -o -name '*.test' -o -name 'build-errors.log' \) -path './bili-proxy-service/*' -delete 2>/dev/null || true
	find . -type f \( -name 'core' -o -name 'a.out' -o -name '*.exe' -o -name '*.test' -o -name 'build-errors.log' \) -path './transcoder-service/*' -delete 2>/dev/null || true
	@echo "✓ 源码目录 in-source build 残留已清理"