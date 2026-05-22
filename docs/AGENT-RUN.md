# V6 Agent 完整体验（带报告存档）

← [路线图](AGENT-ROADMAP.md) · [README](../README.md)

一次运行即可把 **每轮 LLM 决策、RAG 命中、MCP 工具返回、最终结论** 写入 `reports/`，后续查阅**无需再调 API**。

## 前置条件

```bash
cp .env.example .env
# 必填
DEEPSEEK_API_KEY=sk-xxx
# 若要走 MySQL / EXPLAIN / 建索引 dry_run（推荐）
MYSQL_HOST=127.0.0.1
MYSQL_USER=root
MYSQL_PASSWORD=...
MYSQL_DATABASE=test
```

本地需有 `test.products` 表，字段与 `testdata/slowlog-products.txt` 注释一致（`code`, `price`, `created_at` 等；示例 SQL 按 `price` 过滤以触发全表扫描场景）。

## 一条命令跑完全流程

```bash
make agent-run
# 或
go run ./cmd/agent-run

# 指定慢日志与报告目录
go run ./cmd/agent-run -report=reports/my-demo testdata/slowlog-products.txt
```

默认：

- 慢日志：`testdata/slowlog-products.txt`（`products` 表全表扫描场景）
- **guided**：Prompt 内建议 RAG → 连库 → EXPLAIN → `add_mysql_index` 仅 `dry_run=true` → finish
- **trace**：stderr 实时打印每轮决策
- **报告**（同一次运行、同一时间戳）：
  - **精简**（推荐给客户）：`agent-run-*.brief.html` / `.brief.md` — 表格 + 每轮「做了什么 / 为什么 / 结果」
  - **完整**（复盘细节）：`agent-run-*.html` / `.md`
  - **数据**：`agent-run-*.json`

## 报告里有什么

| 文件 | 用途 |
|------|------|
| `*.brief.html` | **一眼看逐轮**：结论摘要 + 四列表格 + 每轮卡片详情 |
| `*.brief.md` | 同上，Markdown 版 |
| `*.html` / `*.md` | 完整版：慢日志原文、工具 JSON、RAG 汇总等 |
| `*.json` | 机器可读，含 `rounds[].llm_raw` |

报告 Markdown 已规避会被当成 HTML 的写法（例如 Go 打印的 nil 尖括号），GoLand 预览应可正常显示；若仍异常可打开 **`.html`**。从 JSON 重生报告：

```bash
make report-md JSON=reports/agent-run-YYYYMMDD-HHMMSS.json
```

完整 LLM 原文在同名 `.json` 的 `rounds[].llm_raw` 中。

## 参数

| 参数 | 说明 |
|------|------|
| 第一个非 flag 参数 | 慢日志文件路径 |
| `-report=目录` | 报告输出目录（默认 `reports`） |
| `-guided=false` | 关闭推荐流程，完全由模型自由规划 |
| `-trace=false` | 关闭 stderr 实时轨迹 |

## 与 `make run` 的区别

| 命令 | 作用 |
|------|------|
| `make run` | 演示 V6，控制台输出为主 |
| `make agent-run` | **完整存档**，适合深度体验与事后复盘 |

`reports/` 已在 `.gitignore`，不会误提交 Token 轨迹。
