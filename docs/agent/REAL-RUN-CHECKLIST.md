# 真实 LLM 跑通抽检清单（G09）

[生产化缺口](PRODUCTION-GAPS.md) · [RUN](RUN.md)

`make agent-eval` 保证**协议与轨迹**；改 Prompt 或换模型后，用本清单对 **真实 API** 做抽检（非每次 CI 必跑）。

---

## 前置

```bash
cp .env.example .env   # DEEPSEEK_API_KEY；可选 MYSQL_*
make mysql-check       # 若要走 EXPLAIN 全链路
```

---

## 四条慢日志各跑一遍（可选）

```bash
go run ./cmd/slowlog-ai testdata/slowlog-products.txt
go run ./cmd/slowlog-ai testdata/slowlog-lock-wait.txt
go run ./cmd/slowlog-ai testdata/slowlog-join-large.txt
go run ./cmd/slowlog-ai testdata/slowlog-index-hit.txt
```

或统一存档：

```bash
make agent-run
```

---

## 报告里必看项

| 检查项 | 通过标准 |
|--------|----------|
| 证据行 | `finish` / 最终结论含「证据：」与 Query_time、Rows_* 等 |
| 锁等待场景 | lock-wait 样例**不**盲目建议加扫描索引 |
| 索引命中场景 | index-hit 样例**不**建议重复加索引 |
| JOIN 场景 | join 样例提到驱动表 / JOIN 键（允许未连库时写推断） |
| EXPLAIN | products 全链路含 `explain_mysql_query`（有 MySQL 时） |
| DDL | 仅有 `dry_run` DDL，无未说明的 `dry_run=false` |

---

## 基线回归（可选）

对满意的一次 `reports/agent-run-*.json`：

```bash
go run ./cmd/agent-eval -report=reports/agent-run-YYYYMMDD-HHMMSS.json
```

---

## 记录

| 日期 | 模型 | 场景 | 问题 | 处理 |
|------|------|------|------|------|
| | | | | |
