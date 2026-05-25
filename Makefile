# 内网 GOPROXY 常缺 go-sql-driver，默认用 goproxy.cn；可覆盖：make deps GOPROXY=...
GOPROXY ?= https://goproxy.cn,direct

export GOPROXY

.PHONY: deps vendor build mysql-check mysql-cap-check rag-check rag-check-compare rag-test doc-diagrams doc-links agent-run agent-run-v5 agent-eval report-md run run-v5

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
