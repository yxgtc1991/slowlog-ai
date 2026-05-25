# Anti-pattern: 见慢 SQL 就加索引

## 典型表现
- `Rows_examined` 已接近 `Rows_sent`，`EXPLAIN` 有 `key` 且 rows 很小
- 仍建议新建复合索引，写入与存储成本上升、收益极低

## 机制说明
- 慢可能来自 **网络、锁、缓存冷启动**，索引已够用时加索引无效
- 重复索引、低选择性列索引会导致优化器选择困难

## 与测试慢日志的对应
- `products` 上 `idx_price_created` 已命中，examined≈22、sent=20
- 应输出「索引已利用，查其它瓶颈」而非 DDL

## 建议动作
- 对比优化前后 `Rows_examined` 与 `key` 是否已存在
- 用 dry_run 评估 DDL 再决定是否上线

## 常见误判
- Agent/规则引擎不读 EXPLAIN 就生成 `ALTER TABLE`
