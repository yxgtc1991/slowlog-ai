# 内网 GOPROXY 常缺 go-sql-driver，默认用 goproxy.cn；可覆盖：make deps GOPROXY=...
GOPROXY ?= https://goproxy.cn,direct

export GOPROXY

.PHONY: deps vendor build mysql-check mysql-cap-check rag-check rag-check-compare rag-test test check real-run-samples doc-diagrams doc-links agent-api api-test agent-run agent-run-v5 agent-eval report-md run run-v5

deps:
	go mod tidy

vendor: deps
	go mod vendor

build: vendor
	go build -o bin/mysql-check ./cmd/mysql-check
	go build -o bin/slowlog-ai ./cmd/slowlog-ai

mysql-check:
	go run ./cmd/mysql-check

mysql-cap-check:
	go run ./cmd/mysql-cap-check

rag-check:
	go run ./cmd/rag-check

rag-check-compare:
	go run ./cmd/rag-check -compare

rag-test:
	go test ./internal/rag/... -count=1

test:
	go test ./internal/... -count=1

check: test agent-eval rag-test doc-links
	@echo "check: all green"

# 真实 LLM 抽检（需 DEEPSEEK_API_KEY；报告写入 reports/，不提交 Git）
agent-api:
	go run ./cmd/agent-api

api-test:
	go test ./cmd/agent-api/... ./internal/service/... -count=1

real-run-samples:
	$(MAKE) mysql-check
	go run ./cmd/agent-run -guided=true -trace=false testdata/slowlog-products.txt
	go run ./cmd/agent-run -guided=false -trace=false testdata/slowlog-lock-wait.txt
	go run ./cmd/agent-run -guided=false -trace=false testdata/slowlog-join-large.txt
	go run ./cmd/agent-run -guided=false -trace=false testdata/slowlog-index-hit.txt
	@echo "见 docs/agent/REAL-RUN-CHECKLIST.md 核对报告"

doc-diagrams:
	./scripts/generate-diagrams.sh

doc-links:
	go run ./scripts/validate-md-links

agent-run:
	go run ./cmd/agent-run

agent-run-v5:
	SLOWLOG_AGENT_MODE=v5 go run ./cmd/agent-run

agent-eval:
	go run ./cmd/agent-eval -v

# 从已有 JSON 重新生成 GoLand 可读的 .md（无需重跑 Agent）
report-md:
	@test -n "$(JSON)" || (echo 'usage: make report-md JSON=reports/agent-run-xxx.json' && exit 1)
	go run ./cmd/report-md $(JSON)

run:
	go run ./cmd/slowlog-ai

run-v5:
	SLOWLOG_AGENT_MODE=v5 go run ./cmd/slowlog-ai
