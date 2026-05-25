# 文档索引

[项目 README](../README.md)

本目录按主题分子文件夹；**细节进子目录**，此处只做入口。

> 本文件名为 **INDEX.md**（避免与项目根目录 README.md 重名，GoLand 才不易开错页）。

---

## 目录结构

```text
docs/
├── INDEX.md           你在这里（文档入口）
├── agent/             Agent 路线、运行、回归、模式、讲解稿
├── design/            版本演进、架构与扩展
├── guides/            RAG 使用说明
└── diagrams/          流程图（PNG + mmd）
```

---

## 我想…

| 目的 | 打开 |
|------|------|
| **总览、路线、命令、汇报提纲** | [agent/ROADMAP.md](agent/ROADMAP.md) |
| 跑 Agent 并生成报告 | [agent/RUN.md](agent/RUN.md) |
| 回归测试（无 API） | [agent/EVAL.md](agent/EVAL.md) |
| V5 / V6 模式切换 | [agent/MODE.md](agent/MODE.md) |
| AI 应用向讲解稿 | [agent/AI-APPLICATION-BRIEF.md](agent/AI-APPLICATION-BRIEF.md) |
| V1–V6 设计与代码示例 | [design/VERSIONS.md](design/VERSIONS.md) |
| 接口、MCP、MySQL、扩展 | [design/ARCHITECTURE.md](design/ARCHITECTURE.md) |
| RAG 模式与环境变量 | [guides/RAG.md](guides/RAG.md) |
| V6 / RAG 流程图 | [diagrams/v6-agent-flow.md](diagrams/v6-agent-flow.md) · [diagrams/rag-flow.md](diagrams/rag-flow.md) |

---

## 命令速查

```bash
make agent-eval          # 回归
make agent-run           # V6 + 报告
make rag-check           # 检索试跑（见 guides/RAG.md）
make doc-links           # 校验文档链接
```

完整说明见 [agent/ROADMAP.md](agent/ROADMAP.md)。

---

## IDE 预览（GoLand）

- 流程图请看 diagrams 目录下的 **PNG**。
- md 正文不要写 mermaid 代码块；改 .mmd 后执行 make doc-diagrams。
- **整页发黑**：File → Invalidate Caches → Restart（最常见修复）。仍不行见根目录 **IDE-GOLAND.txt**，或右键 md → Open in → Text Editor。
