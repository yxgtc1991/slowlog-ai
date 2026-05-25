# V6 Agent 流程图

[文档索引](../INDEX.md) · [V6 执行流程](../../README.md#v6-agent-执行流程)

![V6 Agent 执行流程](./v6-agent-flow.png)

**说明**：决策在顶部「行动类型」；四条分支写入 `context` 后汇总为「构建 Prompt」，再经 LLM 输出 `NextAction` 回到下一轮。`finish` 分支结束循环（图中右侧）。

改图：编辑 [v6-agent-flow.mmd](./v6-agent-flow.mmd)，执行 `make doc-diagrams`（若已为该文件配置生成；当前 PNG 已入库，可直接替换 PNG）。

> **GoLand 预览**：请直接看上方 PNG；勿在 `.md` 里嵌 Mermaid，否则预览常空白。
