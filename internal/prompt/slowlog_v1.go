package prompt

import "fmt"

func BuildSlowLogPrompt(slowLog string) string {
	return fmt.Sprintf(`
你是一个 MySQL 专家，请分析以下慢日志并给出优化建议：

%s
`, slowLog)
}
