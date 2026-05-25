# 版本演进详解

← [路线图](AGENT-ROADMAP.md) · [README](../README.md) · [架构与扩展](ARCHITECTURE.md)

日常速查（总览表、如何切换演示）见 [README · 版本演进速查](../README.md#版本演进速查)。

---

## 📜 Prompt 演进历程（详解）

> 与 [版本演进速查](../README.md#版本演进速查) 配合阅读：速查表看「优化了什么」，本章看「怎么实现的」。

本项目展示从基础 Prompt 到能力感知、Tool Calling、Agent 的完整演进；**每一版针对上一版的明确缺陷**，而不是简单堆功能。

#### 📌 V1 基础版本：问模型（Ask the Model）

**设计思路**：最直接的方式，将问题抛给 LLM，依赖模型的通用知识。

**核心特点**：
- 简单的角色设定："你是一个 MySQL 专家"
- 直接提问："请分析以下慢日志并给出优化建议"
- 无任何约束，完全依赖模型的自由发挥

**代码示例**：
```go
func BuildSlowLogPrompt(slowLog string) string {
    return fmt.Sprintf(`
你是一个 MySQL 专家，请分析以下慢日志并给出优化建议：

%s
`, slowLog)
}
```

**存在的问题**：
- ❌ 输出格式不统一，难以程序化处理
- ❌ 模型可能过度假设，给出无法验证的建议
- ❌ 缺少结构化输出，无法区分"确定的问题"和"推测的问题"
- ❌ 无法控制模型的推理边界

**适用场景**：快速原型、概念验证、人工审查场景

---

#### 🔒 V2 严格模式：约束模型（Constrain the Model）

**设计思路**：通过严格的规则约束，让模型只基于提供的信息进行分析，避免过度假设。

**核心改进**：
- ✅ **角色明确化**：从"专家"变为"分析组件"，强调工具属性
- ✅ **规则约束**：5 条严格规则，禁止假设、禁止无法推导的结论
- ✅ **结构化输出**：强制 JSON 格式，包含 `confirmed_issues` 和 `suspected_issues`
- ✅ **边界明确**：明确区分"可确认"和"需要更多信息"

**代码示例**：
```go
var StrictPromptTemplate = `
你是一个【MySQL 慢查询分析组件】，不是聊天助手。

你必须严格遵守以下规则：
1. 只能基于提供的慢日志内容进行分析
2. 不允许假设表结构、索引或业务逻辑
3. 不确定的信息必须明确标注为"需要更多信息"
4. 禁止给出无法从日志中直接推导的结论
5. 输出必须是 JSON，不允许任何额外说明文字
...
`
```

**解决的问题**：
- ✅ 输出格式统一，可程序化处理
- ✅ 避免模型过度假设
- ✅ 明确区分确定性和推测性结论
- ✅ 可追溯的分析依据

**技术亮点**：
- **Prompt 约束技术**：通过明确的规则限制模型行为
- **结构化输出**：JSON Schema 约束输出格式
- **置信度管理**：通过 `confidence` 字段管理不确定性

**仍存在的问题**：
- ⚠️ 模型缺乏专业知识，可能遗漏常见问题模式
- ⚠️ 无法利用历史经验和最佳实践
- ⚠️ 分析深度受限于模型的通用知识

---

#### 🧠 V3 RAG 增强：引入知识库（Retrieval-Augmented Generation）

**设计思路**：通过 RAG（检索增强生成）技术，将专家知识库注入到 Prompt 中，提升分析的准确性和深度。流程对比见 [diagrams/rag-flow.md](diagrams/rag-flow.md)。

**核心改进**：
- ✅ **知识检索**：从知识库中检索相关的模式、反模式、指标说明
- ✅ **知识注入**：将检索到的知识块注入到 Prompt 中
- ✅ **知识约束**：明确知识仅用于推断 `suspected_issues`，不能作为 `confirmed_issues`
- ✅ **降级机制**：如果没有检索到知识，自动降级到 v2

**代码示例**：
```go
func BuildSlowLogPromptV3(slowLog string, ragChunks []rag.KnowledgeChunk) string {
    // 构建知识块描述
    var sb strings.Builder
    for _, c := range ragChunks {
        sb.WriteString(fmt.Sprintf("- %s：%s\n", c.Title, c.Content))
    }
    
    return fmt.Sprintf(`
%s  // v2 的严格模板

【可参考的专家知识（仅用于推断 suspected_issues，不得作为 confirmed_issues）】
%s
`, BuildSlowLogPromptV2(slowLog), sb.String())
}
```

**解决的问题**：
- ✅ 利用专家知识，提升分析准确性
- ✅ 识别常见性能问题模式
- ✅ 避免常见误解和反模式
- ✅ 提供更专业的优化建议

**技术亮点**：
- **RAG 架构**：检索 + 生成，结合外部知识
- **知识分类**：模式（Patterns）、反模式（Anti-patterns）、指标（Metrics）
- **知识边界**：明确知识的使用范围和限制
- **优雅降级**：无知识时回退到 v2

**知识库结构**：
```
internal/rag/slowlog/docs/
├── patterns/          # 性能问题模式
├── anti-patterns/     # 常见误解
├── metrics/           # 指标说明
├── boundaries/        # 分析边界
└── actions/           # 优化建议
```

**仍存在的问题**：
- ⚠️ 知识库是静态的，需要人工维护
- ⚠️ 检索质量依赖查询策略
- ⚠️ 系统能力是固定的，无法动态扩展

---

#### 🎯 V4 能力感知：能力抽象与感知（Capability Awareness）

**设计思路**：将系统能力抽象为可发现、可描述的接口，让 LLM 能够感知和选择使用哪些能力。

**核心改进**：
- ✅ **能力抽象**：定义 `Capability` 接口，统一能力描述
- ✅ **能力发现**：系统可以自动列出所有可用能力
- ✅ **能力描述**：生成 LLM 可理解的能力说明
- ✅ **动态扩展**：新能力可以动态注册，无需修改核心代码

**代码示例**：
```go
// 能力接口定义
type Capability interface {
    Name() string                    // 能力名称
    Description() string             // 能力说明
    InputSchema() map[string]string  // 输入参数
    Execute(ctx, input) (interface{}, error)
}

// v4 Prompt：生成能力描述
func BuildCapabilityPromptV4(caps []CapabilityV4) string {
    // 为每个能力生成描述
    // 告诉 LLM 可以使用哪些工具
}
```

**解决的问题**：
- ✅ 系统能力可发现、可描述
- ✅ LLM 可以自主选择使用哪些能力
- ✅ 新能力可以动态添加
- ✅ 为 MCP（Model Context Protocol）奠定基础

**技术亮点**：
- **接口抽象**：通过接口统一能力定义
- **能力注册表**：Registry 模式管理能力
- **自描述系统**：系统可以描述自己的能力
- **MCP 兼容**：符合 Model Context Protocol 设计理念

**架构设计**：
```
Capability Interface
    ↓
Registry (能力注册表)
    ↓
Server (能力服务器)
    ↓
Prompt V4 (能力描述生成)
```

**演进价值**：
- 🎯 **可扩展性**：新能力只需实现接口即可
- 🎯 **可发现性**：系统自动发现和描述能力
- 🎯 **可组合性**：多个能力可以组合使用
- 🎯 **标准化**：为 MCP 协议做准备

---

#### ⚡ V5 Tool Calling：协议级工具调用

**设计思路**：不再让模型在正文里写「我要调用某某工具」，而是走 DeepSeek **Tool Calling** API，由模型返回结构化的 `tool_calls`，运行时执行 MCP 能力后再把结果喂回模型。

**相对 V4 的优化**：
- ✅ 工具名与参数由 API 约束，减少幻觉与解析失败
- ✅ 支持多轮：分析 → 调工具 → 再总结
- ✅ 与 V4 共用同一套 `Capability` / `mcp.Server`

**关键文件**：`internal/analyzer/v5_tool_calling.go`、`internal/prompt/slowlog/v5_intent.go`、`internal/llm/deepseek_tool_adapter.go`

**仍存在的问题**：行动类型单一（主要是 call_tool），难以在同一轮里显式「先 RAG 再工具再追问」→ 引出 V6。

---

#### 🤖 V6 Agent：LLM 自主决策下一步行动
- **文件**：`internal/prompt/slowlog/v6_action.go`、`internal/analyzer/v6_agent.go`
- **特点**：LLM 自主决定下一步要做什么（调用工具、检索 RAG、继续分析、提问、完成）
- **适用**：真正的 AI Agent、自主规划任务

**设计思路**：从"LLM 决定是否调用工具"升级到"LLM 决定下一步做什么"，实现真正的 Agent 模式。

**核心改进**：
- ✅ **多行动类型**：支持 5 种行动类型（call_tool、retrieve_rag、analyze、ask_question、finish）
- ✅ **自主规划**：LLM 可以自主规划分析步骤，决定何时调用工具、何时检索知识库、何时完成
- ✅ **上下文感知**：每次决策都基于当前上下文（已执行的工具结果、RAG 检索结果等）
- ✅ **对话历史**：维护完整的对话历史，支持多轮交互

**行动类型说明**：
1. **call_tool** - 调用工具：当需要调用系统工具时使用
2. **retrieve_rag** - 检索知识库：当需要查询相关知识时使用
3. **analyze** - 继续分析：基于已有信息继续分析
4. **ask_question** - 提出问题：需要用户提供额外信息时使用
5. **finish** - 完成分析：收集足够信息，输出最终结果

**代码示例**：
```go
// 创建 V6 Agent 分析器
v6Analyzer := analyzer.NewV6AgentAnalyzer(
    llmClient,              // 使用普通 LLM 客户端（不需要 Tool Calling）
    ragRetriever,           // RAG 检索器
    capabilityExecutor,     // 能力执行器
    availableTools,          // 可用工具列表
)

// 执行分析（LLM 自主决定每一步要做什么）
result, err := v6Analyzer.Analyze(ctx, slowLog)

// 查看执行轨迹
fmt.Printf("迭代次数：%d\n", result.Iterations)
for i, action := range result.Actions {
    fmt.Printf("%d. [%s] %s\n", i+1, action.Type, action.Reasoning)
}
```

**与 V5 的区别**：
- **V5**：LLM 决定"是否调用工具"，系统负责执行工具调用
- **V6**：LLM 决定"下一步做什么"，可以是调用工具、检索 RAG、继续分析、提问或完成

### 📊 版本对比总结（简表）

完整列（含关键文件、痛点对照）见 [版本演进速查 · 总览表](../README.md#总览表v1-v6)。

| 版本 | 核心思想 | 输出 / 交互 | 扩展性 | 典型场景 |
|------|---------|-------------|--------|----------|
| **V1** | 问模型 | 自由文本 | ❌ | PoC |
| **V2** | 约束模型 | JSON | ❌ | 可解析的自动分析 |
| **V3** | RAG | JSON + 知识块 | ⚠️ | 带专家知识的分析 |
| **V4** | 能力感知 | 能力元数据 + 调用 | ✅ | MCP / 工具发现 |
| **V5** | Tool Calling | API tool_calls | ✅ | 可靠自动调工具 |
| **V6** | Agent | NextAction 多轮 | ✅ | 自主规划整条分析链路 |

### 🎓 学习要点（复盘）

1. **V1 → V2**：Prompt 约束与结构化输出
2. **V2 → V3**：RAG 边界（confirmed vs suspected）
3. **V3 → V4**：能力抽象、Registry、系统自描述
4. **V4 → V5**：从「文本意图」到「协议级 Tool Calling」
5. **V5 → V6**：从「只会调工具」到「多行动 Agent 状态机」

### 💡 设计模式应用

- **V2**：Prompt Template
- **V3**：RAG（检索 + 生成）
- **V4**：Capability + Registry
- **V5**：Adapter（`ToolCallingLLMClient`）+ Executor（`mcp.Server`）
- **V6**：Agent 循环 + 上下文累积（`map[string]interface{}` 工具/RAG 结果）
