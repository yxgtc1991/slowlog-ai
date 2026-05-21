# 内网 GOPROXY 常缺 go-sql-driver，默认用 goproxy.cn；可覆盖：make deps GOPROXY=...
GOPROXY ?= https://goproxy.cn,direct

export GOPROXY

.PHONY: deps vendor build mysql-check mysql-cap-check doc-links run

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

doc-links:
	go run ./scripts/validate-md-links

run:
	go run ./cmd/slowlog-ai
