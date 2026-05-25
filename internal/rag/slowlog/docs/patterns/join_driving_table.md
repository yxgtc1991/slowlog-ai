# Pattern: JOIN 驱动表与连接顺序

## 典型场景
- 慢日志 `Rows_examined` 极大，SQL 含 **多表 JOIN**
- `EXPLAIN` 中某表 `type=ALL` 或 `rows` 很大，另一表正常

## 机制说明
- 优化器选择 **驱动表**（通常小结果集或索引过滤强的一方）
- 被驱动表若无 **连接键索引**（如 `order_items.order_id`），会对驱动行逐行探测
- `WHERE` 落在驱动表列上时，应优先保证驱动表有合适索引

## 与测试慢日志的对应
- `orders` + `order_items`，条件在 `orders.created_at`、`status`
- 应检查 `orders(created_at, status)` 与 `order_items(order_id)` 索引，而非只盯 SELECT 列表

## 建议动作
- EXPLAIN 看 **每表** type/key/rows，定位「拖后腿」的那张表
- 避免盲目在 SELECT 列上建索引而忽略 **JOIN 键**

## 常见误判
- 只看总 Query_time，不拆表看执行计划
- 对 JOIN 建议「加 LIMIT」却不改驱动条件或索引
