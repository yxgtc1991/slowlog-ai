package eval

import (
	"context"
	"testing"
)

func TestAllCases(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	for _, r := range RunAll(ctx, "") {
		if r.Pass {
			continue
		}
		t.Errorf("case %q failed: %v (actions=%v)", r.CaseName, r.Errors, r.Actions)
	}
}

func TestAssertReportFile_minimalFixture(t *testing.T) {
	t.Parallel()
	errs := AssertReportFile("testdata/eval/minimal-report.json", ReportExpect{
		MinRounds:     2,
		ActionTypes:   []string{"retrieve_rag", "finish"},
		FinalContains: []string{"全表扫描"},
	})
	if len(errs) > 0 {
		t.Fatalf("report assert: %v", errs)
	}
}
