# 运维：多实例、审计与限流（G14 / G15）

[API](API.md) · [生产化缺口](../agent/PRODUCTION-GAPS.md)

PoC 级进程内能力：多库实例元数据 + 操作可追溯 + 成本护栏，非完整 SaaS 多租户。

---

## 实例注册表

`SLOWLOG_INSTANCES_FILE` 指向 JSON（示例见 `config/instances.example.json`）：

```json
{
  "instances": [
    {"id": "default", "label": "本地 .env MySQL"},
    {"id": "readonly", "allowed_tools": ["connect_mysql_instance", "explain_mysql_query"]}
  ]
}
```

| Header / Query | 说明 |
|----------------|------|
| `X-Instance-ID` | 首选 |
| `?instance_id=` | 备选 |
| `SLOWLOG_REQUIRE_INSTANCE_ID=1` | 未传则 400 |

`GET /v1/instances` 列出已注册实例（不含密码）。

`allowed_tools` 当前仅写入报告元数据，Agent 侧强制过滤为后续项。

---

## 审计 JSONL

默认 `data/audit.jsonl`（`SLOWLOG_AUDIT_PATH` 可改）。每次 `analyze` / `ingest` / `rag_rebuild` 追加一行：

| 字段 | 含义 |
|------|------|
| `action` | `analyze` / `ingest` / `rag_rebuild` |
| `status` | `started` / `ok` / `failed` / `skipped` |
| `instance_id` | 实例 |
| `request_id` | 与响应头 `X-Request-ID` 一致 |
| `report_id` | 成功时报告 ID |
| `actor` | `X-Actor`（调用方标识，PoC 不做鉴权） |

报告 JSON 亦含 `instance_id`、`request_id`、`actor`、`client_ip`。

---

## 管理接口鉴权（PoC RBAC）

`SLOWLOG_ADMIN_TOKEN` 非空时，`POST /v1/rag/rebuild` 与 `GET /v1/rag/status` 需：

```http
Authorization: Bearer <SLOWLOG_ADMIN_TOKEN>
```

未配置 admin token 时，仍可用 `X-Webhook-Secret`（与 ingest 相同）。

---

## G15 限流与配额

| 环境变量 | 默认 | 说明 |
|----------|------|------|
| `SLOWLOG_RATE_LIMIT_PER_MIN` | `0`（关） | 全进程每分钟分析次数 |
| `SLOWLOG_DAILY_ANALYZE_QUOTA` | `0`（关） | 自然日分析次数 |
| `SLOWLOG_MAX_CONCURRENT` | `0`（关） | 同时进行中的 analyze/ingest/rebuild |

对 `/v1/analyze`、`/v1/ingest`、`POST /v1/rag/rebuild` 生效；超限返回 429 或 503。

`GET /v1/health` 在启用限流时附带 `limits` 计数快照。

---

## 示例

```bash
export SLOWLOG_INSTANCES_FILE=config/instances.example.json
export SLOWLOG_RATE_LIMIT_PER_MIN=10
export SLOWLOG_DAILY_ANALYZE_QUOTA=100

curl -s http://127.0.0.1:8080/v1/instances
curl -s -X POST http://127.0.0.1:8080/v1/analyze \
  -H 'X-Instance-ID: default' -H 'X-Actor: ci-bot' \
  -H 'Content-Type: text/plain' --data-binary @testdata/slowlog-index-hit.txt

tail -f data/audit.jsonl
```
