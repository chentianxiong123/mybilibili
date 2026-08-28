BIN      := /tmp
GO       := go
LDFLAGS  := -ldflags="-s -w"

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