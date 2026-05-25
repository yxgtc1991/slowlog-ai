# 本地开发速查

[测试指南](TESTING.md) · [跑通 Agent](RUN.md) · [文档索引](../INDEX.md)

---

## 环境

```bash
cp .env.example .env
# 必填：DEEPSEEK_API_KEY
# 可选：MYSQL_*（MCP 连库 / EXPLAIN）
```

---

## 日常循环

```bash
# 1. 改代码
# 2. 快速护栏（无 Token）
make check

# 3. 仅改 Agent 决策
make agent-eval

# 4. 仅改 RAG / 知识库
make rag-test
make rag-check "price 最左前缀"

# 5. 改 Prompt 或大逻辑后再跑真 LLM
make agent-run
# 打开 reports/*brief.html
```

---

## 常用环境变量

| 变量 | 作用 |
|------|------|
| `SLOWLOG_AGENT_MODE` | `v5` / `v6`（见 [MODE.md](MODE.md)） |
| `SLOWLOG_AGENT_TRACE=1` | stderr 打印每轮 NextAction |
| `SLOWLOG_AGENT_HITL=1` | `ask_question` 时暂停，等待终端输入 |
| `SLOWLOG_RAG` | `tfidf` / `embedding` / `mock` |
| `SLOWLOG_RAG_TOPK` | 检索条数（默认 3） |

---

## 代码入口

| 目标 | 文件 |
|------|------|
| V6 主循环 | `internal/analyzer/v6_agent.go` |
| NextAction 解析 | `internal/prompt/slowlog/v6_action.go` |
| MCP 工具 | `internal/mcp/` |
| RAG | `internal/rag/`、`slowlog/docs/` |
| Golden | `internal/eval/cases.go` |

---

## 新增 MCP 工具（简版）

1. 在 `internal/mcp/` 实现 handler 并注册到 Server  
2. 在 V6 Prompt / 能力列表中补充描述（见 [ARCHITECTURE](../design/ARCHITECTURE.md)）  
3. 可选：在 `eval` 的 `StubExecutor` 加 `WithTool` + 新 golden case  
4. `make agent-eval` + `make test`

---

## 新增知识库条目

1. 在 `internal/rag/slowlog/docs/<category>/` 新建 `.md`，**至少 2 个 `##` 标题**  
2. `make rag-test` 增加或调整 `golden_retrieval_test.go`  
3. `make rag-check "你的查询"` 肉眼确认 TopK
