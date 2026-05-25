# Pattern: EXPLAIN type=ALL 与 key 为空

## 含义
- `type=ALL`：全表扫描（最差一档之一）
- `key=NULL`：未使用二级索引查找

## 与慢日志的关联
- 常伴随 `Rows_examined` 接近表 cardinality
- 需与 **索引是否存在、左前缀是否命中** 区分

## 下一步
- 核对表结构与现有索引列表
- 判断应 **改 SQL** 还是 **加索引** 或 **改索引列顺序**
- 生产变更前用 **dry_run** 生成 DDL，评估写入开销

## 注意
- 个别引擎统计下 type 可能显示 range 但仍扫大量行，以 rows 与慢日志为准
