package paperless

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
)

type pstr map[int]string

func Test_lexer_Init(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		output []TokenType
		posstr pstr
	}{
		{"Empty search", "",
			[]TokenType{TokEOF}, nil},
		{"Text search", "some text",
			[]TokenType{TokString, TokString, TokEOF}, pstr{0: "some", 1: "text"}},
		{"Text within parentheses", "(text)",
			[]TokenType{TokParOpen, TokString, TokParClose, TokEOF}, pstr{1: "text"}},
		{"Text with unbalanced parens", "(text",
			[]TokenType{TokParOpen, TokString, TokError, TokEOF},
			pstr{1: "text"}},
		{"Multiple parens", "(())",
			[]TokenType{TokParOpen, TokParOpen, TokParClose, TokParClose, TokEOF},
			pstr{0: "("}},
		{"Multiple parens 2", "() ()",
			[]TokenType{TokParOpen, TokParClose, TokParOpen, TokParClose, TokEOF}, nil},
		{"Special operators", "a AND b OR c",
			[]TokenType{TokString, TokAnd, TokString, TokOr, TokString, TokEOF},
			pstr{0: "a", 4: "c"}},
		{"Just AND", "AND",
			[]TokenType{TokAnd, TokEOF}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			failed := false
			l := &lexer{}
			l.Init(tt.input)

			// var result []TokenType
			var tokens []Token
			result := make([]TokenType, 0, len(tt.output))
			for {
				t := l.NextToken()
				tokens = append(tokens, t)
				result = append(result, t.Type)
				if t.Type == TokEOF {
					break
				}
			}

			if !structEquals(tt.output, result) {
				t.Error("List of tokens differs:\n", diffStr(tt.output, result))
				failed = true
			}
			for i, val := range tt.posstr {
				if i >= len(tokens) {
					t.Error("List should contain token in position", i, "with value", val)
					failed = true
					continue
				}

				if tokens[i].Value != val {
					t.Error("Token", i, "Expected:", val, "Got:", tokens[i].Value)
					failed = true
					continue
				}
			}

			if failed {
				t.Log("Outputted list of tokens:\n", spew.Sdump(tokens))
			}
			l.Deinit()
		})
	}
}
