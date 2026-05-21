package mysql

import (
	"fmt"
	"strings"
)

// validateIdent 校验 MySQL 标识符（库名、表名、索引名、列名）。
func validateIdent(name, label string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("%s is empty", label)
	}
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			continue
		}
		return fmt.Errorf("invalid %s: %s", label, name)
	}
	return nil
}

func validateIdents(label string, names ...string) error {
	for _, n := range names {
		if err := validateIdent(n, label); err != nil {
			return err
		}
	}
	return nil
}
