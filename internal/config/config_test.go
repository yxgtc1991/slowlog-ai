package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDotEnv_doesNotOverrideExisting(t *testing.T) {
	t.Setenv("MYSQL_PORT", "9999")
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	if err := os.WriteFile(path, []byte("MYSQL_PORT=3306\nMYSQL_USER=from_file\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := LoadDotEnv(path); err != nil {
		t.Fatal(err)
	}
	if os.Getenv("MYSQL_PORT") != "9999" {
		t.Fatalf("port=%s", os.Getenv("MYSQL_PORT"))
	}
	if os.Getenv("MYSQL_USER") != "from_file" {
		t.Fatalf("user=%s", os.Getenv("MYSQL_USER"))
	}
}

func TestLoadMySQLFromEnv_defaults(t *testing.T) {
	t.Setenv("MYSQL_HOST", "")
	t.Setenv("MYSQL_PORT", "")
	cfg := LoadMySQLFromEnv()
	if cfg.Host != "127.0.0.1" || cfg.Port != 3306 {
		t.Fatalf("cfg=%+v", cfg)
	}
}
