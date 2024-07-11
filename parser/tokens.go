package parser

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	// Identifiers + literals

	COMMENT = "COMMENT"
	IDENT   = "IDENT"
	NUMBER  = "NUMBER"
	STRING  = "STRING"
	AT_RULE = "AT_RULE"

	// CSS Specifics
	SELECTOR  = "SELECTOR"
	COLOR     = "COLOR"
	IMPORTANT = "IMPORTANT"
	UNIT      = "UNIT"

	// Symbols and operators
	COLON     = ":"
	SEMICOLON = ";"
	COMMA     = ","
	DOT       = "."
	HASH      = "#"
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

const initialBufferSize = 50

type Token struct {
	Type    TokenType
	Literal []rune
	buffer  []rune
	Line    int
	Column  int
}

func NewToken() *Token {
	return &Token{
		buffer: make([]rune, initialBufferSize),
	}
}

func (t *Token) AppendLiteral(b rune) {
	if len(t.Literal) == cap(t.buffer) {
		newBuffer := make([]rune, cap(t.buffer)*2)
		copy(newBuffer, t.buffer)
		t.buffer = newBuffer
	}
	t.Literal = append(t.Literal, b)
}

func (t *Token) Reset() {
	t.Type = ILLEGAL
	t.Literal = t.Literal[:0]
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

var units = [][]rune{
	// Absolute length units
	[]rune("cm"), []rune("mm"), []rune("in"), []rune("px"), []rune("pt"), []rune("pc"), []rune("Q"),
	// Relative length units
	[]rune("em"), []rune("ex"), []rune("ch"), []rune("rem"), []rune("lh"), []rune("rlh"), []rune("vb"), []rune("vi"),
	// Viewport-percentage lengths
	[]rune("vw"), []rune("vh"), []rune("vmin"), []rune("vmax"), []rune("svw"), []rune("svh"), []rune("lvw"), []rune("lvh"), []rune("dvw"), []rune("dvh"), []rune("vi"), []rune("vb"),
	// Container query length units
	[]rune("cqw"), []rune("cqh"), []rune("cqi"), []rune("cqb"), []rune("cqmin"), []rune("cqmax"),
	// Percentage
	[]rune("%"),
	// Angle units
	[]rune("deg"), []rune("grad"), []rune("rad"), []rune("turn"),
	// Time units
	[]rune("s"), []rune("ms"),
	// Frequency units
	[]rune("Hz"), []rune("kHz"),
	// Resolution units
	[]rune("dpi"), []rune("dpcm"), []rune("dppx"),
	// Flex units
	[]rune("fr"),
}

func isUnit(literal []rune) bool {
	for _, unit := range units {
		if runesEqual(literal, unit) {
			return true
		}
	}
	return false
}

// Helper function to compare two rune slices
func runesEqual(a, b []rune) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
