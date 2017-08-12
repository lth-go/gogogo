package parse

import (
	"errors"
	"fmt"
	"unicode"
)

type TokenType int

const (
	ERROR       TokenType = iota // 0 ERROR
	EOF
	EOL                          // END OF LINE
	BOOL                         // 3 TRUE OF FALSE
	IDENTI                       // 4 变量
	NUMBER                       // 5 数字
	NIL                          // 36 NIL
	STRING                       // 6 字符串
	DOT                          // 8 .
	SPACE                        // 7 空格
	LP                           // 15 (
	RP                           // 16 )
	LC                           // 17 {
	RC                           // 18 }
	LB                           // 19 [
	RB                           // 20 ]
	SEMICOLON                    // 21 ;
	COLON                        // 22 :
	COMMA                        // 23 ,
	PLUS                         // 10 +
	MINUS                        // 11 -
	MULTIPLY                     // 12 *
	DIVIDE                       // 13 /
	ANDAND                       // &&
	AND                          // 25 &
	OROR                         // ||
	OR                           // 27 |
	EQ                           // 28 =
	EQEQ                         // 9 ==
	NEQ                          //!29 =
	GT                           // 30 >
	GE                           // 31 >=
	LT                           // 32 <
	LE                           // 33 <=
	EXCLAMATION                  // 34 !
	KEYWORD                      // 35 关键字分隔
	FUNC                         // 37 FUNC
	RETURN                       // 38 RETURN
	BREAK                        // 44 BREAK
	CONTINUE                     // 45 CONTINUE
	IF                           // 40 IF
	ELIF                         // 42 ELIF
	ELSE                         // 41 ELSE
	FOR                          // 39 FOR
)

var opName = map[string]TokenType{
	"func":     FUNC,
	"return":   RETURN,
	"break":    BREAK,
	"continue": CONTINUE,
	"if":       IF,
	"elif":     ELIF,
	"else":     ELSE,
	"for":      FOR,
	"true":     BOOL,
	"false":    BOOL,
	"nil":      NIL,
}

var symbolMap = map[rune]TokenType{
	'(': LP,
	')': RP,
	':': COLON,
	'{': LC,
	'}': RC,
	',': COMMA,
	';': SEMICOLON,
	'[': LB,
	']': RB,
	'+': PLUS,
	'-': MINUS,
	'*': MULTIPLY,
	'/': DIVIDE,
}

type Error struct {
	Message  string
	Pos      Position
	Filename string
	Fatal    bool
}

func (e *Error) Error() string {
	return e.Message
}

type token struct {
	PosImpl
	typ TokenType
	val string
}

type Scanner struct {
	src      []rune
	offset   int
	lineHead int
	line     int
}

func (s *Scanner) Scan() (typ TokenType, lit string, pos Position, err error) {
	s.skipBlank()
	pos = s.pos()
	switch ch := s.peek(); {
	case isLetter(ch):
		lit, err = s.scanIdentifier()
		if err != nil {
			return
		}
		if name, ok := opName[lit]; ok {
			typ = name
		} else {
			typ = IDENTI
		}
	case isDigit(ch):
		typ = NUMBER
		lit, err = s.scanNumber()
		if err != nil {
			return
		}
	case ch == '"':
		typ = STRING
		lit, err = s.scanString()
		if err != nil {
			return
		}
	default:
		switch ch {
		case -1:
			typ = EOF
		case '!':
			s.next()
			switch s.peek() {
			case '=':
				typ = NEQ
				lit = "!="
			default:
				s.back()
				typ = EXCLAMATION
				lit = string(ch)
			}
		case '=':
			s.next()
			switch s.peek() {
			case '=':
				typ = EQEQ
				lit = "=="
			default:
				s.back()
				typ = EQ
				lit = string(ch)
			}
		case '>':
			s.next()
			switch s.peek() {
			case '=':
				typ = GE
				lit = ">="
			default:
				s.back()
				typ = GT
				lit = string(ch)
			}
		case '<':
			s.next()
			switch s.peek() {
			case '=':
				typ = LE
				lit = "<="
			default:
				s.back()
				typ = LT
				lit = string(ch)
			}
		case '|':
			s.next()
			switch s.peek() {
			case '|':
				typ = OROR
				lit = "||"
			default:
				s.back()
				typ = OR
				lit = string(ch)
			}
		case '&':
			s.next()
			switch s.peek() {
			case '&':
				typ = ANDAND
				lit = "&&"
			default:
				s.back()
				typ = AND
				lit = string(ch)
			}
		case '\n':
			typ = EOL
			lit = "EOL"
		case ',', ':', ';', '(', ')', '{', '}', '[', ']', '+', '-', '*', '/':
			typ = symbolMap[ch]
			lit = string(ch)
		default:
			err = fmt.Errorf(`syntax error "%s"`, string(ch))
			typ = ERROR
			lit = string(ch)
			return
		}
		s.next()
	}
	return
}

//////////////////////////////
// 解析函数
//////////////////////////////
func (s *Scanner) scanIdentifier() (string, error) {
	var ret []rune
	for {
		if !isLetter(s.peek()) && !isDigit(s.peek()) {
			break
		}
		ret = append(ret, s.peek())
		s.next()
	}
	return string(ret), nil
}

func (s *Scanner) scanNumber() (string, error) {
	var ret []rune

	ch := s.peek()
	ret = append(ret, ch)
	s.next()

	if ch == '0' && s.peek() == '.' {
		ret = append(ret, ch)
		s.next()
	} else if ch == '0' && isDigit(s.peek()) {
		return "", errors.New("数字不能以0开头")
	}

	for isDigit(s.peek()) {
		ret = append(ret, s.peek())
		s.next()
	}

	if isLetter(s.peek()) {
		return "", errors.New("identifier starts immediately after numeric literal")
	}
	return string(ret), nil
}
func (s *Scanner) scanString() (string, error) {
	var ret []rune

eos:
	for {
		s.next()
		switch s.peek() {
		case '\n':
			return "", errors.New("Unexpected EOL")
		case -1:
			return "", errors.New("Unexpected EOF")
		case '"':
			s.next()
			break eos
		default:
			ret = append(ret, s.peek())
		}
	}
	return string(ret), nil
}

//////////////////////////////
// 位置
//////////////////////////////

func (s *Scanner) peek() rune {
	if s.reachEOF() {
		return -1
	}
	return s.src[s.offset]
}

func (s *Scanner) next() {
	if !s.reachEOF() {
		if s.peek() == '\n' {
			s.lineHead = s.offset + 1
			s.line++
		}
		s.offset++
	}
}

func (s *Scanner) current() int {
	return s.offset
}

func (s *Scanner) set(o int) {
	s.offset = o
}

func (s *Scanner) back() {
	s.offset--
}

func (s *Scanner) reachEOF() bool {
	return len(s.src) <= s.offset
}

func (s *Scanner) pos() Position {
	return Position{Line: s.line + 1, Column: s.offset - s.lineHead + 1}
}

func (s *Scanner) skipBlank() {
	for isBlank(s.peek()) {
		s.next()
	}
}

//////////////////////////////
// 判断字符
//////////////////////////////

func isLetter(r rune) bool {
	return unicode.IsLetter(r) || r == '_'
}

func isDigit(r rune) bool {
	return '0' <= r && r <= '9'
}

func isEOL(r rune) bool {
	return r == '\n' || r == -1
}

func isBlank(r rune) bool {
	return r == ' ' || r == '\t' || r == '\r'
}

//////////////////////////////
// lexer
//////////////////////////////

type lexer struct {
	s      *Scanner
	tokens chan token
}

func lex(src string) *lexer {
	s := &Scanner{src: []rune(src)}
	l := &lexer{s: s, tokens: make(chan token)}
	go l.run()

	return l
}

func (l *lexer) run() {
	for {
		tok, lit, pos, err := l.s.Scan()
		if err != nil {
			panic(fmt.Sprintf("%s", err.Error()))
		}
		t := token{typ: tok, val: lit}
		t.SetPosition(pos)

		l.tokens <- t

		if tok == EOF {
			break
		}
	}
	close(l.tokens)

}

func (l *lexer) nextToken() token {
	token := <-l.tokens
	return token
}
