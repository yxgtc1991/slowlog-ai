package mysql

import (
	"ai_slow_log/internal/config"
	"testing"
)

func TestValidateExplainSelect(t *testing.T) {
	t.Parallel()

	ok, err := validateExplainSelect("SELECT * FROM products WHERE id = 1")
	if err != nil || ok != "SELECT * FROM products WHERE id = 1" {
		t.Fatalf("ok query: %q err=%v", ok, err)
	}

	_, err = validateExplainSelect("DELETE FROM products")
	if err == nil {
		t.Fatal("expected error for DELETE")
	}

	_, err = validateExplainSelect("SELECT 1; SELECT 2")
	if err == nil {
		t.Fatal("expected error for multiple statements")
	}
}

func TestBuildAddIndexDDL(t *testing.T) {
	t.Parallel()

	c := &Client{cfg: config.MySQLConfig{Database: "test"}}
	ddl, err := c.BuildAddIndexDDL(AddIndexOptions{
		Table:     "products",
		IndexName: "idx_sku",
		Columns:   []string{"sku"},
	})
	if err != nil {
		t.Fatal(err)
	}
	want := "ALTER TABLE `test`.`products` ADD INDEX `idx_sku` (`sku`)"
	if ddl != want {
		t.Fatalf("ddl=%q want %q", ddl, want)
	}
}
