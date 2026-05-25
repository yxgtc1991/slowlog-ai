# 日志接入设计（Fluent Bit → agent-api）

[生产化缺口](../agent/PRODUCTION-GAPS.md) · [HTTP API](API.md) · [架构](ARCHITECTURE.md)

> **状态**：**Webhook 已实现**（`POST /v1/ingest`）；对象存储事件、多租户仍为 P3 后续项。

---

## 目标

将 **已采集的慢日志** 自动送入 Agent 诊断，输出 `report_id` 供平台轮询或回调。

---

## 推荐链路

```text
MySQL slow log
    → Fluent Bit（K8s DaemonSet / Sidecar）
    → 过滤（Query_time 阈值，可选）
    → POST /v1/ingest（async 默认 true）
    → agent-api 后台 RunV6
    → reports/agent-run-*.json + brief.html
    → GET /v1/jobs/{id} 或 callback_url
```

---

## Webhook 接口

见 [API.md · POST /v1/ingest](API.md)。

| 能力 | 说明 |
|------|------|
| 异步 | 默认 `async: true`，立即 `202` + `job_id` |
| 同步 | `"async": false` 等同 `/v1/analyze` |
| 鉴权 | 可选 `SLOWLOG_WEBHOOK_SECRET` + 请求头 `X-Webhook-Secret` |
| 阈值 | `SLOWLOG_INGEST_MIN_QUERY_TIME`（秒），低于则 `skipped` |
| 回调 | `callback_url` 完成后 POST 任务 JSON |

---

## Fluent Bit 示例

仓库内示例配置：[examples/fluent-bit-ingest.conf](../../examples/fluent-bit-ingest.conf)。

要点：

1. `tail` 或多行 parser 读慢日志文件  
2. `grep` 过滤 Query_time（或在服务端用环境变量阈值）  
3. `http` output 指向 `http://<agent-api>:8080/v1/ingest`  
4. Body 为 JSON：`{"slow_log":"<整段慢日志>","source":"fluent-bit","async":true}`  

---

## 触发策略

| 策略 | 实现 |
|------|------|
| 阈值 | Fluent Bit `grep` 或 `SLOWLOG_INGEST_MIN_QUERY_TIME` |
| 手动 | `POST /v1/analyze` / `make agent-run` |
| 聚合 | 未实现（需日志平台侧 digest 聚合） |

---

## 与仓库边界

| 已有 | 后续 |
|------|------|
| `POST /v1/ingest`、内存 JobStore | Redis/DB 持久化任务 |
| callback HTTP POST | 签名、重试、死信 |
| 示例 Fluent Bit 片段 | 各云厂商完整 Helm |

实现跟踪：[PRODUCTION-GAPS · G12](../agent/PRODUCTION-GAPS.md)。
