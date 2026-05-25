# Agent / RAG / 基础设施 — 技术调研问答

[文档索引](../INDEX.md) · [架构](../design/ARCHITECTURE.md) · [测试指南](TESTING.md)

> **文档性质**：把业界常见的 Agent、RAG、数据库与系统基础问题，整理成可查阅的 Q&A，并用 **slowlog-ai** 作对照样例。  
> 用途：工位自学、方案复盘、对外技术交流前的提纲；**不是**外部题单的原样摘录或背诵材料。

---

## 一、对照本题单：slowlog-ai 还可深化的方向

结合常见 **Agent 基建岗** 技术调研范围，本仓库**已覆盖**与**可继续加深**如下（按性价比排序）：

| 调研主题 | 本仓库现状 | 建议深化（仍属 P2，不挡封存演示） |
|----------|------------|----------------------------------|
| 上下文维护 / 压缩 | `AgentState.PromptSummary` 截断摘要；`conversationHistory` 仅保留决策短句 | 超长会话时：滑动窗口 + 按 phase 丢弃旧 RAG 标题；可选「摘要轮」单独调 LLM |
| Query 改写 / 多路召回 | **G16**：慢日志规则抽取 + 多 query + RRF（`SLOWLOG_RAG_MULTI`） | HyDE / LLM 改写；与规则路 A/B eval |
| 意图识别并行 | 单域（慢 SQL），意图隐含在 NextAction | 多域场景再拆「诊断 / 止血 / 变更」分类器 |
| Multi-Agent | **单 Agent** 多轮 ReAct 风格 | 拆「检索专员 + DBA 工具专员」需编排层，演示成本高 |
| Memory 长期记忆 | 仅当次 Run 的 `AgentState` + 报告 JSON | 向量库存历史 case；按 `schema+sql` 去重 |
| Reflection 自检 | 无单独反思轮 | `finish` 前增加 `reflect` action：核对结论是否含 EXPLAIN 证据 |
| HITL | **G10**：`SLOWLOG_AGENT_HITL=1` 时 stdin 阻塞；HTTP 路径仍不等待 | Webhook/工单系统暂停恢复 |
| RAG 分块 | 按 Markdown `##` 切分（可预期、可回归） | 对比 **递归字符切分** 做离线 eval，不必默认上线 |
| 向量库 / Milvus | 内存 embedding PoC | 分类检索用 **partition**（单 collection 多 partition）通常更易运维 |
| Skills 渐进披露 | **G08** `ToolsForPhase` 按阶段过滤工具列表 | 更细粒度 tool schema / 动态技能包 |
| 幻觉治理 | V2 confirmed/suspected、知识库边界文、RAG | 结论强制引用 `Rows_examined`/EXPLAIN 字段；报告里列 evidence |
| 代码分析 / 覆盖率 / Mock | **非本仓库范围** | 若岗位偏「代码 Agent」，需另项目；此处只保留 eval 的 `StubExecutor` 作工具 Mock 样例 |

**已足够支撑演示的**：MCP 真连库、V5/V6 双协议、`make check` 三层回归、结构化 Trace、toolerr、16 篇领域知识库。

---

## 二、Agent 架构与上下文

### Q1. Agent 的上下文怎么维护？常见技术方案有哪些？

**通用要点**

- **短期工作记忆**：当前任务的状态（阶段、工具结果摘要、检索片段），每轮拼进 Prompt。
- **对话历史**：完整多轮 messages，或只保留「决策摘要」以控 Token。
- **长期记忆**：跨会话的用户偏好、历史工单（常接向量库或 OLAP）。

**slowlog-ai 对照**

- 类型化 **`AgentState`**（phase 状态机 + RAG/工具/分析摘要），经 `PromptSummary(400)` 注入，**不把完整工具 JSON 反复灌给模型**（见 `internal/analyzer/agent_state.go`）。
- `conversationHistory` 只追加「决策类型 + reasoning」短句（`v6_agent.go`）。
- 全量轨迹在 **`reports/*.json`** 与 **Trace span** 存档，供复盘而非每轮全量进 Prompt。

---

### Q2. 主流大模型上下文窗口大概多长？

**通用要点**（随型号变化，以厂商文档为准）

- 常见商用 API：**32k～128k tokens**；部分开源/旗舰可达 **200k+**。
- 实际 Agent 仍应做摘要：窗口大 ≠ 成本低，且中间噪声会拉高幻觉率。

**slowlog-ai 对照**

- 通过 **状态摘要 + 截断**（如工具 summary 320 字、分析 800 字）主动控输入规模，而不是依赖超大窗口。

---

### Q3. Agent 架构一般分几层？

**通用分层（调研常用表述）**

| 层 | 职责 |
|----|------|
| 交互层 | CLI / API / IDE，收集输入、展示报告 |
| 编排层 | 多轮循环、状态机、超时与最大轮次 |
| 模型层 | LLM 推理、JSON 协议解析 |
| 工具层 | MCP / Function Calling、外部系统 |
| 知识层 | RAG、规则库、边界说明 |
| 观测层 | Trace、日志、eval golden |

**slowlog-ai 对照**：`cmd/agent-run` → `v6_agent.go` → `llm` → `mcp` + `rag` → `trace.go` / `make agent-eval`。

---

### Q4. Agent 和 LLM 的核心区别？

**答**：LLM 是**单轮概率生成**；Agent 是 **「感知 → 决策 → 行动 → 更新状态」的闭环**，可调用工具、查知识、在多轮中修正路径。  
本仓库 V1 是纯 LLM；V6 在相同模型上增加 **NextAction 协议 + 状态机 + MCP**，即 Agent 化。

---

### Q5. 单 Agent 还是多 Agent？子 Agent 分工怎么设计？

**通用要点**

- **单 Agent**：实现简单，适合域窄、工具 ≤10 个（慢 SQL 诊断属此类）。
- **多 Agent**：规划 / 执行 / _critique_ 分离，适合开放域或组织边界清晰（检索组、执行组、合规组）。

**slowlog-ai 对照**

- 采用 **单 Agent 多轮**；「分工」由 **phase hint** 与 guided 脚本体现（先 RAG → 连库 → EXPLAIN → 索引 dry_run → analyze → finish），而非进程级多 Agent。

---

### Q6. Prompt 模板怎么构造？

**通用要点**

- 系统角色 + 任务约束 + **输出 JSON Schema** + 当前状态 + 工具/能力列表 + 领域边界。
- 模板版本化（v1…v6 同仓），便于 A/B 与回归。

**slowlog-ai 对照**

- `internal/prompt/slowlog/v6_action.go` 的 `BuildAgentPromptV6`：慢日志正文、能力 Meta、历史决策、`AgentState` 摘要、可选 guided。
- V2 的 confirmed/suspected 约束仍在报告语义中体现。

---

### Q7. 上下文满了如何压缩？如何区分有效信息与噪声？

**通用要点**

- **截断**：按时间或相关性丢弃旧 tool 结果（本仓库已用）。
- **摘要**：用廉价模型或规则把多轮压成一段「已知事实」。
- **结构化状态**：只保留 phase、关键数字（Rows_examined）、DDL 一行。
- **噪声**：重复检索、失败且不可重试的工具调用、与当前表无关的 chunk。

**slowlog-ai 对照**

- 以 **结构化摘要为主**；未做 LLM 二次摘要。  
- **可优化**：`PhaseFinished` 后冻结状态；对重复 `retrieve_rag` 合并条目。

---

### Q8. 有没有在上下文里用 to-do / 清单保持模型聚焦？

**通用要点**

- 在 Prompt 中显式列出 **Remaining steps** 或 **Checklist**，每轮划掉已完成项（与 Cursor Plan、OpenAI 的 todo 工具同类思路）。

**slowlog-ai 对照**

- 用 **`AgentPhase.hint()`** 达到类似效果（「已检索知识库，可连库或 EXPLAIN」），比自由文本 todo 更省 Token、可测。  
- **可优化**：guided 模式下在 Prompt 顶部增加 3～5 项固定 checklist（与 eval `guided_flow` 对齐）。

---

### Q9. Agent 里的 Memory 指什么？如何实现？

**通用要点**

- **Working memory**：本轮 AgentState（本仓库）。
- **Episodic**：历史报告 JSON，可按 SQL 指纹检索相似 case。
- **Semantic**：知识库 RAG（`slowlog/docs`）。

**slowlog-ai 对照**：Working + Semantic 已有；Episodic 仅文件系统报告，**未建向量记忆库**。

---

### Q10. Reflection（反思）机制是什么？举例说明。

**通用要点**

- 在 `finish` 前增加一轮：**「根据证据自检结论是否矛盾」**；心理类 Agent 可检查语气是否专业。

**slowlog-ai 对照**

- 暂无独立 reflect action；**可优化**：要求 finish 的 `result` 必须引用 phase≥Explained 或明确写「未连库，仅慢日志推断」（eval `tool_error_then_finish` 已覆盖后者）。

---

## 三、RAG 与检索

### Q11. RAG 标准流程是什么？主流实现方式？

**答**

1. 文档加载 → 2. 切 chunk → 3. 建索引（稀疏/稠密）→ 4. 查询（可改写）→ 5. TopK 召回 → 6. 拼进 Prompt → 7. 生成 → 8.（可选）引用与校验。

**slowlog-ai 对照**

```text
embed slowlog/docs → splitMarkdownBySections(##) → TF-IDF 或 embedding 索引
→ V6 retrieve_rag(rag_query) → AgentState 记录 titles → 后续 Prompt
```

流程图：[diagrams/rag-flow.md](../diagrams/rag-flow.md)。

---

### Q12. 本项目 recall（召回）链路具体是怎样的？

**答（对照实现）**

| 步骤 | 实现 |
|------|------|
| Query 来源 | LLM 在 NextAction 里填 `rag_query`（非独立改写服务） |
| 分词 | `tokenize`：领域短语优先 + 英文词 + 汉字单字 |
| 召回 | `TFIDFRetriever` 余弦相似度 TopK（默认 3）或 `EmbeddingRetriever` |
| 注入 | `RecordRAG` → `PromptSummary` 只带 **标题列表**，不带全文 |
| 验证 | `make rag-test`（10 条 golden）、`make rag-check` |

---

### Q13. 向量检索和关键词检索分别适合什么场景？差异是什么？

| 维度 | 关键词（BM25/TF-IDF） | 向量（Embedding） |
|------|----------------------|-------------------|
| 强项 | 专有名词、指标名、错误码、英文标识符 |  paraphrase、口语化描述 |
| 弱项 | 同义改写、跨语言 | 稀有词、精确数字有时漂移 |
| 成本 | 低、可离线、易回归 | 依赖 embed 模型与索引更新 |
| 本仓库 | **默认 tfidf** | `SLOWLOG_RAG=embedding` PoC |

慢日志场景 **`Rows_examined`、`filesort`、`最左前缀`** 等词面信号强，故默认 TF-IDF；`make rag-check-compare` 可并排对比。

---

### Q14. 什么是「多路 Query 改写」？用户参与补充信息怎么设计？

**通用要点**

- **改写**：同一意图生成多条 query（同义、拆实体、HyDE 假设文档），多路召回后融合。
- **人机协同**：低置信时 Agent `ask_question`，把用户回答合并进下一轮 query（需 HITL 阻塞）。

**slowlog-ai 对照**

- 单路 `rag_query`；`ask_question` **未阻塞**等待用户。  
- **可优化**：对 `rag_query` 自动追加慢日志中的表名、指标名（规则抽取），相当于轻量改写。

---

### Q15. 什么是并行意图识别？为什么要并行？

**通用要点**

- 多意图（查文档 + 改配置 + 跑 SQL）同时分类，避免串行漏意图；并行多分类器或一次 JSON 输出多个 label。

**slowlog-ai 对照**

- 单意图域（慢 SQL 诊断），意图由 **NextAction.type** 表达即可；并行意图 **非刚需**。

---

### Q16. RAG 为什么有人用「递归字符切分」而不是固定长度？

**答**

- 固定长度易在句中断开，语义不完整。
- **递归切分**（先段落 `\n\n`，再句读，再字符）尽量保持局部语义，适合通用长文。

**slowlog-ai 对照**

- 领域文档是 **结构化 Markdown**，用 **`##` 标题切分** 边界清晰、chunk 可预期，利于 golden 回归（`chunks_section_test.go`）。  
- 若引入通用 PDF/日志，可再 **离线对比** 递归切分 vs `##` 的 Recall@K。

---

### Q17. Milvus 按类别查文档：多 Collection 还是 Partition？

**通用要点**（调研结论）

- **单 Collection + 多 Partition**（按 category/schema）：运维简单、共享 embed 模型、过滤 `partition in (...)`。
- **多 Collection**：模型维度不同、租户强隔离、生命周期完全独立时用。

**slowlog-ai 对照**

- 当前 **内存索引 + embed FS**，未上 Milvus；知识库用目录 `patterns/`、`metrics/` 等作逻辑分类，等价于「建索引前按目录过滤」。

---

## 四、工具调用、MCP 与 Skills

### Q18. Tool Calling / Function Calling 的底层原理？

**答**

1. 向模型注册 **工具 schema**（name、description、parameters JSON Schema）。  
2. 模型生成 **结构化输出**（`tool_calls` 或自描述 JSON）。  
3. 运行时 **校验参数 → 执行本地/远程函数 → 将 result 作为 tool message 塞回上下文**。  
4. 模型继续生成直至结束。

**slowlog-ai 对照**

- **V5**：API 级 `tool_calls`（`v5_tool_calling.go`）。  
- **V6**：`NextAction` 中 `call_tool` + `tool_name`/`tool_args`（`v6_action.go`），`normalizeNextAction` 纠正把工具名写在 `type` 的错误（eval `tool_name_as_type`）。  
- 执行经 **`mcp.Server.ExecuteCapability`** → MySQL 客户端。

---

### Q19. MCP 是什么？实现原理？本仓库怎么用？

**答**

- **Model Context Protocol**：用统一协议暴露 **资源 / 工具 / 提示**，让不同客户端（IDE、Agent 运行时）连同一套能力服务。
- 实现上：**Registry 注册能力** + **JSON 描述** + **Execute(ctx, name, input)**。

**slowlog-ai 对照**

- `internal/mcp/`：`connect_mysql_instance`、`explain_mysql_query`、`add_mysql_index`（默认 dry_run）。  
- V4 能力感知 → V5/V6 执行，见 [ARCHITECTURE · MCP](../design/ARCHITECTURE.md)。

---

### Q20. Skills 的原理是什么？和工具有什么关系？

**通用要点**

- **Skills**：把领域工作流、检查清单、脚本约定打包成 **可发现的能力包**（如 Cursor Agent Skills），强调「何时用、步骤是什么」。
- **Tools**：单次可调用、输入输出明确的函数。
- **渐进披露**：先给 Skill 目录与摘要，命中后再加载全文，避免 Prompt 膨胀。

**slowlog-ai 对照**

- **工具** = MCP 能力；**Skill 等价物** = `docs/agent/*.md`、`slowlog/docs` 知识 + ROADMAP guided 顺序。  
- **可优化**：按 `AgentPhase` 动态裁剪 `ListCapabilities()` 返回的子集，接近渐进披露。

---

### Q21. 除了 RAG，还有哪些 Agent 相关技术值得在项目里落地？

| 技术 | 与本项目关系 |
|------|----------------|
| Plan-and-Execute | 先输出固定 plan 再逐步执行，便于审计 |
| ReAct | 本仓库 V6 即 Thought(Action)+Observation 变体 |
| Tool error + retry policy | **已实现** `toolerr` |
| Eval / golden | **已实现** `make agent-eval` |
| Tracing | **已实现** `RunTrace` |
| Guardrails | 索引变更强制 dry_run、禁止无 EXPLAIN 断言加索引 |

---

### Q22. A2A（Agent-to-Agent）协议了解吗？

**答**

- Google 等推动的 **Agent 间互操作**协议，使不同厂商 Agent 可发现彼此能力、委派子任务（与 MCP「Agent↔工具」互补）。

**slowlog-ai 对照**

- 未实现；单进程单 Agent 足够。若未来「编排 Agent」派活给「DBA 工具 Agent」，可评估 A2A 与 MCP 分工。

---

## 五、MySQL 与慢日志（与本项目强相关）

### Q23. MySQL 慢查询如何定位与优化？

**答（方法论 + 本仓库自动化）**

1. **定位**：慢日志 / `performance_schema` → 锁定 SQL、库表、`Query_time`、`Rows_examined`。  
2. **分析**：`EXPLAIN`（type、key、rows、Extra 如 filesort）。  
3. **优化**：索引（左前缀、覆盖）、改写 SQL、拆查询、缓存、架构层读写分离。  
4. **验证**：对比优化前后慢日志指标；变更用 dry_run。

**slowlog-ai 对照**

- 默认场景 `testdata/slowlog-products.txt`；Agent 引导 **RAG → connect → explain → add_index(dry_run) → analyze → finish**。  
- 知识库覆盖：左前缀、filesort、覆盖索引、Rows_sent/examined 等（`make rag-check`）。

---

### Q24. MySQL 事务 ACID 指什么？

| 字母 | 含义 |
|------|------|
| A 原子性 | 事务全成功或全回滚 |
| C 一致性 | 约束不被破坏 |
| I 隔离性 | 并发事务互不不当干扰（隔离级别） |
| D 持久性 | 提交后落盘不丢 |

**slowlog-ai 对照**：诊断以 **只读 EXPLAIN + dry_run DDL** 为主，避免 Agent 直接长事务写库；锁等待见知识库 `lock_contention.md`。

---

### Q25. 什么是索引覆盖（Covering Index）？有什么好处？

**答**

- 查询所需列都在二级索引中，**无需回表**读聚簇索引，减少随机 IO。
- `EXPLAIN` 常见 `Using index`；`SELECT *` 往往破坏覆盖性。

**slowlog-ai 对照**

- 知识库 `patterns/covering_index_basics.md`；演示 SQL 含 `SELECT *`，Agent 应提示 **列裁剪** 或接受回表代价。

---

## 六、可靠性、幻觉与 Mock

### Q26. LLM 幻觉如何缓解？（Prompt 与后处理）

**通用要点**

- Prompt：要求 **引用证据**、区分 confirmed/suspected、知识库边界文。  
- 后处理：规则校验 JSON 字段、与 EXPLAIN 数字交叉验证。  
- 检索：降低温度、RAG 只注入摘要标题强制模型先「点名」知识条目。

**slowlog-ai 对照**

- V2 语义 + RAG 边界文档 + **工具失败时仍要求基于慢日志 finish**（eval）。  
- **可优化**：finish 模板强制 `证据：慢日志 Rows_examined=…；EXPLAIN type=…`。

---

### Q27. Mock 在本项目中指什么？

**答（分层）**

| 层级 | 用途 |
|------|------|
| `SLOWLOG_RAG=mock` | Agent eval 固定检索结果，轨迹确定性 |
| `eval.StubExecutor` | 工具返回固定 JSON，不连真 MySQL |
| `ScriptLLM` | 脚本化 LLM 输出，无 API 费用 |

这与「单元测试 Mock 依赖」同原理；**不是**业务上的 mock 数据生成服务。代码覆盖率、AST 插桩等 **代码分析 Agent** 话题见第七节。

---

## 七、代码分析岗常问、与本仓库边界

> 以下在 **抖音基建 Agent 开发** 类调研里出现频率高，但 **slowlog-ai 未实现**；作答时说明边界，避免强行套项目。

### Q28. 分支覆盖率怎么算？插桩原理？

**答**

- **分支覆盖率** = 已执行分支数 / 总分支数（if/switch、逻辑短路、异常路径等）。
- **插桩**：编译或运行时注入探针，记录边是否走过；Go 常用 `go test -cover`，Java 常用 JaCoCo。

**与本项目**：无关；若做「单测生成 Agent」需另建代码分析管线。

---

### Q29. 代码解析前要不要预分析？结果有效性如何判断？

**答**

- 预分析：语言识别、依赖图、变更范围、风险文件排序。  
- 有效性：AST 能否 parse、LSP 类型是否完整、与 git diff 是否一致。

**与本项目**：仅 **慢日志文本 + SQL** 解析（`slowlog_sql_test.go`），无通用 AST。

---

### Q30. 哪些代码会让模型生成单测准确率下降？AST/LSP 帮不上时如何过滤？

**答**

- 难例：反射、宏、重度泛型、CGO、动态调用、超大函数、无类型脚本。  
- 过滤：parse 失败、符号 unresolved、覆盖工具报 unreachable 的片段跳过，交人工。

**与本项目**：不适用；可参考 `toolerr` 的 **「可重试 / 勿重试」** 思路做「不可生成」标记。

---

## 八、系统与语言基础（简答）

### Q31. LLM 基本原理？输入格式是什么？

**答**

- 自回归 Transformer：预测下一个 token；训练目标为极大似然。  
- 推理输入多为 **chat messages**（system/user/assistant/tool）或拼接 Prompt；本仓库 V6 为 **单条 user Prompt 字符串**（内嵌状态与慢日志），由 DeepSeek API 适配层发送。

---

### Q32. Self-Attention 为什么有 Q、K、V？

**答**

- **Q** 查询「当前 token 需要关注什么」；**K** 被匹配索引；**V** 被聚合的信息。  
- 注意力权重 ≈ `softmax(QK^T/√d)`，输出为权重对 V 的加权和。同一 token 在不同层、不同头 Q/K/V 不同（经投影矩阵）。

---

### Q33. Python 多线程与 GIL？Lock 和 RLock？

**答**

- CPython **GIL** 使同一进程内多线程无法并行执行 Python 字节码（CPU 密集不利），I/O 阻塞时会释放 GIL。  
- **Lock**：不可重入；**RLock**：同线程可多次 acquire。

**与本项目**：主语言 **Go**，并发用 goroutine + channel；MCP/MySQL 调用在 goroutine 中阻塞 I/O 友好。

---

### Q34. Go 协程和线程的区别？

**答**

- **线程**：OS 调度，栈 MB 级，创建成本高。  
- **Goroutine**：用户态，M:N 调度到线程，栈初始 KB 级，适合高并发 I/O。  
- Agent 主循环可单 goroutine 同步跑；工具超时可 `context.WithTimeout`。

---

### Q35. HTTPS 握手过程？

**答（简版）**

1. Client Hello（套件、随机数）  
2. Server Hello + 证书  
3. 客户端验签、生成 premaster，对称密钥  
4. Change Cipher Spec，后续加密通信  

**与本项目**：LLM/MySQL 走 TLS 由部署与驱动配置决定，代码层无自定义握手。

---

### Q36. 信号量底层如何实现？

**答**

- 本质是 **计数器 + 阻塞队列**；P 操作减计数，为负则阻塞；V 操作唤醒。  
- 可用 mutex + condition variable 或 futex 实现。

---

### Q37. C++ 编译链接过程？

**答**

- **预处理** → **编译**（翻译单元 → 汇编）→ **汇编**（→ .o）→ **链接**（符号解析、重定位，生成可执行文件）。  
- 与 Go 的 **compile + link** 一次 `go build` 对比：Go 编译器自带链接，无头文件分离编译传统。

---

## 九、框架与生态（了解即可）

| 名称 | 一句话 | 本仓库 |
|------|--------|--------|
| LangGraph / AutoGen / CrewAI | 图编排 / 对话 / 角色分工 | 未引入；自研 V6 循环更轻 |
| FastAPI | Python ASGI API | 未用；Go CLI |
| OpenClaw 等开源 Agent | 生态项目在演进 | 未集成；可调研其 tool 抽象 |
| Cursor Skills | 用户级工作流说明 | 仓库外；理念见 Q20 |

---

## 十、附录：图论题「岛屿最大面积」

**题意**：二维网格 `1` 为陆地，求最大连通块面积。  
**思路**：DFS/BFS 或并查集，复杂度 O(mn)。  

**与 slowlog-ai**：无直接代码关联；属算法基本功，Agent 岗偶作手写热身，**不影响本仓库封存范围**。

---

## 十一、复习时怎么用本文

1. 先读 **第一节「可深化方向」**，决定要不要动 P2。  
2. 对外交流前，挑 **第二～六节** 与岗位 JD 重合最高的 5～8 题，用「通用 + slowlog-ai 对照」口述。  
3. 演示仍以 `make check`、`make agent-run`、`reports/*brief.html` 为主，本文不替代实操。

相关：[AI-APPLICATION-BRIEF](AI-APPLICATION-BRIEF.md) · [REVIEW-CHECKLIST](REVIEW-CHECKLIST.md) · [DEVELOP](DEVELOP.md)
