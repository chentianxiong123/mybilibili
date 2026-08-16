BIN := /tmp
GO  := go

.PHONY: run build clean build-core build-ai build-search build-msg-danmaku build-media

run:
	$(GO) build -o $(BIN)/mybilibili-core ./core-service/cmd/core
	$(BIN)/mybilibili-core

build: build-core build-ai build-search build-msg-danmaku build-media

build-core:
	$(GO) build -o $(BIN)/mybilibili-core ./core-service/cmd/core

build-ai:
	$(GO) build -o $(BIN)/mybilibili-ai ./ai-service/cmd/ai

build-search:
	$(GO) build -o $(BIN)/mybilibili-search ./search-service/cmd/search

build-msg-danmaku:
	$(GO) build -o $(BIN)/mybilibili-msg-danmaku ./msg-danmaku-service/cmd/msg-danmaku

build-media:
	$(GO) build -o $(BIN)/mybilibili-media ./media-service/cmd/media

clean:
	rm -f $(BIN)/mybilibili-core $(BIN)/mybilibili-ai $(BIN)/mybilibili-search $(BIN)/mybilibili-msg-danmaku $(BIN)/mybilibili-media