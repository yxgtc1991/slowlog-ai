package mysql

import (
	"context"
	"fmt"
	"strings"
)

// ExplainFormat EXPLAIN 输出格式。
type ExplainFormat string

const (
	ExplainFormatTraditional ExplainFormat = "traditional"
	ExplainFormatJSON        ExplainFormat = "json"
)

// Explain 对单条 SELECT 执行 EXPLAIN，返回结果行（便于 JSON 序列化给 LLM）。
func (c *Client) Explain(ctx context.Context, query string, format ExplainFormat) (map[string]interface{}, error) {
	normalized, err := validateExplainSelect(query)
	if err != nil {
		return nil, err
	}

	prefix := "EXPLAIN "
	switch format {
	case ExplainFormatJSON:
		prefix = "EXPLAIN FORMAT=JSON "
	case ExplainFormatTraditional, "":
		format = ExplainFormatTraditional
	default:
		return nil, fmt.Errorf("unsupported explain format: %s", format)
	}

	rows, err := c.db.QueryContext(ctx, prefix+normalized)
	if err != nil {
		return nil, fmt.Errorf("explain: %w", err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var resultRows []map[string]interface{}
	for rows.Next() {
		dest := make([]interface{}, len(cols))
		ptrs := make([]interface{}, len(cols))
		for i := range dest {
			ptrs[i] = &dest[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}
		row := make(map[string]interface{}, len(cols))
		for i, col := range cols {
			row[col] = scanValue(dest[i])
		}
		resultRows = append(resultRows, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"format":    string(format),
		"query":     normalized,
		"rows":      resultRows,
		"row_count": len(resultRows),
	}, nil
}

func scanValue(v interface{}) interface{} {
	switch x := v.(type) {
	case []byte:
		return string(x)
	default:
		return x
	}
}

// validateExplainSelect 仅允许单条 SELECT，降低注入风险。
func validateExplainSelect(query string) (string, error) {
	q := strings.TrimSpace(query)
	q = strings.TrimSuffix(q, ";")
	q = strings.TrimSpace(q)
	if q == "" {
		return "", fmt.Errorf("sql is empty")
	}

	upper := strings.ToUpper(q)
	if !strings.HasPrefix(upper, "SELECT") {
		return "", fmt.Errorf("only SELECT is allowed for EXPLAIN")
	}

	if strings.Contains(q, ";") {
		return "", fmt.Errorf("multiple statements are not allowed")
	}
	for _, forbidden := range []string{
		" INTO ", " FOR UPDATE", " LOCK IN SHARE MODE",
		"DROP ", "ALTER ", "CREATE ", "TRUNCATE ", "DELETE ", "INSERT ", "UPDATE ",
		" GRANT ", " REVOKE ", " EXEC ", " CALL ", " LOAD ",
	} {
		if strings.Contains(upper, forbidden) {
			return "", fmt.Errorf("query contains forbidden keyword")
		}
	}
	if strings.Contains(q, "/*") || strings.Contains(q, "--") {
		return "", fmt.Errorf("comments are not allowed in sql")
	}

	return q, nil
}
