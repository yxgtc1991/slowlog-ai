# AI 应用方向 · 项目讲解与能力说明

> 基于本仓库 **slowlog-ai**（V6 Agent + MCP + RAG）。用于对外交流、技术分享或岗位能力对照时的口述参考。  
> 演示路径：`make agent-eval` → `make agent-run` → 打开 `reports/*.brief.html`

---

## 一、2 分钟项目介绍（口述稿）

我这边用 Go 做了一套 **MySQL 慢日志智能分析** 的完整链路，核心不是单次调大模型，而是一个可落地的 **Agent 应用**：

第一层是 **多轮决策**。V6 里模型每轮输出结构化的 **NextAction**（例如先 `retrieve_rag` 查知识库，再 `call_tool` 连库做 EXPLAIN，最后 `analyze` / `finish`），运行时维护 **AgentState 阶段机**，下一轮 Prompt 只带状态摘要，避免把整段工具 JSON 反复灌给模型，控制 Token 和噪声。

第二层是 **工具与领域能力**。通过 **MCP** 把「连实例、EXPLAIN、索引 dry_run」等封装成统一工具；失败走 **toolerr** 映射成 `code` 和 `retryable`，Agent 和报告里都能看见，方便判断要不要重试。

第三层是 **RAG**。知识库在 `slowlog/docs`，按章节切块，默认 **TF-IDF TopK**；和 V6 的 `rag_query` 联动，检索结果先进 AgentState 再进后续 Prompt。同一套 `Retriever` 接口也接了内存向量 PoC，后续可换 Embedding API 或向量库，Agent 协议不用改。

工程上我补了 **golden 回归**（`make agent-eval`，不调 LLM API）、**结构化 Trace**（LLM/工具耗时进 JSON 和 brief 报告）、以及 **V5 Tool Calling 与 V6 并列切换**，方便在同一条 MCP 上对比「API tool_calls」和「自描述 NextAction」两种形态。

整体上，这是一个 **领域问题 + Agent 编排 + 可观测 + 可回归** 的 AI 应用样例，而不是 Demo 级的一次性 Prompt。

---

## 二、能力对照（AI 应用开发常见关注点）

| 能力点 | 本项目的体现 | 可指向的材料 |
|--------|--------------|--------------|
| Agent 编排 | V6 多轮 NextAction、guided 推荐顺序 | `v6_agent.go`、`MODE.md` |
| 工具集成 | MCP Server、Executor 适配 V5/V6 | `internal/mcp/` |
| Prompt / 上下文 | AgentState 摘要、阶段机 | `agent_state.go` |
| RAG | chunk、TF-IDF / embedding、`SLOWLOG_RAG` | [guides/RAG.md](../guides/RAG.md)、`internal/rag/` |
| 可靠性 | toolerr、连库失败仍可 finish | `toolerr/`、eval case |
| 可观测 | Trace span、brief 报告逐轮 | `trace.go`、`*.brief.html` |
| 可回归 | golden、轨迹断言 | [EVAL.md](EVAL.md) |
| 协议对比 | V5 tool_calls vs V6 NextAction | [MODE.md](MODE.md) |

---

## 三、三个深挖话题 · 标准答法

### 1. V6 NextAction 和 V5 Tool Calling，你为什么两种都保留？

**答：**  
慢日志分析既要 **调工具**（EXPLAIN、索引），也要 **查知识、写中间结论、结束任务**。V5 的 Tool Calling 是业界标准协议：模型返回 `tool_calls`，运行时执行再把结果喂回去，适合「工具链清晰」的场景。  
V6 用自描述的 **NextAction JSON**，除了 `call_tool`，还有 `retrieve_rag`、`analyze`、`finish` 等，一轮里行动类型更完整，状态和报告也更贴近「Agent 产品」的叙事。  
仓库里用同一套 MCP，通过 `SLOWLOG_AGENT_MODE` 切换，是为了 **在同域数据上公平对比**，而不是二选一重写。生产选型上，若团队已统一 Tool Calling 网关，可以只暴露工具类行动；若需要强流程和可读的逐轮报告，NextAction 更灵活。我现在的默认路径是 V6 + `agent-run` 报告。

---

### 2. RAG 在你们项目里具体解决什么？怎么验证不是摆设？

**答：**  
RAG 解决的是 **领域话术和边界**（例如 rows_examined 高不等于一定缺索引、LIMIT 不等于一定快），减少模型凭空编规则。  
实现上是：知识库 Markdown **按 `##` 切块**，启动建索引；V6 由模型给出 `rag_query`，`TFIDFRetriever` 做 TopK，结果写入 **AgentState** 再进入后续 Prompt，和「分析开头固定检索一次」的 V3 路径区分开。  
验证分两层：  
- **检索层**：`make rag-check` / `rag-check-compare`，不跑 LLM，只看 query 命中哪几条 chunk；  
- **Agent 层**：`make agent-eval` 用 Mock RAG 保证轨迹稳定，真 TF-IDF 在 `make run` / `agent-run` 里用。  
知识库已按 **products 慢日志场景** 扩到 11 篇 / 30+ chunk（左前缀、ORDER BY+LIMIT、EXPLAIN 等），`make rag-test` 做检索 golden；更大规模可换 embedding / 向量库，接口 `Retriever` 已隔离。

---

### 3. 你怎么保证改 Prompt / 改 Agent 逻辑之后系统还能用？

**答：**  
三块：  
1. **Eval**：`internal/eval` 里 golden case，用 `ScriptLLM` 固定模型输出，对 **轨迹**（例如先连库再 finish）和 **结论关键词** 断言，`make agent-eval` 秒级、无 API 费用。  
2. **报告复盘**：`agent-run` 落 JSON + brief HTML，含每轮 `llm_raw`、action、耗时 Trace，出问题不用重跑 Token 就能对齐是哪一轮决策错了。  
3. **工具契约**：MCP 返回统一错误结构，Agent 侧 `toolerr` 标准化，避免模型看到一堆不可解析的 stderr。  
若上 CI，就是把 `agent-eval` 和 `go test ./internal/rag/...` 挂在 PR 上；和「线上 A/B 看业务指标」是互补关系，我这套更偏 **工程回归**。

---

## 四、5 分钟演示脚本（建议顺序）

| 步骤 | 操作 | 讲解一句 |
|:----:|------|----------|
| 1 | `make agent-eval` | 改 Agent 相关代码先跑回归，确定性、无大模型调用 |
| 2 | `make agent-run` | 真实慢日志 + MySQL + 多轮 Agent，报告写入 `reports/` |
| 3 | 打开 `*brief.html` | 指 2～3 轮：RAG → 工具 → finish，强调可复盘 |
| 4（可选） | `make rag-check "rows_examined 索引"` | 检索与 Agent 解耦，可单独验证知识库 |
| 5（可选） | 提到 `SLOWLOG_AGENT_MODE=v5` | 同 MCP 下对比 Tool Calling 协议 |

不必展开 V5 全流程，除非对方专门问协议差异。

---

## 五、可主动提的「难点」（显得真做过）

任选其一展开 1～2 分钟即可：

1. **NextAction 解析与纠正**：模型把 `type` 和 `tool_name` 写乱时的 `normalizeNextAction`，避免整轮失败。  
2. **AgentState vs 报告全文**：Prompt 只带摘要，完整工具结果进 JSON 报告，平衡 Token 与可追溯。  
3. **EXPLAIN 与慢日志 SQL 对齐**：防止模型写成 `orders` 等错误表名，运行时改回慢日志里的 SELECT。  
4. **Trace**：`llm.chat` / `tool.*` span 写入 round 与 brief 表，方便看慢在哪一轮。

代码位置：`v6_agent.go`、`agent_state.go`、`slowlog_sql.go`（或项目内 EXPLAIN 对齐逻辑）、`trace.go`。

---

## 六、边界说明（诚实、加分）

交流时建议主动带一句「当前边界」：

- 知识库篇数少，RAG 是 **机制验证**，不是大规模检索产品。  
- Embedding 默认 **local** PoC；生产会接合规的 Embedding API 或向量库。  
- `ask_question` 目前是记录型，**尚未**做进程级暂停等人回复（若对方问 HITL，可说明在 roadmap）。  
- 鉴权、多租户、配额、线上监控对接——属于平台化，本仓库聚焦 **Agent 应用层**。

---

## 七、相关文档索引

| 文档 | 用途 |
|------|------|
| [docs/INDEX.md](../INDEX.md) | 文档总索引 |
| [ROADMAP.md](ROADMAP.md) | 总览、路线、对外提纲 |
| [RUN.md](RUN.md) | 报告字段、参数 |
| [EVAL.md](EVAL.md) | 回归范围 |
| [guides/RAG.md](../guides/RAG.md) | 检索模式与命令 |
| [MODE.md](MODE.md) | V5/V6 切换 |
| [design/VERSIONS.md](../design/VERSIONS.md) | V1–V6 设计说明 |

---

*可按目标团队 emphasis 微调第一节：偏平台则多讲 MCP/错误码/Trace；偏算法应用则多讲 RAG 与多轮 Prompt 摘要。*
