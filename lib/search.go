package paperless

type TokenType int

const (
	eof rune = -1
	TokError TokenType = iota
	TokCanceled
	TokEOF
	TokEmpty
	TokAnd
	TokOr
	TokNot
	TokString
	TokParOpen
	TokParClose
)

type Query struct {
}

type Token struct {
	Type TokenType
	Value string
	Pos   int
}

type stateFunc func(*lexer) stateFunc

type lexer struct {
	input string
	start int
	pos   int
	width int

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

	return 0
}

func lexText(l *lexer) stateFunc {
	return nil
}

func (l *lexer) run() {
	defer func() {
		// close(l.tokens)
		// close(l.cancel)
		// l.clear("")
		// l.initialized = false
		// l.wg.Done()
	}()
	for s := lexText; s != nil; {
		s = s(l)
	}
	close(l.tokens)
}
