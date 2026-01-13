package prompt

import "fmt"

// BuildSlowLogPrompt 构建 v1 版本的慢日志分析 prompt
// 这是最基础的版本，仅包含简单的提示和慢日志内容
func BuildSlowLogPrompt(slowLog string) string {
	return fmt.Sprintf(`
你是一个 MySQL 专家，请分析以下慢日志并给出优化建议：

%s
`, slowLog)
}
