BIN := /tmp
GO  := go

.PHONY: run build clean build-core build-media

run:
	$(GO) build -o $(BIN)/mybilibili-core ./core-service/cmd/core
	$(BIN)/mybilibili-core

build: build-core build-media

build-core:
	$(GO) build -o $(BIN)/mybilibili-core ./core-service/cmd/core

build-media:
	$(GO) build -o $(BIN)/mybilibili-media ./media-service/cmd/media

clean:
	rm -f $(BIN)/mybilibili-core $(BIN)/mybilibili-media