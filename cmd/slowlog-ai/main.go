package main

import (
	"ai_slow_log/internal/analyzer"
	"context"
	"fmt"
)

func main() {

	slowLog := `
# Time: 2025-01-05T10:12:33.123456Z
# User@Host: app_user[app_user] @ 10.0.0.12 []
# Query_time: 12.456  Lock_time: 0.001
# Rows_sent: 10  Rows_examined: 1250000
SET timestamp=1736071953;
SELECT *
FROM orders
WHERE user_id = 123
ORDER BY created_at DESC
LIMIT 10;
`

	result, err := analyzer.AnalyzeSlowLog(context.Background(), slowLog)
	if err != nil {
		panic(err)
	}

	fmt.Println("output =>", result)

}
