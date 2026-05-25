package agentmode

import (
	"fmt"
	"os"
	"strings"
)

// Mode Agent 运行模式：V6 NextAction（默认）或 V5 API Tool Calling。
type Mode string

const (
	V6 Mode = "v6"
	V5 Mode = "v5"
)

// Default 默认 V6。
func Default() Mode { return V6 }

// Resolve 从环境变量与命令行解析模式，并返回去掉模式参数后的 args。
// 支持：SLOWLOG_AGENT_MODE=v5|v6、-agent-mode=v5、--agent-mode v6、-mode=v5
func Resolve(envValue string, args []string) (Mode, []string, error) {
	mode := strings.ToLower(strings.TrimSpace(envValue))
	var rest []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if m, ok := parseModeArg(arg); ok {
			mode = string(m)
			continue
		}
		if arg == "--agent-mode" || arg == "-agent-mode" {
			if i+1 >= len(args) {
				return "", nil, fmt.Errorf("missing value after %s", arg)
			}
			mode = strings.ToLower(strings.TrimSpace(args[i+1]))
			i++
			continue
		}
		rest = append(rest, arg)
	}
	if mode == "" {
		return V6, rest, nil
	}
	switch Mode(mode) {
	case V6, V5:
		return Mode(mode), rest, nil
	default:
		return "", rest, fmt.Errorf("unknown agent mode %q (use v5 or v6)", mode)
	}
}

// ResolveFromEnv 读取 SLOWLOG_AGENT_MODE 并解析 os.Args[1:]。
func ResolveFromEnv() (Mode, []string, error) {
	return Resolve(os.Getenv("SLOWLOG_AGENT_MODE"), os.Args[1:])
}

func parseModeArg(arg string) (Mode, bool) {
	for _, p := range []string{"-agent-mode=", "-mode=", "--agent-mode="} {
		if strings.HasPrefix(arg, p) {
			return Mode(strings.ToLower(strings.TrimPrefix(arg, p))), true
		}
	}
	return "", false
}
