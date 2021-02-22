// Package lexer provides a lexer for PlanOut scripts.
package lexer

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/stitchfix/planout-golang/compiler/token"
)

const (
	eof      = -1
	alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digits   = "0123456789"
)

type Lexer struct {
	input  string           // string to scan
	state  stateFn          // current state
	start  int              // start position of candidate token
	line   int              // line number of candidate token
	pos    int              // current position of the input
	width  int              // width of last rune read
	tokens chan token.Token // buffer for tokens
}

// New returns a new Lexer for lexing a PlanOut script.
func New(input string) *Lexer {
	s := &Lexer{
		input:  input,
		state:  lexCode,
		tokens: make(chan token.Token, len(input)), // buffer is same size as input to avoid deadlock
		line:   1,
	}
	return s
}

// NextToken is used to iterate over tokens in the PlanOut script.
// Callers should stop iterating when receiving a token.ERROR or token.EOF token type.
func (lx *Lexer) NextToken() token.Token {
Loop:
	for {
		select {
		case tok := <-lx.tokens:
			return tok
		default:
			if lx.state == nil {
				break Loop
			}
			lx.state = lx.state(lx)
		}
	}
	return token.Token{Type: token.EOF, Line: lx.line}
}

type stateFn func(*Lexer) stateFn

func lexCode(lx *Lexer) stateFn {
	for {
		r := lx.next()
		switch {
		default:
			return lx.errorf("unexpected character: '%s'", lx.input[lx.start:lx.pos])
		case isSpace(r):
			lx.ignore()
		case isEndOfLine(r):
			lx.ignore()
			lx.line += 1
		case unicode.IsLetter(r):
			lx.backup()
			return lexIdentifier
		case unicode.IsDigit(r):
			lx.backup()
			return lexNumber
		case r == '\'' || r == '"':
			return lexString(r)
		case r == '=':
			return lexEquals
		case r == '<':
			return lexLessThan
		case r == '>':
			return lexGreaterThan
		case r == '|':
			return lexPipe
		case r == '&':
			return lexAmpersand
		case r == '?':
			return lexQuestionMark
		case r == '!':
			return lexBang
		case r == '#':
			return lexComment
		case r == '+':
			lx.emit(token.ADD)
		case r == '-':
			lx.emit(token.SUB)
		case r == '%':
			lx.emit(token.REM)
		case r == '*':
			lx.emit(token.MUL)
		case r == '/':
			lx.emit(token.QUO)
		case r == ':':
			lx.emit(token.COLON)
		case r == '@':
			return lexJSON
		case r == '[':
			lx.emit(token.LBRACKET)
		case r == ']':
			lx.emit(token.RBRACKET)
		case r == '(':
			lx.emit(token.LPAREN)
		case r == ')':
			lx.emit(token.RPAREN)
		case r == '{':
			lx.emit(token.LBRACE)
		case r == '}':
			lx.emit(token.RBRACE)
		case r == ',':
			lx.emit(token.COMMA)
		case r == ';':
			lx.emit(token.SEMICOLON)
		case r == eof:
			return nil
		}
	}
}

func (lx *Lexer) errorf(format string, args ...interface{}) stateFn {
	lx.tokens <- token.Token{
		Type: token.ERROR,
		Val:  fmt.Sprintf(format, args...),
		Line: lx.line,
	}
	return nil
}

// lexString scans until it finds a closeQuote, then emits a STRING token
// emits an error if a new line or EOF is encountered before closeQuote is found
func lexString(closeQuote rune) stateFn {
	return func(lx *Lexer) stateFn {
		lx.ignore() // ignore the open quote

		for {
			switch r := lx.next(); {
			case r == eof:
				return lx.errorf("EOF while scanning string")
			case isEndOfLine(r):
				return lx.errorf("new line while scanning string")
			case r == closeQuote:
				lx.backup() // don't include the close quote in emitted string
				lx.emit(token.STRING)

				lx.next() // now skip the close quote
				lx.ignore()
				return lexCode
			}
		}
	}
}

// lexComment scans until a new line OR EOF is found, then ignores everything
func lexComment(lx *Lexer) stateFn {
	for {
		switch r := lx.next(); {
		case r == eof:
			lx.ignore()
			return nil
		case isEndOfLine(r):
			lx.ignore()
			lx.line += 1
			return lexCode
		}
	}
}

// lexIdentifier accepts a run of alphanumeric characters, including underscores
func lexIdentifier(lx *Lexer) stateFn {
	lx.accept(alphabet)                         // first character must be a letter
	lx.acceptRun(alphabet + digits + "_" + ".") // after that accept letters, numbers, or underscores

	tokType := lx.keywordOrIdent()

	lx.emit(tokType)

	return lexCode
}

// lexEquals accepts a bare equals sign, double equal sign, or '=>' and then emits the appropriate token
func lexEquals(lx *Lexer) stateFn {
	if lx.accept("=") {
		lx.emit(token.EQUALS)
		return lexCode
	}

	if lx.accept(">") {
		lx.emit(token.THEN)
		return lexCode
	}

	lx.emit(token.ASSIGN)
	return lexCode
}

// lexJSON decodes the next JSON object from the input stream
// emits an error if the decoding fails
func lexJSON(lx *Lexer) stateFn {

	lx.ignore() // ignore the @ symbol

	dec := json.NewDecoder(strings.NewReader(lx.input[lx.start:]))

	var jsonLit interface{}
	if err := dec.Decode(&jsonLit); err != nil {
		lx.errorf("failed to parse JSON literal: %s", err.Error())
	}

	offset := dec.InputOffset()
	lx.pos += int(offset)

	lx.emit(token.JSON)

	return lexCode
}

// lexNumber accepts integers or floats, including scientific notation, then emits a NUMBER token
func lexNumber(lx *Lexer) stateFn {
	lx.acceptRun(digits)
	if lx.accept(".") {
		lx.acceptRun(digits)
	}
	if lx.accept("eE") {
		lx.accept("+-")
		lx.acceptRun(digits)
	}

	// At this point we've scanned the entire possible number so if the next rune
	// is alphanumeric the number is invalid
	if isAlphaNumeric(lx.peek()) {
		lx.next()
		return lx.errorf("bad number syntax: \"%s\"", lx.input[lx.start:lx.pos])
	}

	lx.emit(token.NUMBER)
	return lexCode
}

// lexLessThan accepts '<', '<=' or '<-' and emits the appropriate token
func lexLessThan(lx *Lexer) stateFn {
	if lx.accept("=") {
		lx.emit(token.LTE)
		return lexCode
	}

	if lx.accept("-") {
		lx.emit(token.ARROW)
		return lexCode
	}

	lx.emit(token.LSS)
	return lexCode
}

// lexGreaterThan accepts '>' or '>=' and emits the appropriate token
func lexGreaterThan(lx *Lexer) stateFn {
	if lx.accept("=") {
		lx.emit(token.GTE)
		return lexCode
	}

	lx.emit(token.GTR)
	return lexCode
}

// lexPipe accepts '||' and emits an OR token, or else emits an error
func lexPipe(lx *Lexer) stateFn {
	if lx.accept("|") {
		lx.emit(token.OR)
		return lexCode
	}

	return lx.errorf("invalid token: \"|\" (use \"||\" for OR)")
}

// lexAmpersand accepts '&&' and emits an AND token, or else emits an error
func lexAmpersand(lx *Lexer) stateFn {
	if lx.accept("&") {
		lx.emit(token.AND)
		return lexCode
	}

	return lx.errorf("invalid token: \"&\" (use \"&&\" for AND)")
}

// lexQuestionMark accepts '??' and emits a COALESCE token, or else emits an error
func lexQuestionMark(lx *Lexer) stateFn {
	if lx.accept("?") {
		lx.emit(token.COALESCE)
		return lexCode
	}

	return lx.errorf("invalid token: \"?\" (use \"??\" for COALESCE)")
}

// lexBang accepts '!' or '!=' and emits the appropriate token
func lexBang(lx *Lexer) stateFn {
	if lx.accept("=") {
		lx.emit(token.NEQ)
		return lexCode
	}

	lx.emit(token.NOT)
	return lexCode
}

// accept consumes the next rune if it's from the valid set
func (lx *Lexer) accept(valid string) bool {
	if strings.IndexRune(valid, lx.next()) >= 0 {
		return true
	}
	lx.backup()
	return false
}

// acceptRun consumes a run of runes from the valid set
func (lx *Lexer) acceptRun(valid string) {
	for strings.IndexRune(valid, lx.next()) >= 0 {
	}
	lx.backup()
}

// emit sends a token to the tokens channel with the specified type and value extracted from
// the input string. If given an identifier token, will check if the
func (lx *Lexer) emit(t token.Type) {
	lx.tokens <- token.Token{
		Type: t,
		Val:  lx.input[lx.start:lx.pos],
		Line: lx.line,
	}
	lx.start = lx.pos
}

// keywordOrIdent returns the correct keyword token if the candidate token is a keyword, otherwise returns IDENT
func (lx *Lexer) keywordOrIdent() token.Type {
	return token.Lookup(lx.input[lx.start:lx.pos])
}

// next consumes the next rune in the input
func (lx *Lexer) next() (r rune) {
	if lx.pos >= len(lx.input) {
		lx.width = 0
		return eof
	}
	r, lx.width = utf8.DecodeRuneInString(lx.input[lx.pos:])
	lx.pos += lx.width
	return r
}

// ignore skips over the pending input before this point
func (lx *Lexer) ignore() {
	lx.start = lx.pos
}

// backup steps back one rune
// can be called only once per call of next
func (lx *Lexer) backup() {
	lx.pos -= lx.width
}

// peek checks the next rune without consuming it
func (lx *Lexer) peek() rune {
	r := lx.next()
	lx.backup()
	return r
}

func isAlphaNumeric(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

func isSpace(r rune) bool {
	return r == ' ' || r == '\t'
}

func isEndOfLine(r rune) bool {
	return r == '\r' || r == '\n'
}
