# external 服务统一 Makefile
# 用法:
#   make build                    # 编译全部
#   make build-whisper-local      # 编译单个
#   BIN=/usr/local/bin make build # 指定输出目录

BIN      ?= ../../bin
GO       := go
LDFLAGS  := -ldflags="-s -w"

.PHONY: build clean \
	build-whisper-local build-whisper \
	build-transcoder-cpu build-transcoder-nvenc build-transcoder-vaapi \
	build-embedding-llamacpp-vulkan

build: build-whisper-local build-whisper \
	build-transcoder-cpu build-transcoder-nvenc build-transcoder-vaapi \
	build-embedding-llamacpp-vulkan

# ---------- whisper ----------

build-whisper-local:
	$(GO) build $(LDFLAGS) -o $(BIN)/whisper-local ./whisper-local/cmd/whisper-local

build-whisper:
	$(GO) build $(LDFLAGS) -o $(BIN)/whisper ./whisper/cmd/whisper

# ---------- transcoder ----------
# cpu 版默认（soft），nvenc/vaapi 需 build tag

build-transcoder-cpu:
	$(GO) build $(LDFLAGS) -o $(BIN)/transcoder-cpu ./transcoder-cpu/cmd/transcoder-cpu

build-transcoder-nvenc:
	$(GO) build $(LDFLAGS) -tags nvenc -o $(BIN)/transcoder-nvenc ./transcoder-nvenc/cmd/transcoder-nvenc

build-transcoder-vaapi:
	$(GO) build $(LDFLAGS) -tags vaapi -o $(BIN)/transcoder-vaapi ./transcoder-vaapi/cmd/transcoder-vaapi

# ---------- embedding ----------

build-embedding-llamacpp-vulkan:
	$(GO) build $(LDFLAGS) -o $(BIN)/embedding-llamacpp-vulkan ./embedding-llamacpp-vulkan/cmd/embedding-llamacpp-vulkan

# ---------- clean ----------

clean:
	rm -f $(BIN)/whisper-local $(BIN)/whisper \
		$(BIN)/transcoder-cpu $(BIN)/transcoder-nvenc $(BIN)/transcoder-vaapi \
		$(BIN)/embedding-llamacpp-vulkan
