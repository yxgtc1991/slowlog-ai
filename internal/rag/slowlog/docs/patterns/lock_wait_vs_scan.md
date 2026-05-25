# Pattern: 锁等待 vs 全表扫描

## 典型场景
- `Lock_time` 接近 `Query_time`，且 `Lock_time` 明显 > 0.1s
- `Rows_examined` 很小（个位数～几百），但语句仍「慢」

## 机制说明
- 瓶颈在 **InnoDB 行锁/间隙锁等待**，而非读盘扫描
- 常见触发：热点行 UPDATE、长事务未提交、死锁重试
- 加索引 **不能** 直接降低锁等待时间

## 与测试慢日志的对应
- `UPDATE orders SET status=... WHERE id=...`，Lock_time 7.9s、examined 12
- 应先查 **并发事务、锁监控**，而非建议 `(id)` 索引（主键通常已有）

## 建议动作
- 看 `performance_schema` / `innodb_lock_waits`、应用重试与事务边界
- 缩短事务、拆分热点更新、乐观锁或队列削峰

## 常见误判
- 见慢日志就建议加索引或改 SQL 扫描路径
- 忽略 `Lock_time` 字段，只分析 `Rows_examined`
