# slowlog-ai

MySQL 慢日志 **V6 Agent** 分析（Go + DeepSeek + RAG + MCP）。**V1–V6 同仓保留**，便于对比演进。细节见 [docs/AGENT-ROADMAP.md](docs/AGENT-ROADMAP.md)。

---

## 命令速查（先看这张表）

| 命令 | 一步说明 | 需要 API Key | 需要 MySQL |
|------|----------|:------------:|:----------:|
| `make run` | 控制台跑一遍 **V6 Agent**（默认慢日志） | ✓ | 可选 |
| `make agent-run` | **完整跑通** + 写入 `reports/`（HTML/MD/JSON，可复盘） | ✓ | 推荐 |
| `make agent-eval` | **回归测试**：标准用例检查轨迹与结论，**不调 LLM** | — | — |
| `make mysql-check` | 只测 `.env` 里 MySQL 能否连通 | — | ✓ |
| `make report-md JSON=reports/xxx.json` | 从已有 JSON **重生** MD/HTML，不重跑 Agent | — | — |
| `make doc-links` | 校验文档内 Markdown 链接与锚点 | — | — |
| `SLOWLOG_AGENT_TRACE=1 make run` | 同上，但 **stderr 打印每轮** 决策与工具/RAG | ✓ | 可选 |

```bash
cp .env.example .env    # DEEPSEEK_API_KEY；可选 MYSQL_*=test
make agent-eval         # 改代码后先跑这个（秒级、无 Token）
make agent-run          # 真 LLM + 报告存档（推荐演示/复盘）
make run                # 轻量演示
```

| 文档 | 何时打开 |
|------|----------|
| [AGENT-ROADMAP](docs/AGENT-ROADMAP.md) | 总览、后续路线、汇报提纲 |
| [AGENT-RUN](docs/AGENT-RUN.md) | `agent-run` 报告里有什么、参数说明 |
| [AGENT-EVAL](docs/AGENT-EVAL.md) | 回归测什么、如何对比 `agent-run` |
| [VERSIONS](docs/VERSIONS.md) | V1–V6 设计与代码示例 |
| [ARCHITECTURE](docs/ARCHITECTURE.md) | 接口、MCP、扩展开发 |

---

## V6 Agent 执行流程

```text
拼 Prompt → LLM 输出 NextAction → 执行 → 结果写入 context → 下一轮 … → finish 结束
```

| `NextAction.type` | 本轮做什么 |
|-------------------|------------|
| `retrieve_rag` | 查知识库，结果进 `context` |
| `call_tool` | 调 MCP（慢日志分析 / 连库 / EXPLAIN / 建索引 dry_run） |
| `analyze` | 把中间结论写进 `context` |
| `ask_question` | 记录待问（演示模式不暂停等人） |
| `finish` | 输出最终 `result`，**结束** |

流程图：[docs/diagrams/v6-agent-flow.md](docs/diagrams/v6-agent-flow.md)

**`make agent-run` 推荐顺序（guided）**：RAG → 连库 → EXPLAIN → 索引 dry_run → analyze → finish → 打开 `reports/*.brief.html` 看逐轮。

---

## 版本演进速查

> 实现细节：[VERSIONS](docs/VERSIONS.md) · V5/V6 区别见下节

### 总览表（V1 → V6）

| 版本 | 一句话 |
|------|--------|
| V1 | 直接问模型 |
| V2 | JSON + confirmed/suspected |
| V3 | + RAG |
| V4 | 能力注册表 / MCP |
| V5 | API `tool_calls` |
| **V6** | **NextAction 多轮 Agent（默认）** |

### V5 与 V6 {#v5-vs-v6}

| | V5 | V6 |
|---|----|----|
| 协议 | DeepSeek Tool Calling | 自描述 `NextAction` JSON |
| 能做 |  mainly 调工具 | 工具 + RAG + 分析 + 提问 + 结束 |
| 入口 | `main.go` 注释块 | **`make run` 默认** |

切换 V4/V5：编辑 `cmd/slowlog-ai/main.go` 对应注释块。

---

## MCP 能力

| 工具 | 作用 |
|------|------|
| `analyze_slow_log` | 慢日志结构化分析 |
| `connect_mysql_instance` | 校验 MySQL |
| `explain_mysql_query` | SELECT 的 EXPLAIN |
| `add_mysql_index` | 建索引 DDL（默认 **dry_run**） |

---

## 目录

```text
cmd/slowlog-ai/     # 默认 V6 演示
cmd/agent-run/      # 全流程 + reports/
cmd/agent-eval/     # 回归（标准用例）
internal/analyzer/  # V1–V6，含 v6_agent、报告
internal/eval/      # 回归逻辑
internal/mcp/ · mysql/ · rag/ · llm/
docs/               # ROADMAP、RUN、EVAL、VERSIONS、ARCHITECTURE
testdata/           # slowlog-products.txt 等
```

---

## 环境

Go 1.23+ · `DEEPSEEK_API_KEY` · 本地演示建议 `MYSQL_DATABASE=test` 与 `testdata/slowlog-products.txt` 中表结构一致。

```bash
make deps          # go mod tidy（内网可设 GOPROXY，见 Makefile）
make build         # 编译 bin/
```

---

## 贡献与许可

欢迎 Issue / PR。许可证：[待定]
