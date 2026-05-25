package toolerr

import (
	"errors"
	"testing"
)

func TestClassify_tableNotFound(t *testing.T) {
	t.Parallel()
	te := Classify("explain_mysql_query",
		errors.New(`explain: Error 1146 (42S02): Table 'test.orders' doesn't exist`))
	if te.Code != CodeMySQLTableNotFound || te.Retryable {
		t.Fatalf("%+v", te)
	}
}

func TestClassify_connectionRetryable(t *testing.T) {
	t.Parallel()
	te := Classify("connect_mysql_instance", errors.New("connection refused"))
	if te.Code != CodeMySQLConnection || !te.Retryable {
		t.Fatalf("%+v", te)
	}
}

func TestFrom_codedError(t *testing.T) {
	t.Parallel()
	te := From("add_mysql_index", New("add_mysql_index", CodeInvalidInput, "table is required", false))
	if te.Code != CodeInvalidInput || te.Retryable {
		t.Fatalf("%+v", te)
	}
}
