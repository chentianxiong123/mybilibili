# 编译产物输出目录。默认 ./bin (Go 生态约定), 可被命令行覆盖:
#   BIN=/custom/path make build
#   BIN=$HOME/bin   make build-transcoder
# 任何带 -o 的 in-source build 都用 clean-source 兜底清理.
BIN      ?= ./bin
GO       := go
LDFLAGS  := -ldflags="-s -w"

# 准备 BIN 目录
$(shell mkdir -p $(BIN))

.PHONY: run build clean test test-all test-integration test-integration-only lint coverage coverage-html build-core build-ai build-search build-msg-danmaku build-work build-live build-studio build-bili build-transcoder build-transcoder-nvenc build-transcoder-vaapi

run:
	$(GO) build $(LDFLAGS) -o $(BIN)/mybilibili-core ./services/core/cmd/core
	$(BIN)/mybilibili-core

build: build-core build-ai build-search build-msg-danmaku build-work build-live build-studio build-bili build-transcoder

build-core:
	$(GO) build $(LDFLAGS) -o $(BIN)/mybilibili-core ./services/core/cmd/core

build-ai:
	$(GO) build $(LDFLAGS) -o $(BIN)/mybilibili-ai ./services/ai/cmd/ai

build-search:
	$(GO) build $(LDFLAGS) -o $(BIN)/mybilibili-search ./services/search/cmd/search

build-msg-danmaku:
	$(GO) build $(LDFLAGS) -o $(BIN)/mybilibili-msg-danmaku ./services/msg-danmaku/cmd/msg-danmaku

build-work:
	$(GO) build $(LDFLAGS) -o $(BIN)/mybilibili-work ./services/work/cmd/work

build-studio:
	$(GO) build $(LDFLAGS) -o $(BIN)/mybilibili-studio ./services/studio/cmd/studio

build-live:
	$(GO) build $(LDFLAGS) -o $(BIN)/mybilibili-live ./services/live/cmd/live

build-bili:
	$(GO) build $(LDFLAGS) -o $(BIN)/mybilibili-bili ./services/bili-proxy/cmd/bili-proxy

# transcoder 一次编译三个版本（软编/VAAPI/NVENC），哪台机器用哪个
build-transcoder: build-transcoder-soft build-transcoder-vaapi build-transcoder-nvenc

build-transcoder-soft:
	$(GO) build $(LDFLAGS) -o $(BIN)/mybilibili-transcoder ./services/transcoder/cmd/transcoder

build-transcoder-vaapi:
	$(GO) build $(LDFLAGS) -tags vaapi -o $(BIN)/mybilibili-transcoder-vaapi ./services/transcoder/cmd/transcoder

build-transcoder-nvenc:
	$(GO) build $(LDFLAGS) -tags nvenc -o $(BIN)/mybilibili-transcoder-nvenc ./services/transcoder/cmd/transcoder

clean:
	rm -f $(BIN)/mybilibili-core $(BIN)/mybilibili-ai $(BIN)/mybilibili-search $(BIN)/mybilibili-msg-danmaku $(BIN)/mybilibili-work $(BIN)/mybilibili-studio $(BIN)/mybilibili-live $(BIN)/mybilibili-bili $(BIN)/mybilibili-transcoder

# ---------- 测试 ----------

.PHONY: test test-verbose test-cover test-race test-pkg test-core test-auth test-frontend test-integration test-all test-integration-only

# 单元测试：逐 Go 模块（services/** + shared）
test:
	@echo "== 单元测试 =="
	@for d in services/*/ shared/pkg ; do \
		if [ -f "$$d/go.mod" ]; then \
			echo "  → $$d"; \
			(cd "$$d" && go test -count=1 ./...) || exit 1; \
		fi; \
	done
	@echo "✓ 所有单元测试通过"

test-verbose:  ## 详细输出（逐模块）
	@for d in services/*/ shared/pkg ; do \
		[ -f "$$d/go.mod" ] && (cd "$$d" && go test -count=1 -v ./...); \
	done

test-cover:  ## 带覆盖率
	@for d in services/*/ shared/pkg ; do \
		if [ -f "$$d/go.mod" ]; then \
			echo "  → $$d"; \
			(cd "$$d" && go test -count=1 -cover ./...) || exit 1; \
		fi; \
	done

test-race:  ## 竞态检测
	@for d in services/*/ shared/pkg ; do \
		if [ -f "$$d/go.mod" ]; then \
			echo "  → $$d"; \
			(cd "$$d" && go test -count=1 -race ./...) || exit 1; \
		fi; \
	done

test-pkg:  ## 只测共享库 shared/pkg
	cd shared/pkg && go test -count=1 ./...

test-core:  ## 只测 core 服务
	cd services/core && go test -count=1 ./...

test-auth:  ## 只测 shared/pkg/auth
	cd shared/pkg && go test -count=1 ./auth/...

test-frontend:  ## 前端测试（vitest）
	cd web/user && npx vitest run
	cd web/admin && npx vitest run

test-integration: ## 集成测试（需要 docker）
	cd tests/integration && docker compose -f docker-compose.test.yml up -d --wait
	go test -tags=integration -count=1 -v ./tests/integration/...
	cd tests/integration && docker compose -f docker-compose.test.yml down -v

test-all: test test-frontend test-integration  ## 全部测试（Go单元+前端+集成）

test-integration-only:  ## 只跑集成测试
	go test -tags=integration -count=1 -v ./tests/integration/...

# 兜底清理: 扫描源码目录里所有产物, 防止某天 in-source build 污染.
# 用通配后缀, 不依赖服务名, 任何 *.exe / core / a.out 都会被识别.
clean-source:
	find . -type f \( -name 'core' -o -name 'a.out' -o -name '*.exe' -o -name '*.test' -o -name 'build-errors.log' \) -path './services/core/*' -delete 2>/dev/null || true
	find . -type f \( -name 'core' -o -name 'a.out' -o -name '*.exe' -o -name '*.test' -o -name 'build-errors.log' \) -path './services/ai/*' -delete 2>/dev/null || true
	find . -type f \( -name 'core' -o -name 'a.out' -o -name '*.exe' -o -name '*.test' -o -name 'build-errors.log' \) -path './services/search/*' -delete 2>/dev/null || true
	find . -type f \( -name 'core' -o -name 'a.out' -o -name '*.exe' -o -name '*.test' -o -name 'build-errors.log' \) -path './services/msg-danmaku/*' -delete 2>/dev/null || true
	find . -type f \( -name 'core' -o -name 'a.out' -o -name '*.exe' -o -name '*.test' -o -name 'build-errors.log' \) -path './services/live/*' -delete 2>/dev/null || true
	find . -type f \( -name 'core' -o -name 'a.out' -o -name '*.exe' -o -name '*.test' -o -name 'build-errors.log' \) -path './services/studio/*' -delete 2>/dev/null || true
	find . -type f \( -name 'core' -o -name 'a.out' -o -name '*.exe' -o -name '*.test' -o -name 'build-errors.log' \) -path './services/work/*' -delete 2>/dev/null || true
	find . -type f \( -name 'core' -o -name 'a.out' -o -name '*.exe' -o -name '*.test' -o -name 'build-errors.log' \) -path './services/bili-proxy/*' -delete 2>/dev/null || true
	find . -type f \( -name 'core' -o -name 'a.out' -o -name '*.exe' -o -name '*.test' -o -name 'build-errors.log' \) -path './services/transcoder/*' -delete 2>/dev/null || true
	@echo "✓ 源码目录 in-source build 残留已清理"

# ---------- Lint ----------

lint:  ## Go lint（需安装 golangci-lint）
	@for d in services/*/ shared/pkg ; do \
		if [ -f "$$d/go.mod" ]; then \
			echo "  → lint $$d"; \
			(cd "$$d" && golangci-lint run ./...) || true; \
		fi; \
	done

# ---------- 覆盖率 ----------

COVERAGE_DIR ?= /tmp/mybilibili-coverage

coverage:  ## 汇总各模块覆盖率到终端
	@echo "========== Go 模块覆盖率 =========="
	@for d in shared/pkg services/core services/search services/ai services/msg-danmaku services/live services/studio services/work services/bili-proxy; do \
		if [ -f "$$d/go.mod" ]; then \
			result=$$(cd "$$d" && go test -count=1 -coverprofile=$(COVERAGE_DIR)/cov_$$(echo $$d | tr '/' '_').out ./... 2>&1 | grep "^ok" | awk -F'coverage: ' '{print $$2}' | awk -F'%' '{s+=$$1; n++}END{if(n>0) printf "%.1f%%", s/n; else print "N/A"}'); \
			printf "  %-30s %s\n" "$$d" "$$result"; \
		fi; \
	done
	@echo "===================================="

coverage-html:  ## 生成各模块覆盖率 HTML 报告到 $(COVERAGE_DIR)/html/
	@mkdir -p $(COVERAGE_DIR)/html
	@for d in shared/pkg services/core services/search services/ai services/msg-danmaku services/live services/studio services/work services/bili-proxy; do \
		if [ -f "$$d/go.mod" ]; then \
			name=$$(echo $$d | tr '/' '_'); \
			cd "$$d" && go test -count=1 -coverprofile=$(COVERAGE_DIR)/$$name.out ./... 2>/dev/null; \
			go tool cover -html=$(COVERAGE_DIR)/$$name.out -o $(COVERAGE_DIR)/html/$$name.html 2>/dev/null; \
			cd $(OLDPWD); \
			echo "  → $(COVERAGE_DIR)/html/$$name.html"; \
		fi; \
	done
	@echo "✓ HTML 报告已生成到 $(COVERAGE_DIR)/html/"

coverage-check:  ## 门禁：总覆盖率不低于阈值（默认 70%）
	@THRESHOLD=$${THRESHOLD:-70}; \
	FAIL=0; \
	for d in shared/pkg services/core services/search services/ai services/msg-danmaku services/live services/studio services/work services/bili-proxy; do \
		if [ -f "$$d/go.mod" ]; then \
			pct=$$(cd "$$d" && go test -count=1 -coverprofile=$(COVERAGE_DIR)/cov_check.out ./... 2>&1 | grep "^ok" | awk -F'coverage: ' '{print $$2}' | awk -F'%' '{s+=$$1; n++}END{if(n>0) printf "%.0f", s/n; else print "0"}'); \
			if [ "$$pct" -lt "$$THRESHOLD" ] 2>/dev/null; then \
				echo "  ✗ $$d: $$pct% < $$THRESHOLD%"; \
				FAIL=1; \
			else \
				echo "  ✓ $$d: $$pct%"; \
			fi; \
		fi; \
	done; \
	if [ "$$FAIL" -eq 1 ]; then \
		echo "✗ 覆盖率低于阈值 $$THRESHOLD%，CI 失败"; \
		exit 1; \
	fi; \
	echo "✓ 所有模块覆盖率达标 (≥$$THRESHOLD%)"
