# Metric: 慢日志头字段速查

## Query_time
- 语句 **总耗时**（秒），含锁等待与执行
- 与 `Lock_time` 一起看：锁占比高时优先查并发/长事务

## Lock_time
- **等锁** 时间；高而 `Query_time` 也高 → 锁竞争或热点行
- 详见 `boundaries/lock_contention.md`

## Rows_sent
- 返回给客户端的行数；很小但 `Rows_examined` 很大 → 典型 **过滤重、索引差**

## Rows_examined
- 优化器 **估算或统计的扫描行数**；接近表行数常伴随 `type=ALL`
- 不等于「物理读盘行数」，但适合 Agent 第一轮怀疑全表/索引未命中

## 与演示慢日志的对应
- `test.products` 场景：`Rows_examined` 约 48000、`Rows_sent` 约 20 → 优先查 price 索引与左前缀

## 建议动作
- 先对照 `Rows_examined` / `Rows_sent` 比值，再决定是否 EXPLAIN
- 勿单独用 `Query_time` 判断根因（可能是锁或网络）
