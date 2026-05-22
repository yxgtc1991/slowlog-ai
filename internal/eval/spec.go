package eval

// TrajectoryStep 期望的轨迹一步（type 必填；call_tool 时可校验 tool_name）。
type TrajectoryStep struct {
	Type     string
	ToolName string
}

// Expect 对一次 Agent 运行的断言。
type Expect struct {
	Trajectory       []TrajectoryStep // 按序匹配 action 序列（可只写关键步）
	FinalContains    []string         // 最终结论须包含的子串（任一大小写不敏感）
	FinalNotContains []string
	MinIterations    int
	MaxIterations    int
	ToolsMustCall    []string // 轨迹中须出现过的 call_tool（顺序不限）
	NoActionErrors   bool     // 每轮 ActionError 为空
}

// Case 一条 golden eval：脚本化 LLM + 断言。
type Case struct {
	Name        string
	SlowLogPath string
	Guide       bool
	Script      []string
	Executor    *StubExecutor
	Expect      Expect
}

// Result 单条 case 运行结果。
type Result struct {
	CaseName   string
	Pass       bool
	Errors     []string
	Iterations int
	Actions    []string // 人类可读：type 或 call_tool:name
	Final      string
}
