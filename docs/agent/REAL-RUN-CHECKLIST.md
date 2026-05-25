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

或一次跑齐四条（耗 API Token，约 3～5 分钟/条）：

```bash
make real-run-samples
```

单场景存档仍可用 `make agent-run`（默认 products）。

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

| 日期 | 模型 | 场景 | 结果 | 报告 |
|------|------|------|------|------|
| 2026-05-25 | DeepSeek（`.env`） | products（guided） | ✅ 证据+EXPLAIN+dry_run DDL；`-report` 断言 OK | `reports/agent-run-20260525-181138.*` |
| 2026-05-25 | DeepSeek | lock-wait | ✅ 锁等待分析，未盲目加扫描索引；含证据行 | `reports/agent-run-20260525-181237.*` |
| 2026-05-25 | DeepSeek | join-large | ✅ JOIN/驱动表/连接键索引建议；表不存在时标明仅慢日志 | `reports/agent-run-20260525-181318.*` |
| 2026-05-25 | DeepSeek | index-hit | ✅ 明确无需加索引；EXPLAIN range+Using index | `reports/agent-run-20260525-181401.*` |

**G09 结论**：四条场景均通过上表「报告里必看项」；products 可作 `-report` 基线。

复现命令：

```bash
make mysql-check
go run ./cmd/agent-run -guided=true testdata/slowlog-products.txt
go run ./cmd/agent-run -guided=false testdata/slowlog-lock-wait.txt
go run ./cmd/agent-run -guided=false testdata/slowlog-join-large.txt
go run ./cmd/agent-run -guided=false testdata/slowlog-index-hit.txt
go run ./cmd/agent-eval -report=reports/agent-run-20260525-181138.json
```
