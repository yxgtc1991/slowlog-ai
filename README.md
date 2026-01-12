# slowlog-ai

🚀 **基于 LLM 的 MySQL 慢日志智能分析工具（Golang）**

`slowlog-ai` 是一个使用 Go 编写的实验性项目，旨在探索 **将大语言模型（LLM）引入云数据库 MySQL 慢日志分析场景**，自动完成 SQL 慢查询的原因分析与优化建议生成。

---

## ✨ 项目背景

在云数据库 MySQL 场景中，慢日志通常存在以下痛点：

- 慢 SQL 数量大，人工分析成本高
- SQL 优化依赖经验，质量不稳定
- 现有工具多聚焦统计，缺少“原因 + 建议”的语义分析

本项目尝试引入 **LLM（如 DeepSeek）**，通过 Prompt Engineering，对 MySQL 慢日志进行**自动化智能分析**，作为 DBA / 平台侧的辅助决策工具。

---

## 🧠 核心能力（当前）

- ✅ 解析 MySQL 慢日志文本
- ✅ 构造慢日志分析 Prompt
- ✅ 调用 LLM（DeepSeek Chat API）
- ✅ 输出 SQL 慢查询的潜在问题与优化建议

> 当前阶段以 **Prompt 设计与工程结构演进** 为重点，暂不涉及完整日志解析器与 RAG。

---

## 📁 项目结构

```text
slowlog-ai/
├── cmd/
│   └── slowlog-ai/
│       └── main.go          # 程序入口
├── internal/
│   ├── llm/
│   │   └── deepseek.go      # LLM 客户端封装
│   ├── prompt/
│   │   └── slowlog.go       # 慢日志 Prompt 构造
│   └── config/
│       └── config.go        # 配置与环境变量读取
├── .env.example             # 环境变量示例（不提交密钥）
├── go.mod
├── go.sum
└── README.md