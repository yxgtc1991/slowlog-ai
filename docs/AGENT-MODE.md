# V5 / V6 模式怎么切换

同一套 MCP 能力，两种 **LLM 协议**：

| 模式 | 协议 | 典型能力 |
|------|------|----------|
| **v6**（默认） | 自描述 `NextAction` JSON | RAG、工具、analyze、finish 多轮 |
| **v5** | API **Tool Calling**（`tool_calls`） | 主要靠模型返回工具调用，再汇总 |

不必改代码注释块，用 **环境变量** 或 **命令行** 即可切换。

---

## 1. 环境变量（推荐写进 `.env`）

```bash
# 默认不写 = v6
SLOWLOG_AGENT_MODE=v6

# 走 V5 Tool Calling
SLOWLOG_AGENT_MODE=v5
```

与 RAG 独立：`SLOWLOG_RAG` 仍控制 tfidf / embedding / mock（见 [RAG.md](RAG.md)）。

---

## 2. 命令行（覆盖 `.env`）

```bash
# 控制台演示
go run ./cmd/slowlog-ai -agent-mode=v5
go run ./cmd/slowlog-ai -agent-mode=v6

# 等价简写
go run ./cmd/slowlog-ai -mode=v5

# 带慢日志路径（模式参数可放任意位置）
go run ./cmd/slowlog-ai -agent-mode=v5 testdata/slowlog-products.txt
```

Makefile：

```bash
make run-v5          # V5 Tool Calling 演示
make run             # 默认 V6
```

---

## 3. agent-run 写报告

```bash
# 默认 V6：reports/agent-run-*.json / brief.html 等
make agent-run

# V5：reports/v5-run-*.json + v5-run-*.md（无 brief.html）
make agent-run-v5
# 或
go run ./cmd/agent-run -agent-mode=v5
```

V5 **不支持** V6 的 `guided` 流程与逐轮 brief HTML；适合对比「纯 Tool Calling」路径。

---

## 4. 何时用哪个

| 场景 | 建议 |
|------|------|
| 演示完整 Agent、客户复盘 | **v6** + `make agent-run` |
| 回归、无 API | **v6** + `make agent-eval`（固定 Mock RAG） |
| 对比业界 Tool Calling 协议 | **v5** + `make run-v5` |
| 汇报演进故事 | V4 能力 → **V5 tool_calls** → **V6 NextAction**（本仓同存） |

代码入口：

- V5：`internal/analyzer/v5_tool_calling.go`、`internal/llm/deepseek_tool_adapter.go`
- V6：`internal/analyzer/v6_agent.go`
- 模式解析：`internal/agentmode/mode.go`
