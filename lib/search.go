package paperless

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

type TokenType int

const (
	eof      rune      = -1
	TokError TokenType = iota
	TokCanceled
	TokEOF
	TokEmpty
	TokAnd
	TokOr
	TokNot
	TokQuote
	TokString
	TokParOpen
	TokParClose
)

type Query struct {
}

type Token struct {
	Type  TokenType
	Value string
	Pos   int
}

type stateFunc func(*lexer) stateFunc

type lexer struct {
	input string
	start int
	pos   int
	width int

	states      []stateFunc
	initialized bool

	tokens chan Token

	// wg     sync.WaitGroup
	// cancel chan bool
}

// using functionality

// TODO This is in the wrong place
// EmitSQL
func (s *Query) EmitSQL() string {
	return ""
}

// Lexer public interface

// func (l *lexer) clear(input string) {
// 	l.input = input
// 	l.start = 0
// 	l.pos = 0
// }

func (l *lexer) Init(input string) {
	// if l.initialized {
	// 	l.cancel <- true
	// 	l.wg.Wait()
	// 	l.Deinit()
	// }
	l.Deinit()
	// l.clear(input)
	l.input = input
	l.start = 0
	l.pos = 0
	l.width = 0
	l.states = nil
	l.tokens = make(chan Token)
	// l.cancel = make(chan bool)
	// l.wg = sync.WaitGroup{}
	l.initialized = true

	// l.wg.Add(1)
	go l.run()
}

func (l *lexer) NextToken() (ret Token) {
	ret, ok := <-l.tokens
	if !ok {
		ret = Token{
			Type: TokEOF,
		}
	}
	return
}

func (l *lexer) Deinit() {
	if l.initialized {
		for t := l.NextToken(); t.Type != TokEOF; {
		}
		l.initialized = false
	}
}

func (l *lexer) emit(t TokenType) {
	l.tokens <- Token{t, l.input[l.start:l.pos], l.start}
	l.start = l.pos
}

func (l *lexer) next() rune {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}

	var ret rune
	ret, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	return ret
}

func (l *lexer) rewind() {
	l.pos -= l.width
}

func (l *lexer) ignore() {
	l.start = l.pos
}

func (l *lexer) peek() rune {
	ret := l.next()
	l.rewind()
	return ret
}

func (l *lexer) hasPrefix(prefix string) bool {
	if strings.HasPrefix(l.input[l.pos:], prefix) {
		// l.pos += len(prefix)
		return true
	}
	return false
}

func (l *lexer) acceptRune(charclass string) bool {
	// TODO
	return false
}

func (l *lexer) errorf(format string, args ...interface{}) stateFunc {
	l.tokens <- Token{TokError, fmt.Sprintf(format, args...), l.pos}
	return nil
}

func (l *lexer) push(s stateFunc) {
	l.states = append(l.states, s)
}

func (l *lexer) pop() stateFunc {
	if l.states == nil || len(l.states) == 0 {
		return l.errorf("Internal error: Popping an empty stack")
	}

	ret := l.states[len(l.states)-1]
	l.states = l.states[:len(l.states)-1]
	return ret
}

func lexTop(l *lexer) stateFunc {
	for {
		r := l.next()
		if r == eof {
			break
		}
		switch {
		case unicode.IsSpace(r):
			l.ignore()
		case r == '"':
			l.push(lexTop)
			return lexQuoted
		}
	}

	if l.pos > l.start {
		l.errorf("Internal error: Following not handled: %s", l.input[l.start:l.pos])
	}

	l.emit(TokEOF)
	return nil
}

func lexString(l *lexer) stateFunc {
	// r := l.next()

	if l.hasPrefix("AND ") {
		l.emit(TokString)
		
	}

	r := l.next()
	switch r {
	case '"':
		l.push(lexString)
		return lexQuoted
	}

	return lexString
}

func lexQuoted(l *lexer) stateFunc {
	r := l.next()

	switch {
	case r == '"':
		l.emit(TokString)
		return l.pop()
	case r == eof:
		return l.errorf("Unbalanced quotes")
	}
	return lexQuoted
}

func (l *lexer) run() {
	// defer func() {
	// close(l.tokens)
	// close(l.cancel)
	// l.clear("")
	// l.initialized = false
	// l.wg.Done()
	// }()
	for s := lexTop; s != nil; {
		s = s(l)
	}
	close(l.tokens)
}
