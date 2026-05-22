package prompt

import (
	"encoding/json"
	"strings"
)

// FlexString 解析 LLM 输出的文本字段：可为 JSON 字符串，或对象/数组（序列化为文本）。
type FlexString string

func (f *FlexString) UnmarshalJSON(data []byte) error {
	data = []byte(strings.TrimSpace(string(data)))
	if len(data) == 0 || string(data) == "null" {
		*f = ""
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*f = FlexString(s)
		return nil
	}
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	out, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	*f = FlexString(out)
	return nil
}

func (f FlexString) String() string {
	return string(f)
}
