package analyzer

import (
	"ai_slow_log/internal/llm"
	"ai_slow_log/internal/prompt"
	"context"
)

func AnalyzeSlowLog(ctx context.Context, slowLog string) (string, error) {
	p := prompt.BuildSlowLogPromptV2(slowLog)

	client, err := llm.NewDeepSeekClient()
	if err != nil {
		return "", err
	}

	return client.Chat(ctx, p)
}
