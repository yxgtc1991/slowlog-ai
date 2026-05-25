# Boundary: 锁竞争 vs 索引问题

## 边界说明
- 本 Agent **不能**仅凭慢日志断定锁等待根因
- `Lock_time` 高时，优化索引可能 **无效**

## 应优先排查
- 长事务、未提交事务
- 热点行更新、间隙锁
- 并发 DDL

## 与索引优化的分工
- Rows_examined 高 → 倾向扫描/索引
- Lock_time 高、Rows_examined 不高 → 倾向锁，转 DBA 看 `performance_schema`

## Agent 话术
- 明确区分「扫行多」与「等锁久」
- 避免一律建议加索引
