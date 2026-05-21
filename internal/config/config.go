package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// MySQLConfig 本地 / 环境变量中的 MySQL 实例连接配置。
// 凭证只从环境变量或 .env 读取，不要写死在代码里。
type MySQLConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
}

// LoadMySQLFromEnv 从环境变量加载 MySQL 配置。
func LoadMySQLFromEnv() MySQLConfig {
	port, _ := strconv.Atoi(getEnv("MYSQL_PORT", "3306"))
	return MySQLConfig{
		Host:     getEnv("MYSQL_HOST", "127.0.0.1"),
		Port:     port,
		User:     getEnv("MYSQL_USER", "root"),
		Password: os.Getenv("MYSQL_PASSWORD"),
		Database: getEnv("MYSQL_DATABASE", ""),
	}
}

// LoadDotEnv 从 .env 加载 KEY=VALUE（不覆盖已存在的环境变量）。
func LoadDotEnv(path string) error {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		value = strings.Trim(value, `"'`)
		if key == "" {
			continue
		}
		if os.Getenv(key) == "" {
			_ = os.Setenv(key, value)
		}
	}
	return scanner.Err()
}

// MustLoadMySQL 加载 .env（若存在）并返回 MySQL 配置；密码缺失时返回 error。
func MustLoadMySQL() (MySQLConfig, error) {
	if err := LoadDotEnv(".env"); err != nil {
		return MySQLConfig{}, fmt.Errorf("load .env: %w", err)
	}
	cfg := LoadMySQLFromEnv()
	if cfg.Password == "" {
		return MySQLConfig{}, fmt.Errorf("MYSQL_PASSWORD is required (set in .env or environment)")
	}
	if cfg.User == "" {
		return MySQLConfig{}, fmt.Errorf("MYSQL_USER is required")
	}
	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
