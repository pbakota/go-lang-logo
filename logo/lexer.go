package logo

import (
	"fmt"
	"log"
	"strings"
)

type Lexer struct {
	Expr          []rune
	Position      int
	Line          uint32
	String        string
	Number        int
	Literal       rune
	Debug         bool
	CommentSymbol rune
}

type Token int

const (
	TkEOF     Token = iota // 0 End of file
	TkLiteral Token = iota // 1 Literal token
	TkIdent   Token = iota // 2 Identifier
	TkString  Token = iota // 3 String
	TkNumber  Token = iota // 4 Number (decimal)
	TkComment Token = iota // 5 Comment
	TkEOL     Token = iota // 6 End of line
)

func NewLexer(expr string) *Lexer {
	return &Lexer{
		Expr:          []rune(expr),
		Position:      0,
		Line:          1,
		String:        "",
		Number:        0,
		Literal:       0,
		Debug:         false,
		CommentSymbol: '#',
	}
}

func (l *Lexer) dbg(fmt string, args ...any) {
	if l.Debug {
		log.Printf(fmt, args...)
	}
}

func (l *Lexer) isEof() bool {
	return l.Position >= len(l.Expr)
}

func (l *Lexer) isWhiteSpace() bool {
	return !l.isEof() && (l.Expr[l.Position] == ' ' || l.Expr[l.Position] == '\t')
}

func (l *Lexer) isEol() bool {
	return !l.isEof() && l.Expr[l.Position] == '\n'
}

func (l *Lexer) isQuote() bool {
	return !l.isEof() && (l.Expr[l.Position] == '"' || l.Expr[l.Position] == '\'')
}

// The map for literals is not complete, feel free to add yours
var LITERALS = map[rune]bool{
	'!': true,
	'@': true,
	'#': true,
	'$': true,
	'%': true,
	'^': true,
	'&': true,
	'*': true,
	'(': true,
	')': true,
	'[': true,
	']': true,
	'{': true,
	'}': true,
	':': true,
	';': true,
	'.': true,
	',': true,
	'<': true,
	'>': true,
	'/': true,
	'?': true,
	'+': true,
	'-': true,
}

func (l *Lexer) isLiteral() bool {
	if !l.isEof() {
		_, ok := LITERALS[l.Expr[l.Position]]
		return ok
	}

	return false
}

func (l *Lexer) isNumeric() bool {
	return !l.isEof() && (l.Expr[l.Position] >= '0' && l.Expr[l.Position] <= '9')
}

func (l *Lexer) isAlpha() bool {
	return !l.isEof() && ((l.Expr[l.Position] >= 'A' && l.Expr[l.Position] <= 'Z') || (l.Expr[l.Position] >= 'a' && l.Expr[l.Position] <= 'z'))
}

func (l *Lexer) NextToken() (Token, error) {
	l.String = ""
	l.Number = 0
	l.Literal = 0
	var sb strings.Builder

	if l.isEol() {
		l.dbg("End of line at position %d", l.Position)

		l.Position += 1
		l.Line += 1
		l.Literal = '\n'
		return TkEOL, nil
	}

	for l.isWhiteSpace() {
		l.Position += 1
	}

	if l.isEof() {
		return TkEOF, nil
	}

	if l.Expr[l.Position] == l.CommentSymbol {
		l.dbg("Found comment at position %d", l.Position)

		for !l.isEof() && !l.isEol() {
			sb.WriteRune(l.Expr[l.Position])
			l.Position += 1
		}
		l.String = sb.String()
		return TkComment, nil
	}

	if l.isLiteral() {
		l.dbg("Literal 0x%02x at position %d", int(l.Expr[l.Position]), l.Position)

		l.Literal = l.Expr[l.Position]
		l.Position += 1
		return TkLiteral, nil
	}

	if l.isQuote() {
		l.dbg("Found string literal at position %d", l.Position)

		l.Position += 1
		for !l.isEof() && !l.isQuote() {
			sb.WriteRune(l.Expr[l.Position])
			l.Position += 1
		}
		l.Position += 1
		l.String = sb.String()
		return TkString, nil
	}

	if l.isNumeric() {
		l.dbg("Found number at position %d", l.Position)

		number := 0
		for l.isNumeric() {
			number = number*10 + int(l.Expr[l.Position]-'0')
			l.Position += 1
		}
		if !l.isWhiteSpace() && !l.isEol() && !l.isEof() {
			return TkEOF, fmt.Errorf("parse error at line %d", l.Line)
		}
		l.Number = number
		return TkNumber, nil
	}

	if l.isAlpha() {
		l.dbg("Found identifier at position %d", l.Position)

		for !l.isWhiteSpace() && !l.isLiteral() && !l.isEol() && !l.isEof() {
			sb.WriteRune(l.Expr[l.Position])
			l.Position += 1
		}

		if sb.Len() != 0 {
			l.String = sb.String()
			return TkIdent, nil
		}
	}

	return TkEOF, fmt.Errorf("unknown character 0x%02x at line %d", l.Expr[l.Position], l.Line)
}
