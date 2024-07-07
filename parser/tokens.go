package parser

import "fmt"

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	// Identifiers + literals

	COMMENT   = "COMMENT"
	IDENT     = "IDENT"
	NUMBER    = "NUMBER"
	UNIT      = "UNIT"
	STRING    = "STRING"
    COLOR     = "COLOR"
	IMPORTANT = "IMPORTANT"

	// Symbols and operators
	COLON     = ":"
	SEMICOLON = ";"
	COMMA     = ","
	DOT       = "."
	HASH      = "#"
	AT        = "@"
	ASTERISK  = "*"
	PLUS      = "+"
	MINUS     = "-"
	DIVIDE    = "/"
	GREATER   = ">"
	TILDE     = "~"
	EQUALS    = "="
	PIPE      = "|"
	CARET     = "^"
	DOLLAR    = "$"
	AMPERSAND = "&"

	// Brackets
	LPAREN   = "("
	RPAREN   = ")"
	LBRACKET = "["
	RBRACKET = "]"
	LBRACE   = "{"
	RBRACE   = "}"
)

type TokenType string

type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}

var keywords = map[string]TokenType{}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}

	return IDENT
}

var units = []string{
	// Absolute length units
	"cm", "mm", "in", "px", "pt", "pc", "Q",

	// Relative length units
	"em", "ex", "ch", "rem", "lh", "rlh", "vb", "vi",

	// Viewport-percentage lengths
	"vw", "vh", "vmin", "vmax", "svw", "svh", "lvw", "lvh", "dvw", "dvh", "vi", "vb",

	// Container query length units
	"cqw", "cqh", "cqi", "cqb", "cqmin", "cqmax",

	// Percentage
	"%",

	// Angle units
	"deg", "grad", "rad", "turn",

	// Time units
	"s", "ms",

	// Frequency units
	"Hz", "kHz",

	// Resolution units
	"dpi", "dpcm", "dppx",

	// Flex units
	"fr",
}

func isUnit(literal string) bool {
	if literal == "%" {
		fmt.Println("is percent")
	}
	for _, unit := range units {
		if literal == unit {
			return true
		}
	}
	return false
}
