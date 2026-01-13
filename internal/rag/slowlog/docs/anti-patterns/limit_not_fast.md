# Anti-Pattern: 有 LIMIT 就一定快

## 错误结论
“有 LIMIT 10，所以查询很快”

## 实际情况
- ORDER BY + LIMIT 仍可能全表扫描
- 未使用覆盖索引时仍需排序

## 纠正方式
- 判断是否 Using index / Using filesort
- 查看索引顺序是否匹配 ORDER BY