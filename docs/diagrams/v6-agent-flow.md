# V6 Agent 流程图

← [README · V6 Agent 执行流程](../../README.md#v6-agent-执行流程)

![V6 Agent 执行流程](./v6-agent-flow.png)

**说明**：决策在顶部「行动类型」；四条分支写入 `context` 后汇总为「构建 Prompt」，再经 LLM 输出 `NextAction` 回到下一轮。`finish` 分支结束循环（图中右侧）。

可选 Mermaid 源码（GitHub 可渲染）：

```mermaid
flowchart TD
  C{行动类型}
  C -->|retrieve_rag| D[检索知识库 写入 context]
  C -->|call_tool| E[执行 MCP 写入 context]
  C -->|analyze| F[分析写入 context]
  C -->|ask_question| G[记录问题]
  C -->|finish| H[结束]
  D --> P[构建 Prompt]
  E --> P
  F --> P
  G --> P
  P --> B[LLM 输出 NextAction]
  B --> C
```
