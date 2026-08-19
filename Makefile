# Go 面试复习项目 Makefile
#
# 提供 GOCACHE 重定向，避免在沙箱/受限环境下构建失败。
# 同时支持多个 topic（gmp / goroutine / ...），按需切换。
#
# 用法:
#   make run TOPIC=gmp DEMO=preemptive       # 跑 gmp 下的 preemptive demo
#   make test TOPIC=goroutine                # 跑 goroutine 测试
#   make bench TOPIC=gmp                     # 跑 gmp benchmark
#   make trace TOPIC=gmp DEMO=preemptive     # 抓调度 trace
#   make help                      # 查看所有命令

GOCACHE_DIR := .gocache
GOCACHE_ENV := GOCACHE=$(CURDIR)/$(GOCACHE_DIR)

# 默认 topic
TOPIC ?= gmp

.PHONY: help
help:
	@echo "Go 面试复习项目"
	@echo ""
	@echo "用法:"
	@echo "  make run TOPIC=<topic> [DEMO=<demo>]   # 跑 demo"
	@echo "  make test TOPIC=<topic>                # 跑测试"
	@echo "  make bench TOPIC=<topic>               # 跑 benchmark"
	@echo "  make trace TOPIC=<topic> [DEMO=<demo>] # 跑 demo + 抓调度 trace"
	@echo "  make list                              # 列出所有 topic"
	@echo ""
	@echo "当前默认 TOPIC = $(TOPIC)"
	@echo ""
	@echo "可用 topic:"
	@echo "  gmp         GMP 调度模型"
	@echo "  goroutine   Goroutine 演示"

.PHONY: list
list:
	@echo "已创建 topic:"
	@ls -d code/01-basics/*/ 2>/dev/null | sed 's|code/01-basics/||;s|/||' | sort

.PHONY: ensure-topic
ensure-topic:
	@if [ ! -d code/01-basics/$(TOPIC) ]; then \
		echo "❌ Topic 不存在: $(TOPIC)"; \
		echo "可用 topic:"; \
		ls -d code/01-basics/*/ 2>/dev/null | sed 's|code/01-basics/||;s|/||' | sort; \
		exit 1; \
	fi

.PHONY: run
run: ensure-gocache ensure-topic
	cd code/01-basics/$(TOPIC) && $(GOCACHE_ENV) go run . $(DEMO)

.PHONY: test
test: ensure-gocache ensure-topic
	cd code/01-basics/$(TOPIC) && $(GOCACHE_ENV) go test -v ./...

.PHONY: bench
bench: ensure-gocache ensure-topic
	cd code/01-basics/$(TOPIC) && $(GOCACHE_ENV) go test -bench=. -benchtime=1s -run=^$$ ./...

.PHONY: trace
trace: ensure-gocache ensure-topic
	cd code/01-basics/$(TOPIC) && $(GOCACHE_ENV) GODEBUG=schedtrace=200,scheddetail=1 go run . $(DEMO)

.PHONY: ensure-gocache
ensure-gocache:
	@mkdir -p $(GOCACHE_DIR)

.PHONY: clean
clean:
	rm -rf $(GOCACHE_DIR)

.PHONY: tidy
tidy:
	go mod tidy
