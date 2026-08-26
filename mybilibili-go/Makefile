BIN      := /tmp
GO       := go
LDFLAGS  := -ldflags="-s -w"

.PHONY: run build clean build-core build-ai build-search build-msg-danmaku build-work build-live build-studio build-bili

run:
	$(GO) build $(LDFLAGS) -o $(BIN)/mybilibili-core ./core-service/cmd/core
	$(BIN)/mybilibili-core

build: build-core build-ai build-search build-msg-danmaku build-work build-live build-studio build-bili

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

clean:
	rm -f $(BIN)/mybilibili-core $(BIN)/mybilibili-ai $(BIN)/mybilibili-search $(BIN)/mybilibili-msg-danmaku $(BIN)/mybilibili-work $(BIN)/mybilibili-studio $(BIN)/mybilibili-live $(BIN)/mybilibili-bili