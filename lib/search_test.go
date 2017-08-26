package paperless

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func Test_lexer_Init(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		output []TokenType
	}{
		{"Empty search", "", []TokenType{TokEOF}},
		{"Text search", "some text", []TokenType{TokString, TokString, TokEOF}},
		{"Text within parentheses", "(text)", []TokenType{TokParOpen, TokString, TokParClose, TokEOF}},
		{"Text with unbalanced parens", "(text", []TokenType{TokParOpen, TokString, TokError}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
				t.Error("List of tokens differs:\n", diffStr(tt.output, result),
					"Outputted list of tokens:\n", spew.Sdump(tokens))
			}
			l.Deinit()
		})
	}
}
