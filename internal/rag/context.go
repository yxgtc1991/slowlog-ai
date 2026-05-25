package rag

import (
	"context"
	"strings"
)

type slowLogCtxKey struct{}

// ContextWithSlowLog 将慢日志原文放入 context，供多路检索改写使用。
func ContextWithSlowLog(ctx context.Context, slowLog string) context.Context {
	if strings.TrimSpace(slowLog) == "" {
		return ctx
	}
	return context.WithValue(ctx, slowLogCtxKey{}, slowLog)
}

// SlowLogFromContext 读取慢日志 hint。
func SlowLogFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(slowLogCtxKey{}).(string); ok {
		return v
	}
	return ""
}
