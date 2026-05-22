# Agent 学习与实现路线图

← 建议 **从这里开始查阅** · [README](../README.md) · [版本详解](VERSIONS.md) · [架构与扩展](ARCHITECTURE.md) · [完整跑通](AGENT-RUN.md)

> 文档变多、实现变复杂时，用本页做**唯一入口**：先看清「做到哪了、下一步做什么、去哪份文档查细节」。

---

## 文档去哪查

| 我想… | 打开 |
|--------|------|
| **总览 + 命令速查** | 本页 |
| 版本对比表、怎么切换 V4/V5/V6 | [README · 版本演进速查](../README.md#版本演进速查) |
| V1–V6 设计思路、代码示例、学习要点 | [VERSIONS.md](VERSIONS.md) |
| 接口、Options、MCP、MySQL、扩展开发 | [ARCHITECTURE.md](ARCHITECTURE.md) |
| 跑一遍 Agent 并生成报告 | [AGENT-RUN.md](AGENT-RUN.md) |
| Agent 回归测试（无 API） | [AGENT-EVAL.md](AGENT-EVAL.md) |
| V6 流程图 | [diagrams/v6-agent-flow.md](diagrams/v6-agent-flow.md) |
| 校验文档内链 | `make doc-links` |

---

## 项目一句话

MySQL 慢日志 **多轮 Agent 分析**（Go + DeepSeek）：**RAG** 补专家知识、**MCP** 连真实库做 EXPLAIN / 索引建议、**V6** 自主规划步骤，**`make agent-run`** 可存档全程报告。

---

## 演进主线（V1 → V6）

> 详表见 [README · 总览表](../README.md#总览表v1-v6)；实现细节见 [VERSIONS](VERSIONS.md)。

```text
V1 问模型 → V2 JSON 约束 → V3 +RAG → V4 能力感知 → V5 Tool Calling → V6 Agent 循环
```

| 版本 | 一句话 | 关键文件 |
|------|--------|----------|
| V1 | 直接问 LLM | `prompt/slowlog/v1_basic.go` |
| V2 | 结构化 JSON，少编造 | `v2_strict.go` |
| V3 | RAG 注入（当前检索层 Mock） | `v3_rag.go` · `rag/slowlog/docs/` |
| V4 | 能力注册表 / MCP 基础 | `v4_capability.go` · `mcp/*` |
| V5 | API 级 `tool_calls` | `v5_tool_calling.go` |
| **V6** | **NextAction 多轮（默认）** | `v6_agent.go` · `v6_action.go` |

**V5 vs V6**（易混）：[README · V5 与 V6](../README.md#v5-vs-v6)

---

## 当前能力清单（已实现）

| 能力 | 说明 | 入口 |
|------|------|------|
| V6 Agent 循环 | `retrieve_rag` / `call_tool` / `analyze` / `ask_question` / `finish` | `make run` |
| MCP | 慢日志分析、连库、EXPLAIN、建索引 dry_run | [ARCHITECTURE · MCP](ARCHITECTURE.md#mcp) |
| 轨迹 | stderr 每轮决策 | `SLOWLOG_AGENT_TRACE=1` … `-agent-trace` |
| 报告存档 | JSON + 完整/精简 MD/HTML | `make agent-run` → [AGENT-RUN](AGENT-RUN.md) |
| 报告重生 | 从 JSON 再生成 MD/HTML | `make report-md JSON=reports/xxx.json` |
| LLM 容错 | 工具名误写 `type`、`finish` 的 object `result` | `v6_action.go` · `flex_string.go` |

---

## 后续路线（Agent 工程化）

按 **性价比** 排序；状态随开发更新。

| 优先级 | 目标 | 价值 | 状态 |
|:------:|------|------|:----:|
| P0 | **Agent Eval**（golden case、轨迹/结论断言） | 证明可回归、工程化 | **已实现** → [AGENT-EVAL](AGENT-EVAL.md) |
| P0 | **类型化 AgentState** + context 摘要进 Prompt | 状态机、控 Token | 计划 |
| P0 | 工具统一错误码（`retryable` 等） | MCP / Agent 协作 | 计划 |
| P1 | **结构化 Trace**（span、耗时写入报告 JSON） | 可观测 | 计划 |
| P1 | **真 RAG**（chunk + 向量/TF-IDF TopK，替换 Mock） | V3 做实、query 相关 | 计划 |
| P1 | **Tool Calling 模式**（与 V6 NextAction 并列） | 对齐业界协议 | 计划 |
| P2 | `ask_question` 真人机协同（暂停/恢复） | HITL | 计划 |
| P2 | Plan-and-Execute（先 plan 再执行） | 对比 ReAct | 计划 |

说明：

- **不必新增 V7**：横切能力挂在「V3 检索层」「V6 Agent 层」即可，演进故事仍用 V1–V6。
- **建议优先**：先做 P0（Eval + State），RAG 向量 1～2 天 PoC 即可挂在 P1。

---

## 命令速查

```bash
cp .env.example .env          # DEEPSEEK_API_KEY；可选 MYSQL_*=test

make run                      # 默认 V6 演示
make agent-run                # 全流程 + 报告（推荐复盘）
make report-md JSON=reports/agent-run-xxx.json
make mysql-check              # 仅测 MySQL
make doc-links                # 校验 md 链接
make agent-eval               # V6 golden 回归（无 API）

# 观察每轮决策（stderr）
SLOWLOG_AGENT_TRACE=1 go run ./cmd/slowlog-ai -agent-trace
```

报告阅读建议：

| 文件 | 用途 |
|------|------|
| `*.brief.html` | 给客户 / 自己一眼看每轮做了什么 |
| `*.html` / `*.md` | 完整复盘 |
| `*.json` | 机器读、`llm_raw`、重生报告 |

---

## 代码地图（常改哪里）

```text
cmd/slowlog-ai/main.go     # 默认 V6；注释块可开 V4/V5
cmd/agent-run/main.go      # 完整体验 + 写 reports/
internal/analyzer/v6_agent.go
internal/prompt/slowlog/v6_action.go
internal/mcp/              # MCP 工具实现
internal/rag/              # 知识库（当前 MockRetriever）
internal/analyzer/v6_report*.go
```

---

## 对外讲解提纲

### 5 分钟版

1. 背景：慢日志要「原因 + 建议」，不是只统计。  
2. 架构：V6 Agent 循环 + MCP（EXPLAIN、索引 dry_run）+ RAG。  
3. 演示：`make agent-run` → 打开 `*.brief.html` 看逐轮。  
4. 演进：V1→V6 一张表（README）。  

### 15 分钟版（加深）

1. V2→V3：confirmed / suspected 与 RAG 边界。  
2. V4→V5：能力发现 → API `tool_calls`。  
3. V5→V6：不只调工具，还有 analyze / finish；为何用 NextAction（灵活）以及和 Tool Calling 的 trade-off。  
4. 难点：`normalizeNextAction`、`FlexString`、报告 MD 与可观测。  
5. 下一步：Eval、真 RAG、双协议（见上表 P0/P1）。  

### 可主动提的亮点

- 同一仓库保留 V1–V6，**对比学习**而非黑盒 Demo。  
- 真实 MySQL + EXPLAIN，索引默认 **dry_run**。  
- 报告体系：**不必重跑 LLM 即可复盘**。

---

## 复习顺序建议

**第一次通读（2～3 小时）**

1. 本页 → README 速查表 → [AGENT-RUN](AGENT-RUN.md) 跑一遍  
2. [VERSIONS](VERSIONS.md) 只看 V3、V5、V6 三节  
3. [ARCHITECTURE](ARCHITECTURE.md) 的 analysis-flow、mcp  

**汇报 / 演示前（1 小时）**

1. README 总览表 + V5/V6 对比  
2. 本页「讲解提纲」+「后续路线 P0」  
3. 本地再跑一次 `make agent-run`，能口述每轮在干什么  

**动手改 Agent 时**

1. `v6_agent.go` / `v6_action.go`  
2. `internal/mcp/` 新工具注册方式见 ARCHITECTURE  
3. 改完跑 `make agent-eval` 与 `go test ./internal/...`；大改 Prompt 再 `make agent-run`  

---

## 变更记录（手动维护）

| 日期 | 里程碑 |
|------|--------|
| — | V1–V6 演进、MCP MySQL、文档拆分 |
| — | `agent-run`、多格式报告、GoLand 友好 MD |
| — | commit `97f47c5` Agent 完整体验与生成报告 |
