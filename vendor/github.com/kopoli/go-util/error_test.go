package util

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/pmezard/go-difflib/difflib"
)

func structEquals(a, b interface{}) bool {
	return spew.Sdump(a) == spew.Sdump(b)
}

func diffStr(a, b interface{}) (ret string) {
	diff := difflib.UnifiedDiff{
		A:        difflib.SplitLines(spew.Sdump(a)),
		B:        difflib.SplitLines(spew.Sdump(b)),
		FromFile: "Expected",
		ToFile:   "Received",
		Context:  3,
	}

	ret, _ = difflib.GetUnifiedDiffString(diff)
	return
}

type testErr struct {
	msg string
}

func (t *testErr) Error() string {
	return t.msg
}

type testOp interface {
	run(*ErrorList)
}

type testFunc func(*ErrorList)

func (t testFunc) run(d *ErrorList) {
	t(d)
}

func TestErrorList(t *testing.T) {
	ae := func(errmsg string) testFunc {
		return func(e *ErrorList) {
			e.Append(&testErr{errmsg})
		}
	}
	tests := []struct {
		name    string
		message string
		ops     []testOp
		result  string
		isEmpty bool
	}{
		{"Empty list", "empty", []testOp{}, "", true},
		{"One error", "one", []testOp{ae("a")}, "Error: one: Error 1: a; ", false},
		{"Two errors", "two", []testOp{ae("a"), ae("b")}, "Error: two: Error 1: a; Error 2: b; ", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			el := NewErrorList(tt.message)

			for _, op := range tt.ops {
				op.run(el)
			}

			gotRet := el.Error()
			if !structEquals(gotRet, tt.result) {
				t.Errorf("Expected error message differs:\n %s", diffStr(gotRet, tt.result))
			}

			if el.IsEmpty() != tt.isEmpty {
				t.Errorf("Expected to be empty: \"%v\" Reported empty: \"%v\"", tt.isEmpty, el.IsEmpty())
			}
		})
	}
}
