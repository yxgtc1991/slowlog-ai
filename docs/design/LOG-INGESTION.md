# 日志接入设计（P3 路线图）

[生产化缺口](../agent/PRODUCTION-GAPS.md) · [架构](ARCHITECTURE.md)

> **状态**：设计草案，**未实现**。用于说明 slowlog-ai 与现有日志采集（如 Fluent Bit on K8s）如何衔接，不表示已上线。

---

## 目标

将 **已落盘的慢日志** 自动或半自动送入 Agent 诊断，输出结构化报告供 DBA/研发复核。

---

## 推荐链路（与 Fluent Bit 叙事对齐）

```text
MySQL slow log
    → Fluent Bit（DaemonSet / Sidecar）采集
    → 对象存储或日志平台（S3 / ES / Loki / 自建）
    → 规则/阈值触发（Query_time、Rows_examined）
    → 调用诊断服务（未来：POST /analyze）
    → slowlog-ai Agent（RAG + MCP + 报告 JSON/HTML）
    → 工单 / IM / 邮件（人工审批 DDL）
```

---

## 触发策略

| 策略 | 说明 |
|------|------|
| 阈值 | 单条 Query_time > N 秒或 Rows_examined > M |
| 聚合 | 同一 digest 5 分钟内出现 K 次 |
| 手动 | 平台粘贴慢日志，等同当前 `make agent-run` |

---

## 与当前仓库边界

| 已有 | 待建（G11/G12） |
|------|-----------------|
| CLI `agent-run`、报告 `reports/` | HTTP 服务、任务队列 |
| MCP dry_run | 变更审批流、审计库 |
| `make check` 回归 | 接入层监控与 SLA |

---

## 最小可行接入（建议实施顺序）

1. **Webhook**：日志平台 POST 慢日志正文 → 异步跑 Agent → 回调报告 URL  
2. **对象存储**：新文件事件 → Lambda/Job 拉取 → 同上游  
3. **多租户**：`instance_id` 映射不同 `MYSQL_*` 凭证（密钥托管，不进 Git）  

实现跟踪见 [PRODUCTION-GAPS · G11/G12](../agent/PRODUCTION-GAPS.md)。
