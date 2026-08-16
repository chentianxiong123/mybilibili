BIN      := /tmp
GO       := go
LDFLAGS  := -ldflags="-s -w"

.PHONY: run build clean build-core build-ai build-search build-msg-danmaku build-media build-live build-admin

run:
	$(GO) build $(LDFLAGS) -o $(BIN)/mybilibili-core ./core-service/cmd/core
	$(BIN)/mybilibili-core

build: build-core build-ai build-search build-msg-danmaku build-media build-live build-admin

build-core:
	$(GO) build $(LDFLAGS) -o $(BIN)/mybilibili-core ./core-service/cmd/core

build-ai:
	$(GO) build $(LDFLAGS) -o $(BIN)/mybilibili-ai ./ai-service/cmd/ai

build-search:
	$(GO) build $(LDFLAGS) -o $(BIN)/mybilibili-search ./search-service/cmd/search

build-msg-danmaku:
	$(GO) build $(LDFLAGS) -o $(BIN)/mybilibili-msg-danmaku ./msg-danmaku-service/cmd/msg-danmaku

build-media:
	$(GO) build $(LDFLAGS) -o $(BIN)/mybilibili-media ./media-service/cmd/media

build-live:
	$(GO) build $(LDFLAGS) -o $(BIN)/mybilibili-live ./live-service/cmd/live

build-admin:
	$(GO) build $(LDFLAGS) -o $(BIN)/mybilibili-admin ./admin-service/cmd/admin

clean:
	rm -f $(BIN)/mybilibili-core $(BIN)/mybilibili-ai $(BIN)/mybilibili-search $(BIN)/mybilibili-msg-danmaku $(BIN)/mybilibili-media $(BIN)/mybilibili-live $(BIN)/mybilibili-admin