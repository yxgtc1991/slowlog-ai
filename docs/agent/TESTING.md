# 测试与回归指南

[文档索引](../INDEX.md) · [Agent Eval](EVAL.md) · [RAG](../guides/RAG.md)

本仓库用 **三层护栏** 保证改代码后可快速验证，无需每次手跑 `make agent-run`（耗 Token）。

---

## 三层结构

| 层级 | 命令 | 耗时 | 需要 API | 测什么 |
|------|------|------|:--------:|--------|
| 包级单测 | `make test` | ~2s | 否 | 解析、报告、分词、配置、toolerr 等 |
| RAG 回归 | `make rag-test` | ~1s | 否 | TF-IDF / 分块 / **10 条**检索 golden |
| Agent 回归 | `make agent-eval` | ~1s | 否 | **5 条** V6 轨迹 golden（ScriptLLM） |

**合并前推荐**（与 CI 一致）：

```bash
make check    # test + agent-eval + rag-test + doc-links
```

推送 `main` 后 GitHub Actions 自动跑 `agent-eval`、`rag-test`、`test`、`doc-links`（见 `.github/workflows/ci.yml`）。

---

## 包级单测（`make test`）

覆盖 `internal/` 下除薄封装外的核心包：

| 包 | 代表测试 |
|----|----------|
| `analyzer` | AgentState、Trace、报告 MD/HTML、慢日志 SQL 对齐 |
| `rag` | tokenize、chunk、`splitMarkdownBySections`、factory、golden 检索 |
| `eval` | `TestAllCases` 与报告断言 fixture |
| `prompt/slowlog` | `normalizeNextAction`、`FlexString`、`ParseAgentDecision` |
| `toolerr` | 错误码映射 |
| `agentmode` | V5/V6 模式解析 |
| `mysql` | EXPLAIN 结果解析 |
| `config` | `.env` 加载不覆盖已有环境变量 |

```bash
make test
go test ./internal/analyzer/... -run TestAgentState -v
```

---

## RAG 回归（`make rag-test`）

- 知识库：`internal/rag/slowlog/docs/**/*.md`（**16 篇**，按 `##` 切块）
- 改知识库或 `tokenize` 后 **必跑**
- 试查不跑 LLM：`make rag-check "你的查询词"`

---

## Agent 回归（`make agent-eval`）

| Case | 验证点 |
|------|--------|
| `guided_flow` | RAG → 三工具 → analyze → finish（6 轮） |
| `tool_name_as_type` | 工具名误写 `type` 仍能 `call_tool` |
| `rag_then_finish` | 仅 RAG + finish，结论含左前缀 |
| `tool_error_then_finish` | 连库失败仍有 toolerr 且能 finish |
| `analyze_then_finish` | 无工具，analyze → finish 仍可结论 |

```bash
make agent-eval
go run ./cmd/agent-eval -list
go run ./cmd/agent-eval -case=analyze_then_finish -v
```

对真实报告断言（需先有 `reports/*.json`）：

```bash
go run ./cmd/agent-eval -report=reports/agent-run-xxx.json
```

---

## 何时跑哪一层

| 你改了… | 至少跑 |
|---------|--------|
| `v6_action.go` / `v6_agent.go` | `make agent-eval` |
| `slowlog/docs` / `tokenize` / TF-IDF | `make rag-test` |
| 报告模板 / Trace | `make test` + 可选 `make agent-run` 肉眼 |
| Prompt 大改 / 模型行为 | `make agent-run` + `-report` 对比 |
| 仅文档 | `make doc-links` |

---

## 扩展 golden

- **Agent**：在 `internal/eval/cases.go` 的 `AllCases()` 增加 `Script` + `Expect`
- **RAG**：在 `golden_retrieval_test.go` 增加 `query` / `wantInTitle`
- **报告基线**：复制 `reports/xxx.json` → `testdata/eval/`，用 `AssertReportFile`
