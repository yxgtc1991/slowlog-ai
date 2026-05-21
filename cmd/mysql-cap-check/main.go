// 校验 explain_mysql_query 与 add_mysql_index（add_index 固定 dry_run，不改表）。
package main

import (
	"ai_slow_log/internal/config"
	"ai_slow_log/internal/mcp"
	"ai_slow_log/internal/mysql"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
)

func main() {
	cfg, err := config.MustLoadMySQL()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	client, err := mysql.NewClient(cfg)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")

	database := "test"
	if cfg.Database != "" {
		database = cfg.Database
	}

	fmt.Println("=== connect_mysql_instance ===")
	connectOut, err := (&mcp.ConnectMySQLCapability{Client: client}).Execute(ctx, map[string]interface{}{
		"database": database,
	})
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	_ = enc.Encode(connectOut)
	fmt.Println()

	sql := "SELECT * FROM products LIMIT 5"
	if len(os.Args) > 1 {
		sql = os.Args[1]
	}

	fmt.Println("=== explain_mysql_query ===")
	explainOut, err := (&mcp.ExplainMySQLQueryCapability{Client: client}).Execute(ctx, map[string]interface{}{
		"database": database,
		"sql":      sql,
	})
	if err != nil {
		log.Fatalf("explain: %v", err)
	}
	_ = enc.Encode(explainOut)
	fmt.Println()

	fmt.Println("=== add_mysql_index (dry_run=true, 不执行 DDL) ===")
	indexOut, err := (&mcp.AddMySQLIndexCapability{Client: client}).Execute(ctx, map[string]interface{}{
		"database":   database,
		"table":      "products",
		"index_name": "idx_verify_dry_run",
		"columns":    "id",
		"dry_run":    true,
	})
	if err != nil {
		log.Fatalf("add_index: %v", err)
	}
	_ = enc.Encode(indexOut)
	fmt.Println()
}
