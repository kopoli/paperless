package paperless

import "testing"

func Test_lexer_Init(t *testing.T) {
	tests := []struct {
		name   string
		input string
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &lexer{}
			l.Init(tt.input)
		})
	}
}
