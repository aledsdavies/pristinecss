package tokens

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	// Identifiers + literals

	COMMENT = "COMMENT"
	IDENT   = "IDENT"
	NUMBER  = "NUMBER"
	STRING  = "STRING"
	COLOR   = "COLOR"
	URI     = "URI"

	// Constructs

	STARTS_WITH = "^="
	DBLCOLON    = "::"

	// Symbols and operators
	AT          = "AT"
	COLON       = ":"
	SEMICOLON   = ";"
	COMMA       = ","
	DOT         = "."
	HASH        = "#"
	ASTERISK    = "*"
	PLUS        = "+"
	MINUS       = "-"
	DIVIDE      = "/"
	GREATER     = ">"
	TILDE       = "~"
	EQUALS      = "="
	PIPE        = "|"
	CARET       = "^"
	PERCENTAGE  = "%"
	DOLLAR      = "$"
	AMPERSAND   = "&"
	EXCLAMATION = "!"

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
	Literal []byte
	Line    int
	Column  int
}

// Token needs to implement the Erasable interface
func (t *Token) Erase() {
	t.Type = ILLEGAL
	t.Literal = t.Literal[:0]
	t.Line = 0
	t.Column = 0
}

func NewToken() *Token {
	return &Token{}
}

func (t *Token) Reset() {
	t.Type = ILLEGAL
	t.Literal = make([]byte, 0)
	t.Line = 0
	t.Column = 0
}

var keywords = map[string]TokenType{}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}

	return IDENT
}
