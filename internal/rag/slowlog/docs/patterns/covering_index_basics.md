# Pattern: 覆盖索引与回表

## 典型场景
- EXPLAIN 出现 `Using index`（覆盖索引）时，通常 **无需回表** 读主键行
- 若 `Extra` 含 `Using index condition` 或仍需回表，扫描成本仍可能偏高
- 慢日志 `Rows_examined` 明显大于 `Rows_sent`

## 机制说明
- **覆盖索引**：查询所需列都在二级索引叶子里，优化器只扫索引
- **回表**：索引未覆盖 SELECT 列，需按主键再读聚簇索引行
- `ORDER BY` / `LIMIT` 列若不在索引中，易出现 filesort 或额外排序

## 与演示慢日志的对应
- `SELECT * FROM products WHERE price >= 100 ORDER BY created_at DESC LIMIT 20`
- `SELECT *` 往往 **无法覆盖**，即使存在 `(price, created_at)` 也可能要回表读其余列

## 建议动作
- 评估改为 **只查必要列**（列裁剪）以争取覆盖
- 或接受 `(price, created_at)` 索引 + 回表，对比 `Rows_examined` 降幅

## 常见误判
- 看到 `Using index` 就认为一定很快（可能是 ICP 或部分列）
- 忽略 `SELECT *` 对覆盖性的破坏
