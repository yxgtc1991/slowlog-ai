# Anti-pattern: 深分页 OFFSET

## 典型表现
- `ORDER BY ... LIMIT 500 OFFSET 100000` 越来越慢
- `Rows_examined` 随 OFFSET 增大而上升

## 机制说明
- 优化器常需 **排序并跳过** 前 OFFSET 行，无法只读最后 LIMIT 行
- 即使有索引，大 OFFSET 仍可能大量回表或 filesort

## 建议动作
- 改用 **游标/keyset 分页**（`WHERE id > ? ORDER BY id LIMIT n`）
- 或延迟关联：先查主键 id 分页再 JOIN 明细

## 常见误判
- 认为加 LIMIT 就一定「快」
- 在无序大表上单独加 OFFSET 列索引而不改分页模式
