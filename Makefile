# Go 面试复习项目 Makefile
#
# 提供 GOCACHE 重定向，避免在沙箱/受限环境下构建失败。
# 正常使用也可以加速构建（缓存就在项目里）。
#
# 用法:
#   make run GMP=gmp preemptive    # 跑 gmp 下的 preemptive demo
#   make test GMP=gmp              # 跑测试
#   make bench GMP=gmp             # 跑 benchmark
#   make trace GMP=gmp preemptive  # 跑 demo + 抓调度 trace

GOCACHE_DIR := .gocache
GOCACHE_ENV := GOCACHE=$(CURDIR)/$(GOCACHE_DIR)

# 默认目标：列出所有可用的 GMP demo
.PHONY: help
help:
	@echo "Go 面试复习项目"
	@echo ""
	@echo "用法:"
	@echo "  make run DEMO=preemptive     # 跑 demo（GMP=gmp 默认）"
	@echo "  make test                    # 跑测试"
	@echo "  make bench                   # 跑 benchmark"
	@echo "  make trace DEMO=preemptive   # 跑 demo 并抓调度 trace"
	@echo ""
	@echo "可选参数:"
	@echo "  DEMO=<name>    demo 名称（preemptive / work-steal / gomaxprocs / count）"
	@echo ""
	@echo "示例:"
	@echo "  make run DEMO=preemptive"
	@echo "  make trace DEMO=work-steal"

.PHONY: run
run: ensure-gocache
	cd code/01-basics/gmp && $(GOCACHE_ENV) go run . $(DEMO)

.PHONY: test
test: ensure-gocache
	cd code/01-basics/gmp && $(GOCACHE_ENV) go test -v ./...

.PHONY: bench
bench: ensure-gocache
	cd code/01-basics/gmp && $(GOCACHE_ENV) go test -bench=. -benchtime=1s -run=^$$ ./...

.PHONY: trace
trace: ensure-gocache
	cd code/01-basics/gmp && $(GOCACHE_ENV) GODEBUG=schedtrace=200,scheddetail=1 go run . $(DEMO)

.PHONY: ensure-gocache
ensure-gocache:
	@mkdir -p $(GOCACHE_DIR)

.PHONY: clean
clean:
	rm -rf $(GOCACHE_DIR)

.PHONY: tidy
tidy:
	go mod tidy
