# HTTP API（G11 最小服务）

[生产化缺口](../agent/PRODUCTION-GAPS.md) · [RUN](../agent/RUN.md)

`cmd/agent-api` 提供同步分析接口，适合 Webhook / 内部平台调用。默认 **V6**、写入 `reports/`。

---

## 启动

```bash
cp .env.example .env   # DEEPSEEK_API_KEY
make agent-api
# 或
SLOWLOG_API_ADDR=:8080 go run ./cmd/agent-api
```

| 环境变量 | 默认 | 说明 |
|----------|------|------|
| `SLOWLOG_API_ADDR` | `:8080` | 监听地址 |
| `SLOWLOG_REPORT_DIR` | `reports` | 报告目录 |
| `SLOWLOG_ANALYZE_TIMEOUT` | `15m` | 单次分析超时 |

---

## 接口

### `GET /v1/health`

```json
{"status":"ok"}
```

### `POST /v1/analyze`

**Body**（二选一）：

- `Content-Type: application/json` → `{"slow_log":"...", "guided": true}`
- `Content-Type: text/plain` → 慢日志原文

**Query**：`?guided=false` 关闭推荐流程（缩短轮次）。

**Response 200**：

```json
{
  "report_id": "agent-run-20260525-120000",
  "iterations": 6,
  "final_result": "...",
  "json_path": "reports/agent-run-....json",
  "brief_html_path": "reports/agent-run-....brief.html"
}
```

**错误**：`{"error":"..."}`（400/500）

### `GET /v1/reports/{report_id}`

返回报告 JSON 文件内容。

### `GET /v1/reports/{report_id}/brief.html`

返回精简 HTML 报告。

---

## 示例

```bash
curl -s http://127.0.0.1:8080/v1/health

curl -s -X POST http://127.0.0.1:8080/v1/analyze?guided=false \
  -H 'Content-Type: text/plain' \
  --data-binary @testdata/slowlog-index-hit.txt

curl -s http://127.0.0.1:8080/v1/reports/agent-run-20260525-120000 | head
```

---

## 边界（PoC）

- **同步阻塞**：长慢日志分析可能数分钟，调用方需设客户端超时 ≥ `SLOWLOG_ANALYZE_TIMEOUT`。
- **无鉴权**：生产需加 API Key / mTLS / 内网隔离。
- **HITL 关闭**：HTTP 路径不等待 stdin；人机协同仍用 CLI。
- **每请求独立 MCP 连接**：高 QPS 时需连接池（P3）。
