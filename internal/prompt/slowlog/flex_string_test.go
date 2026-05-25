package prompt

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestFlexString_unmarshalString(t *testing.T) {
	t.Parallel()
	var f FlexString
	if err := json.Unmarshal([]byte(`"hello"`), &f); err != nil {
		t.Fatal(err)
	}
	if f.String() != "hello" {
		t.Fatalf("got %q", f)
	}
}

func TestFlexString_unmarshalObject(t *testing.T) {
	t.Parallel()
	var f FlexString
	raw := `{"diagnosis":"全表扫描","risk":"写入略增"}`
	if err := json.Unmarshal([]byte(raw), &f); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(f.String(), "全表扫描") {
		t.Fatalf("got %s", f)
	}
}
