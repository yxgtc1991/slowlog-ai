// 校验 .env 中的 MySQL 配置是否可连接（不调用 LLM）。
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

	cap := &mcp.ConnectMySQLCapability{Client: client}
	out, err := cap.Execute(context.Background(), map[string]interface{}{})
	if err != nil {
		log.Fatalf("execute: %v", err)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		log.Fatalf("encode: %v", err)
	}
	fmt.Println()
}
