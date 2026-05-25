# Metric: Rows_sent

## 含义
慢日志中 **返回给客户端的行数**（结果集行数）。

## 与 Rows_examined 对比
- **Rows_sent 小、Rows_examined 大**：扫描多、返回少，常见于过滤差或排序后 LIMIT
- **两者接近**：过滤选择性较好，或命中覆盖索引

## 演示慢日志解读
- `Rows_sent: 20` 且 `LIMIT 20`：返回行符合预期
- 不能因 Rows_sent 小就认为查询便宜

## 注意
- 须结合 Query_time、EXPLAIN 一起看
