package analyzer

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// UserInputReader 人机协同：Agent 提问后读取用户输入。
type UserInputReader interface {
	ReadLine(prompt string) (string, error)
}

// HITLEnabledFromEnv 是否开启 HITL（SLOWLOG_AGENT_HITL=1|true|yes|on）。
func HITLEnabledFromEnv() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("SLOWLOG_AGENT_HITL"))) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

type stdinReader struct {
	in io.Reader
}

func (r stdinReader) ReadLine(prompt string) (string, error) {
	if prompt != "" {
		fmt.Fprint(os.Stderr, prompt)
	}
	sc := bufio.NewScanner(r.in)
	if sc.Scan() {
		return strings.TrimSpace(sc.Text()), sc.Err()
	}
	if err := sc.Err(); err != nil {
		return "", err
	}
	return "", nil
}

// DefaultStdinReader 从 os.Stdin 读取（用于 CLI）。
func DefaultStdinReader() UserInputReader {
	return stdinReader{in: os.Stdin}
}

// StubUserInput 固定答案序列（用于 eval / 单测）。
type StubUserInput struct {
	answers []string
	n       int
}

func NewStubUserInput(answers ...string) *StubUserInput {
	return &StubUserInput{answers: answers}
}

func (s *StubUserInput) ReadLine(string) (string, error) {
	if s.n >= len(s.answers) {
		return "", nil
	}
	a := s.answers[s.n]
	s.n++
	return a, nil
}
