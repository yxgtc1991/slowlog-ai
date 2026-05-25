# Agent 学习与实现路线图

[文档索引](../INDEX.md) · [项目 README](../../README.md) · [版本详解](../design/VERSIONS.md) · [架构](../design/ARCHITECTURE.md) · [跑通](RUN.md)

> Agent 路线与命令的**正文**在本页；全库文档入口见 [docs/INDEX.md](../INDEX.md)。

---

## 文档去哪查

| 我想… | 打开 |
|--------|------|
| **文档总索引** | [docs/INDEX.md](../INDEX.md) |
| **总览 + 命令速查** | 本页 |
| 版本对比表、怎么切换 V4/V5/V6 | [README · 版本演进速查](../../README.md#版本演进速查) |
| V1–V6 设计思路、代码示例 | [design/VERSIONS.md](../design/VERSIONS.md) |
| 接口、MCP、MySQL、扩展 | [design/ARCHITECTURE.md](../design/ARCHITECTURE.md) |
| 跑 Agent 并生成报告 | [RUN.md](RUN.md) |
| Agent 回归（无 API） | [EVAL.md](EVAL.md) · [TESTING.md](TESTING.md) |
| 本地开发循环 | [DEVELOP.md](DEVELOP.md) |
| V6 流程图 | [diagrams/v6-agent-flow.md](../diagrams/v6-agent-flow.md) |
| RAG 流程 / V3 vs V6 | [diagrams/rag-flow.md](../diagrams/rag-flow.md) |
| RAG 命令与环境变量 | [guides/RAG.md](../guides/RAG.md) |
| AI 应用讲解稿 | [AI-APPLICATION-BRIEF.md](AI-APPLICATION-BRIEF.md) |
| Agent/RAG 调研问答（题单对照） | [RESEARCH-QA.md](RESEARCH-QA.md) |
| 生产化缺口与优化路线 | [PRODUCTION-GAPS.md](PRODUCTION-GAPS.md) |
| 校验文档内链 | `make doc-links` |

---

## 项目一句话

MySQL 慢日志 **多轮 Agent 分析**（Go + DeepSeek）：**RAG** 补专家知识、**MCP** 连真实库做 EXPLAIN / 索引建议、**V6** 自主规划步骤，**`make agent-run`** 可存档全程报告。

---

## 演进主线（V1 → V6）

> 详表见 [README · 总览表](../../README.md#总览表v1-v6)；实现细节见 [VERSIONS](../design/VERSIONS.md)。

```text
V1 问模型 → V2 JSON 约束 → V3 +RAG → V4 能力感知 → V5 Tool Calling → V6 Agent 循环
```

| 版本 | 一句话 | 关键文件 |
|------|--------|----------|
| V1 | 直接问 LLM | `prompt/slowlog/v1_basic.go` |
| V2 | 结构化 JSON，少编造 | `v2_strict.go` |
| V3 | RAG 注入（TF-IDF / embedding，eval 用 mock） | `v3_rag.go` · `rag/slowlog/docs/` |
| V4 | 能力注册表 / MCP 基础 | `v4_capability.go` · `mcp/*` |
| V5 | API 级 `tool_calls` | `v5_tool_calling.go` |
| **V6** | **NextAction 多轮（默认）** | `v6_agent.go` · `v6_action.go` |

**V5 vs V6**（易混）：[README · V5 与 V6](../../README.md#v5-vs-v6)

---

## 当前能力清单（已实现）

| 能力 | 说明 | 入口 |
|------|------|------|
| V6 Agent 循环 | `retrieve_rag` / `call_tool` / `analyze` / `ask_question` / `finish` | `make run` |
| MCP | 慢日志分析、连库、EXPLAIN、建索引 dry_run | [ARCHITECTURE · MCP](../design/ARCHITECTURE.md#mcp) |
| 轨迹 | stderr 每轮决策 | `SLOWLOG_AGENT_TRACE=1` … `-agent-trace` |
| 类型化状态 | `AgentState` 阶段机 + Prompt 摘要（控 Token） | `agent_state.go` |
| 报告存档 | JSON + 完整/精简 MD/HTML | `make agent-run` → [RUN](RUN.md) |
| 报告重生 | 从 JSON 再生成 MD/HTML | `make report-md JSON=reports/xxx.json` |
| LLM 容错 | 工具名误写 `type`、`finish` 的 object `result` | `v6_action.go` · `flex_string.go` |
| Agent 回归 | 5 条 golden，无 API | `make agent-eval` → [EVAL](EVAL.md) · [TESTING](TESTING.md) |
| 工具错误码 | `code` + `retryable` 写入状态与报告 | `toolerr/tool_error.go` |
| 结构化 Trace | `llm.chat` / `tool.*` span + 耗时进 JSON/brief | `trace.go` · `agent-run` 且记录 rounds 时 |
| 真 RAG | 16 篇知识库 + chunk + TF-IDF / embedding | [guides/RAG.md](../guides/RAG.md) · `make rag-test` |

---

## 后续路线（Agent 工程化）

按 **性价比** 排序；状态随开发更新。

| 优先级 | 目标 | 价值 | 状态 |
|:------:|------|------|:----:|
| P0 | **Agent Eval**（golden case、轨迹/结论断言） | 证明可回归、工程化 | **已实现** → [EVAL](EVAL.md) |
| P0 | **类型化 AgentState** + context 摘要进 Prompt | 状态机、控 Token | **已实现** |
| P0 | 工具统一错误码（`retryable` 等） | MCP / Agent 协作 | **已实现** |
| P1 | **结构化 Trace**（span、耗时写入报告 JSON） | 可观测 | **已实现** |
| P1 | **真 RAG**（chunk + TF-IDF TopK，替换 Mock） | V3 做实、query 相关 | **已实现** |
| P1 | **Tool Calling 模式**（与 V6 NextAction 并列） | 对齐业界协议 | **已实现** → [MODE](MODE.md) |
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
make run-v5                   # V5 Tool Calling 演示
make agent-run                # V6 全流程 + 报告（推荐复盘）
make agent-run-v5             # V5 + v5-run 报告
make report-md JSON=reports/agent-run-xxx.json
make mysql-check              # 仅测 MySQL
make doc-links                # 校验 md 链接
make check                    # test + agent-eval + rag-test + doc-links
make test                     # internal 包级单测
make agent-eval               # V6 golden 回归（无 API）
make rag-check                # RAG 试跑（见 RAG.md）
make rag-check-compare        # tfidf vs embedding

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
cmd/slowlog-ai/main.go     # 默认 V6；SLOWLOG_AGENT_MODE=v5 切 Tool Calling
cmd/agent-run/main.go      # 完整体验 + 写 reports/
internal/analyzer/v6_agent.go
internal/prompt/slowlog/v6_action.go
internal/mcp/              # MCP 工具实现
internal/rag/              # slowlog/docs + TF-IDF（eval 时 SLOWLOG_RAG=mock）
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
3. V5→V6：不只调工具，还有 analyze / finish；NextAction 与 Tool Calling 的 trade-off（`make run-v5` 可对比）。  
4. 工程化：**`make agent-eval`**（Agent golden）+ **`make rag-test`**（检索 golden）+ Trace / toolerr。  
5. 知识库：products 场景、左前缀 / ORDER BY+LIMIT；`make rag-check` 现场演示检索。  
6. 难点：`normalizeNextAction`、报告 MD 折行、GoLand 看 **PNG / brief.html**。  

### 可主动提的亮点

- 同一仓库保留 V1–V6，**对比学习**而非黑盒 Demo。  
- 真实 MySQL + EXPLAIN，索引默认 **dry_run**。  
- 报告体系：**不必重跑 LLM 即可复盘**。  
- CI：`test` + `agent-eval` + `rag-test` + `doc-links`（见 `.github/workflows/ci.yml`）；本地等价 `make check`。

---

## 复习顺序建议

**第一次通读（2～3 小时）**

1. [docs/INDEX](../INDEX.md) → 本页 → [RUN](RUN.md) 跑一遍  
2. [VERSIONS](../design/VERSIONS.md) 只看 V3、V5、V6 三节  
3. [ARCHITECTURE](../design/ARCHITECTURE.md) 的 analysis-flow、mcp  

**汇报 / 演示前（1 小时）**

1. [AI-APPLICATION-BRIEF](AI-APPLICATION-BRIEF.md) 口述 2 分钟稿  
2. `make agent-eval` + `make rag-test`（30 秒，证明可回归）  
3. `make agent-run` → 打开 `reports/*brief.html` 指 3 轮  
4. 可选：`make rag-check "price 最左前缀"`、`make run-v5` 对比协议  

**动手改 Agent 时**

1. `v6_agent.go` / `v6_action.go`  
2. `internal/mcp/` 新工具注册方式见 ARCHITECTURE  
3. 改完跑 `make agent-eval`、`make rag-test`；大改 Prompt 再 `make agent-run`  

---

## 变更记录（与 git 提交一一对应）

> 新 **Agent / MCP** commit 合并后补一行：`git log -1 --format='%ad %h %s' --date=short`。纯文档 commit 可不记。

### Agent 工程化

| 日期 | Commit | 说明 |
|------|--------|------|
| 2026-05-21 | （待提交） | **封存加厚**：`make check`；5 条 Agent eval；知识库 +2；TESTING/DEVELOP；CI + `make test` |
| 2026-05-25 | `980cf55` | **CI + 回归补强**：GitHub Actions；`rag_then_finish`；知识库 +3；ROADMAP/口述更新 |
| 2026-05-25 | `1ab991d` | **工业级 RAG**：知识库 + `golden_retrieval_test` + `make rag-test`；docs 分目录 INDEX |
| 2026-05-25 | `d189eab` | **V5/V6 并列**：`SLOWLOG_AGENT_MODE` + `make run-v5` / `agent-run-v5` |
| 2026-05-25 | `1d75e28` | **真 RAG**：TF-IDF + `slowlog/docs` embed |
| 2026-05-25 | `a73aae8` | **RAG chunk/向量**：按 `##` 切分 + embedding 内存 TopK |
| 2026-05-25 | `fecce6e` | **结构化 Trace**：`trace.spans`（`llm.chat` / `tool.*` + `duration_ms`）；brief/HTML 逐轮「耗时」列 |
| 2026-05-25 | `c6017c2` | **工具错误码**：`toolerr`（`code` / `retryable`）；失败写入 AgentState 摘要与 `action_outcome` |
| 2026-05-22 | `3030261` | **Agent 状态**：`AgentState` 阶段机 + Prompt 摘要；默认 `products` 慢日志；EXPLAIN 对齐慢日志 SQL |
| 2026-05-22 | `e3c48ed` | **Agent 回归**：golden 标准用例、`make agent-eval`、轨迹/结论断言 |
| 2026-05-22 | `97f47c5` | **完整体验**：`agent-run`、报告 JSON/MD/HTML/brief、guided 流程、轨迹与报告修复 |

### MCP 真连库（2026-05）

| 日期 | Commit | 说明 |
|------|--------|------|
| 2026-05-21 | `11261f4` | **MySQL 工具**：`connect_mysql_instance`、`explain_mysql_query`、`add_mysql_index`（默认 dry_run） |

### V6 Agent（2026-01）

| 日期 | Commit | 说明 |
|------|--------|------|
| 2026-01-16 | `d957542` | **V6**：`NextAction` 多轮循环（RAG / 工具 / analyze / finish），`v6_agent.go` |

### MCP 与能力层（2026-01）

| 日期 | Commit | 说明 |
|------|--------|------|
| 2026-01-16 | `cad265d` | **MCP 决策**：LLM 按能力描述选工具（V5 方向雏形） |
| 2026-01-15 | `8dcea79` | **能力感知**：Registry 列出能力 Meta，供 Prompt 注入 |
| 2026-01-14 | `25504a8` | **能力抽象**：`Capability` 接口与 MCP Server 骨架 |

### Prompt 与 RAG（2026-01）

| 日期 | Commit | 说明 |
|------|--------|------|
| 2026-01-13 | `e64b41e` | **V3 RAG**：知识库检索注入 Prompt |
| 2026-01-13 | `fd87f92` | **V2**：JSON 输出、confirmed/suspected，减少编造 |
| 2026-01-12 | `09571b0` | **V1**：最简慢日志 Prompt，直连 LLM |
