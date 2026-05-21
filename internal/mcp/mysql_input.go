package mcp

import (
	"fmt"
	"strings"
)

func stringInput(input map[string]interface{}, key string) (string, bool) {
	v, ok := input[key].(string)
	if !ok {
		return "", false
	}
	v = strings.TrimSpace(v)
	return v, v != ""
}

func boolInput(input map[string]interface{}, key string, defaultVal bool) bool {
	v, ok := input[key]
	if !ok {
		return defaultVal
	}
	switch x := v.(type) {
	case bool:
		return x
	case string:
		return strings.EqualFold(x, "true") || x == "1"
	default:
		return defaultVal
	}
}

func parseColumns(input map[string]interface{}) ([]string, error) {
	if raw, ok := input["columns"]; ok {
		switch v := raw.(type) {
		case string:
			s := strings.TrimSpace(v)
			if s == "" {
				return nil, fmt.Errorf("columns is empty")
			}
			parts := strings.Split(s, ",")
			out := make([]string, 0, len(parts))
			for _, p := range parts {
				p = strings.TrimSpace(p)
				if p == "" {
					continue
				}
				out = append(out, p)
			}
			if len(out) == 0 {
				return nil, fmt.Errorf("columns is empty")
			}
			return out, nil
		case []interface{}:
			out := make([]string, 0, len(v))
			for _, item := range v {
				s, ok := item.(string)
				if !ok || strings.TrimSpace(s) == "" {
					return nil, fmt.Errorf("columns must be strings")
				}
				out = append(out, strings.TrimSpace(s))
			}
			if len(out) == 0 {
				return nil, fmt.Errorf("columns is empty")
			}
			return out, nil
		case []string:
			if len(v) == 0 {
				return nil, fmt.Errorf("columns is empty")
			}
			return v, nil
		}
	}
	return nil, fmt.Errorf("columns is required")
}
