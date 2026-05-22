package analyzer

import (
	"strings"
	"testing"
)

func TestExtractSelectSQL_products(t *testing.T) {
	t.Parallel()
	slowLog := `# comment
SELECT *
FROM products
WHERE price >= 100
ORDER BY created_at DESC
LIMIT 20;`
	got := ExtractSelectSQL(slowLog)
	if !strings.Contains(got, "products") || !strings.Contains(got, "price") {
		t.Fatalf("got %q", got)
	}
}

func TestNormalizeExplainArgs_replacesOrders(t *testing.T) {
	t.Parallel()
	slowLog := "SELECT * FROM products WHERE price >= 100 LIMIT 20;"
	args := NormalizeExplainArgs(map[string]interface{}{
		"database": "test",
		"sql":      "SELECT * FROM orders WHERE user_id = 1",
	}, slowLog)
	sql, _ := args["sql"].(string)
	if !strings.Contains(sql, "products") {
		t.Fatalf("sql=%q", sql)
	}
}
