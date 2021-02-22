package token

import "fmt"

type Token struct {
	Type Type
	Val  string
	Line int
}

func (i Token) String() string {
	switch i.Type {
	case EOF:
		return "EOF"
	case ERROR:
		return i.Val
	}
	return fmt.Sprintf("%s", i.Val)
}

type Type string

const (
	// special tokens
	ERROR = "ERROR"
	EOF   = "EOF"

	// literals
	IDENT  = "IDENTIFIER"
	NUMBER = "NUMBER"
	STRING = "STRING"
	JSON   = "JSON"

	// operators
	ARROW     = "<-"
	OR        = "||"
	AND       = "&&"
	COALESCE  = "??"
	EQUALS    = "=="
	GTE       = ">="
	GTR       = ">"
	LTE       = "<="
	LSS       = "<"
	NOT       = "!"
	NEQ       = "!="
	THEN      = "=>"
	SEMICOLON = ";"
	ASSIGN    = "="
	COLON     = "="
	LBRACKET  = "["
	RBRACKET  = "]"
	LPAREN    = "("
	RPAREN    = ")"
	COMMA     = ","
	LBRACE    = "{"
	RBRACE    = "}"
	ADD       = "+"
	SUB       = "-"
	REM       = "%"
	MUL       = "*"
	QUO       = "/"

	// keywords
	IF     = "if"
	ELSE   = "else"
	RETURN = "return"
	SWITCH = "switch"
	TRUE   = "true"
	FALSE  = "false"
	NULL   = "null"
)

var keywords = map[string]Type{
	"if":     IF,
	"else":   ELSE,
	"return": RETURN,
	"switch": SWITCH,
	"true":   TRUE,
	"false":  FALSE,
	"null":   NULL,
}

// Lookup checks if a candidate keyword token matches a keyword and returns the appropriate token.
// If the identifier is not a keyword, returns the IDENT token type instead.
func Lookup(ident string) Type {
	if tok, isKeyword := keywords[ident]; isKeyword {
		return tok
	}
	return IDENT
}
