package lexer_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aledsdavies/pristinecss/pkg/lexer"
	"github.com/aledsdavies/pristinecss/pkg/tokens"
)

func TestLexerPositiveCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []tokens.Token
		ignore   bool
	}{
		{
			name:  "Comments",
			input: "/* This is a comment */",
			expected: []tokens.Token{
				{Type: tokens.COMMENT, Literal: []byte("/* This is a comment */")},
			},
		},
		{
			name: "Universal Selector Cases",
			input: `
* { margin: 0; padding: 0; }
*, *::before, *::after { box-sizing: border-box; }
div * p { color: red; }
*[lang^=en] { color: green; }
*.warning { color: yellow; }
*#myid { font-weight: bold; }
    `,
			expected: []tokens.Token{
				{Type: tokens.ASTERISK, Literal: []byte("*")},
				{Type: tokens.LBRACE, Literal: []byte("{")},
				{Type: tokens.IDENT, Literal: []byte("margin")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.NUMBER, Literal: []byte("0")},
				{Type: tokens.SEMICOLON, Literal: []byte(";")},
				{Type: tokens.IDENT, Literal: []byte("padding")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.NUMBER, Literal: []byte("0")},
				{Type: tokens.SEMICOLON, Literal: []byte(";")},
				{Type: tokens.RBRACE, Literal: []byte("}")},

				{Type: tokens.ASTERISK, Literal: []byte("*")},
				{Type: tokens.COMMA, Literal: []byte(",")},
				{Type: tokens.ASTERISK, Literal: []byte("*")},
				{Type: tokens.DBLCOLON, Literal: []byte("::")},
				{Type: tokens.IDENT, Literal: []byte("before")},
				{Type: tokens.COMMA, Literal: []byte(",")},
				{Type: tokens.ASTERISK, Literal: []byte("*")},
				{Type: tokens.DBLCOLON, Literal: []byte("::")},
				{Type: tokens.IDENT, Literal: []byte("after")},
				{Type: tokens.LBRACE, Literal: []byte("{")},
				{Type: tokens.IDENT, Literal: []byte("box-sizing")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.IDENT, Literal: []byte("border-box")},
				{Type: tokens.SEMICOLON, Literal: []byte(";")},
				{Type: tokens.RBRACE, Literal: []byte("}")},

				{Type: tokens.IDENT, Literal: []byte("div")},
				{Type: tokens.ASTERISK, Literal: []byte("*")},
				{Type: tokens.IDENT, Literal: []byte("p")},
				{Type: tokens.LBRACE, Literal: []byte("{")},
				{Type: tokens.IDENT, Literal: []byte("color")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.IDENT, Literal: []byte("red")},
				{Type: tokens.SEMICOLON, Literal: []byte(";")},
				{Type: tokens.RBRACE, Literal: []byte("}")},

				{Type: tokens.ASTERISK, Literal: []byte("*")},
				{Type: tokens.LBRACKET, Literal: []byte("[")},
				{Type: tokens.IDENT, Literal: []byte("lang")},
				{Type: tokens.STARTS_WITH, Literal: []byte("^=")},
				{Type: tokens.IDENT, Literal: []byte("en")},
				{Type: tokens.RBRACKET, Literal: []byte("]")},

				{Type: tokens.LBRACE, Literal: []byte("{")},
				{Type: tokens.IDENT, Literal: []byte("color")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.IDENT, Literal: []byte("green")},
				{Type: tokens.SEMICOLON, Literal: []byte(";")},
				{Type: tokens.RBRACE, Literal: []byte("}")},

				{Type: tokens.ASTERISK, Literal: []byte("*")},
				{Type: tokens.DOT, Literal: []byte(".")},
				{Type: tokens.IDENT, Literal: []byte("warning")},
				{Type: tokens.LBRACE, Literal: []byte("{")},
				{Type: tokens.IDENT, Literal: []byte("color")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.IDENT, Literal: []byte("yellow")},
				{Type: tokens.SEMICOLON, Literal: []byte(";")},
				{Type: tokens.RBRACE, Literal: []byte("}")},

				{Type: tokens.ASTERISK, Literal: []byte("*")},
				{Type: tokens.HASH, Literal: []byte("#")},
				{Type: tokens.IDENT, Literal: []byte("myid")},
				{Type: tokens.LBRACE, Literal: []byte("{")},
				{Type: tokens.IDENT, Literal: []byte("font-weight")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.IDENT, Literal: []byte("bold")},
				{Type: tokens.SEMICOLON, Literal: []byte(";")},
				{Type: tokens.RBRACE, Literal: []byte("}")},
			},
		},
		{
			name:  "Combinators",
			input: "div > p + ul ~ span { color: red; }",
			expected: []tokens.Token{
				{Type: tokens.IDENT, Literal: []byte("div")},
				{Type: tokens.GREATER, Literal: []byte(">")},
				{Type: tokens.IDENT, Literal: []byte("p")},
				{Type: tokens.PLUS, Literal: []byte("+")},
				{Type: tokens.IDENT, Literal: []byte("ul")},
				{Type: tokens.TILDE, Literal: []byte("~")},
				{Type: tokens.IDENT, Literal: []byte("span")},
				{Type: tokens.LBRACE, Literal: []byte("{")},
				{Type: tokens.IDENT, Literal: []byte("color")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.IDENT, Literal: []byte("red")},
				{Type: tokens.SEMICOLON, Literal: []byte(";")},
				{Type: tokens.RBRACE, Literal: []byte("}")},
			},
		},
		{
			name:  "Pseudo-elements",
			input: "p::first-line { text-transform: uppercase; }",
			expected: []tokens.Token{
				{Type: tokens.IDENT, Literal: []byte("p")},
				{Type: tokens.DBLCOLON, Literal: []byte("::")},
				{Type: tokens.IDENT, Literal: []byte("first-line")},
				{Type: tokens.LBRACE, Literal: []byte("{")},
				{Type: tokens.IDENT, Literal: []byte("text-transform")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.IDENT, Literal: []byte("uppercase")},
				{Type: tokens.SEMICOLON, Literal: []byte(";")},
				{Type: tokens.RBRACE, Literal: []byte("}")},
			},
		},
		{
			name:  "Multiple Selectors",
			input: "h1, h2, h3 { font-family: sans-serif; }",
			expected: []tokens.Token{
				{Type: tokens.IDENT, Literal: []byte("h1")},
				{Type: tokens.COMMA, Literal: []byte(",")},
				{Type: tokens.IDENT, Literal: []byte("h2")},
				{Type: tokens.COMMA, Literal: []byte(",")},
				{Type: tokens.IDENT, Literal: []byte("h3")},
				{Type: tokens.LBRACE, Literal: []byte("{")},
				{Type: tokens.IDENT, Literal: []byte("font-family")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.IDENT, Literal: []byte("sans-serif")},
				{Type: tokens.SEMICOLON, Literal: []byte(";")},
				{Type: tokens.RBRACE, Literal: []byte("}")},
			},
		},
		{
			name: "Pseudo-classes",
			input: `
        :root { --main-color: blue; }
        li:nth-child(2n+1) { background: lightgray; }
        tr:nth-child(odd) { background-color: #f2f2f2; }
        div:nth-child(-n+3) { font-weight: bold; }
    `,
			expected: []tokens.Token{
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.IDENT, Literal: []byte("root")},
				{Type: tokens.LBRACE, Literal: []byte("{")},
				{Type: tokens.IDENT, Literal: []byte("--main-color")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.IDENT, Literal: []byte("blue")},
				{Type: tokens.SEMICOLON, Literal: []byte(";")},
				{Type: tokens.RBRACE, Literal: []byte("}")},

				{Type: tokens.IDENT, Literal: []byte("li")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.IDENT, Literal: []byte("nth-child")},
				{Type: tokens.LPAREN, Literal: []byte("(")},
				{Type: tokens.NUMBER, Literal: []byte("2")},
				{Type: tokens.IDENT, Literal: []byte("n")},
				{Type: tokens.PLUS, Literal: []byte("+")},
				{Type: tokens.NUMBER, Literal: []byte("1")},
				{Type: tokens.RPAREN, Literal: []byte(")")},
				{Type: tokens.LBRACE, Literal: []byte("{")},
				{Type: tokens.IDENT, Literal: []byte("background")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.IDENT, Literal: []byte("lightgray")},
				{Type: tokens.SEMICOLON, Literal: []byte(";")},
				{Type: tokens.RBRACE, Literal: []byte("}")},

				{Type: tokens.IDENT, Literal: []byte("tr")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.IDENT, Literal: []byte("nth-child")},
				{Type: tokens.LPAREN, Literal: []byte("(")},
				{Type: tokens.IDENT, Literal: []byte("odd")},
				{Type: tokens.RPAREN, Literal: []byte(")")},
				{Type: tokens.LBRACE, Literal: []byte("{")},
				{Type: tokens.IDENT, Literal: []byte("background-color")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.COLOR, Literal: []byte("#f2f2f2")},
				{Type: tokens.SEMICOLON, Literal: []byte(";")},
				{Type: tokens.RBRACE, Literal: []byte("}")},

				{Type: tokens.IDENT, Literal: []byte("div")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.IDENT, Literal: []byte("nth-child")},
				{Type: tokens.LPAREN, Literal: []byte("(")},
				{Type: tokens.IDENT, Literal: []byte("-n")},
				{Type: tokens.PLUS, Literal: []byte("+")},
				{Type: tokens.NUMBER, Literal: []byte("3")},
				{Type: tokens.RPAREN, Literal: []byte(")")},
				{Type: tokens.LBRACE, Literal: []byte("{")},
				{Type: tokens.IDENT, Literal: []byte("font-weight")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.IDENT, Literal: []byte("bold")},
				{Type: tokens.SEMICOLON, Literal: []byte(";")},
				{Type: tokens.RBRACE, Literal: []byte("}")},
			},
		},
		{
			name: "Pseudo-Elements",
			input: `
        p::first-line { font-weight: bold; }
        div::before { content: ""; display: block; }
    `,
			expected: []tokens.Token{
				{Type: tokens.IDENT, Literal: []byte("p")},
				{Type: tokens.DBLCOLON, Literal: []byte("::")},
				{Type: tokens.IDENT, Literal: []byte("first-line")},
				{Type: tokens.LBRACE, Literal: []byte("{")},
				{Type: tokens.IDENT, Literal: []byte("font-weight")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.IDENT, Literal: []byte("bold")},
				{Type: tokens.SEMICOLON, Literal: []byte(";")},
				{Type: tokens.RBRACE, Literal: []byte("}")},

				{Type: tokens.IDENT, Literal: []byte("div")},
				{Type: tokens.DBLCOLON, Literal: []byte("::")},
				{Type: tokens.IDENT, Literal: []byte("before")},
				{Type: tokens.LBRACE, Literal: []byte("{")},
				{Type: tokens.IDENT, Literal: []byte("content")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.STRING, Literal: []byte("\"\"")},
				{Type: tokens.SEMICOLON, Literal: []byte(";")},
				{Type: tokens.IDENT, Literal: []byte("display")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.IDENT, Literal: []byte("block")},
				{Type: tokens.SEMICOLON, Literal: []byte(";")},
				{Type: tokens.RBRACE, Literal: []byte("}")},
			},
		},
		{
			name:  "Media Query",
			input: "@media screen and (max-width: 600px) { body { font-size: 14px; } }",
			expected: []tokens.Token{
				{Type: tokens.AT, Literal: []byte("@")},
				{Type: tokens.IDENT, Literal: []byte("media")},
				{Type: tokens.IDENT, Literal: []byte("screen")},
				{Type: tokens.IDENT, Literal: []byte("and")},
				{Type: tokens.LPAREN, Literal: []byte("(")},
				{Type: tokens.IDENT, Literal: []byte("max-width")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.NUMBER, Literal: []byte("600")},
				{Type: tokens.IDENT, Literal: []byte("px")},
				{Type: tokens.RPAREN, Literal: []byte(")")},
				{Type: tokens.LBRACE, Literal: []byte("{")},
				{Type: tokens.IDENT, Literal: []byte("body")},
				{Type: tokens.LBRACE, Literal: []byte("{")},
				{Type: tokens.IDENT, Literal: []byte("font-size")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.NUMBER, Literal: []byte("14")},
				{Type: tokens.IDENT, Literal: []byte("px")},
				{Type: tokens.SEMICOLON, Literal: []byte(";")},
				{Type: tokens.RBRACE, Literal: []byte("}")},
				{Type: tokens.RBRACE, Literal: []byte("}")},
			},
		},
		{
			name:  "Complex Selectors",
			input: "#main > .article p:first-child { color: #ff0000; }",
			expected: []tokens.Token{
				{Type: tokens.HASH, Literal: []byte("#")},
				{Type: tokens.IDENT, Literal: []byte("main")},
				{Type: tokens.GREATER, Literal: []byte(">")},
				{Type: tokens.DOT, Literal: []byte(".")},
				{Type: tokens.IDENT, Literal: []byte("article")},
				{Type: tokens.IDENT, Literal: []byte("p")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.IDENT, Literal: []byte("first-child")},
				{Type: tokens.LBRACE, Literal: []byte("{")},
				{Type: tokens.IDENT, Literal: []byte("color")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.COLOR, Literal: []byte("#ff0000")},
				{Type: tokens.SEMICOLON, Literal: []byte(";")},
				{Type: tokens.RBRACE, Literal: []byte("}")},
			},
		},
		{
			name: "Handle Important",
			input: `main { color: #ff0000 !important; }
p { margin: 10px ! important; }
div { padding: 5px !   important; }
span { font-weight: bold !
important; }
.custom { border: 1px solid black!important }`,
			expected: []tokens.Token{
				{Type: tokens.IDENT, Literal: []byte("main")},
				{Type: tokens.LBRACE, Literal: []byte("{")},
				{Type: tokens.IDENT, Literal: []byte("color")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.COLOR, Literal: []byte("#ff0000")},
				{Type: tokens.EXCLAMATION, Literal: []byte("!")},
				{Type: tokens.IDENT, Literal: []byte("important")},
				{Type: tokens.SEMICOLON, Literal: []byte(";")},
				{Type: tokens.RBRACE, Literal: []byte("}")},

				{Type: tokens.IDENT, Literal: []byte("p")},
				{Type: tokens.LBRACE, Literal: []byte("{")},
				{Type: tokens.IDENT, Literal: []byte("margin")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.NUMBER, Literal: []byte("10")},
				{Type: tokens.IDENT, Literal: []byte("px")},
				{Type: tokens.EXCLAMATION, Literal: []byte("!")},
				{Type: tokens.IDENT, Literal: []byte("important")},
				{Type: tokens.SEMICOLON, Literal: []byte(";")},
				{Type: tokens.RBRACE, Literal: []byte("}")},

				{Type: tokens.IDENT, Literal: []byte("div")},
				{Type: tokens.LBRACE, Literal: []byte("{")},
				{Type: tokens.IDENT, Literal: []byte("padding")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.NUMBER, Literal: []byte("5")},
				{Type: tokens.IDENT, Literal: []byte("px")},
				{Type: tokens.EXCLAMATION, Literal: []byte("!")},
				{Type: tokens.IDENT, Literal: []byte("important")},
				{Type: tokens.SEMICOLON, Literal: []byte(";")},
				{Type: tokens.RBRACE, Literal: []byte("}")},

				{Type: tokens.IDENT, Literal: []byte("span")},
				{Type: tokens.LBRACE, Literal: []byte("{")},
				{Type: tokens.IDENT, Literal: []byte("font-weight")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.IDENT, Literal: []byte("bold")},
				{Type: tokens.EXCLAMATION, Literal: []byte("!")},
				{Type: tokens.IDENT, Literal: []byte("important")},
				{Type: tokens.SEMICOLON, Literal: []byte(";")},
				{Type: tokens.RBRACE, Literal: []byte("}")},

				{Type: tokens.DOT, Literal: []byte(".")},
				{Type: tokens.IDENT, Literal: []byte("custom")},
				{Type: tokens.LBRACE, Literal: []byte("{")},
				{Type: tokens.IDENT, Literal: []byte("border")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.NUMBER, Literal: []byte("1")},
				{Type: tokens.IDENT, Literal: []byte("px")},
				{Type: tokens.IDENT, Literal: []byte("solid")},
				{Type: tokens.IDENT, Literal: []byte("black")},
				{Type: tokens.EXCLAMATION, Literal: []byte("!")},
				{Type: tokens.IDENT, Literal: []byte("important")},
				{Type: tokens.RBRACE, Literal: []byte("}")},
			},
		},
		{
			name:  "Attribute Selector",
			input: "a[href^=\"https://\"] { color: green; }",
			expected: []tokens.Token{
				{Type: tokens.IDENT, Literal: []byte("a")},
				{Type: tokens.LBRACKET, Literal: []byte("[")},
				{Type: tokens.IDENT, Literal: []byte("href")},
				{Type: tokens.STARTS_WITH, Literal: []byte("^=")},
				{Type: tokens.STRING, Literal: []byte("\"https://\"")},
				{Type: tokens.RBRACKET, Literal: []byte("]")},

				{Type: tokens.LBRACE, Literal: []byte("{")},
				{Type: tokens.IDENT, Literal: []byte("color")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.IDENT, Literal: []byte("green")},
				{Type: tokens.SEMICOLON, Literal: []byte(";")},
				{Type: tokens.RBRACE, Literal: []byte("}")},
			},
		},
		{
			name:  "Keyframes",
			input: "@keyframes fadeIn { 0% { opacity: 0; } 100% { opacity: 1; } }",
			expected: []tokens.Token{
				{Type: tokens.AT, Literal: []byte("@")},
				{Type: tokens.IDENT, Literal: []byte("keyframes")},
				{Type: tokens.IDENT, Literal: []byte("fadeIn")},
				{Type: tokens.LBRACE, Literal: []byte("{")},
				{Type: tokens.NUMBER, Literal: []byte("0")},
				{Type: tokens.PERCENTAGE, Literal: []byte("%")},
				{Type: tokens.LBRACE, Literal: []byte("{")},
				{Type: tokens.IDENT, Literal: []byte("opacity")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.NUMBER, Literal: []byte("0")},
				{Type: tokens.SEMICOLON, Literal: []byte(";")},
				{Type: tokens.RBRACE, Literal: []byte("}")},
				{Type: tokens.NUMBER, Literal: []byte("100")},
				{Type: tokens.PERCENTAGE, Literal: []byte("%")},
				{Type: tokens.LBRACE, Literal: []byte("{")},
				{Type: tokens.IDENT, Literal: []byte("opacity")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.NUMBER, Literal: []byte("1")},
				{Type: tokens.SEMICOLON, Literal: []byte(";")},
				{Type: tokens.RBRACE, Literal: []byte("}")},
				{Type: tokens.RBRACE, Literal: []byte("}")},
			},
		},
		{
			name:  "CSS Variables",
			input: ":root { --main-color: blue; } body { color: var(--main-color); }",
			expected: []tokens.Token{
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.IDENT, Literal: []byte("root")},
				{Type: tokens.LBRACE, Literal: []byte("{")},
				{Type: tokens.IDENT, Literal: []byte("--main-color")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.IDENT, Literal: []byte("blue")},
				{Type: tokens.SEMICOLON, Literal: []byte(";")},
				{Type: tokens.RBRACE, Literal: []byte("}")},

				{Type: tokens.IDENT, Literal: []byte("body")},
				{Type: tokens.LBRACE, Literal: []byte("{")},
				{Type: tokens.IDENT, Literal: []byte("color")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.IDENT, Literal: []byte("var")},
				{Type: tokens.LPAREN, Literal: []byte("(")},
				{Type: tokens.IDENT, Literal: []byte("--main-color")},
				{Type: tokens.RPAREN, Literal: []byte(")")},
				{Type: tokens.SEMICOLON, Literal: []byte(";")},
				{Type: tokens.RBRACE, Literal: []byte("}")},
			},
		},
		{
			name:  "Calc Function",
			input: "div { width: calc(100% - 20px); height: 100vh; }",
			expected: []tokens.Token{
				{Type: tokens.IDENT, Literal: []byte("div")},
				{Type: tokens.LBRACE, Literal: []byte("{")},
				{Type: tokens.IDENT, Literal: []byte("width")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.IDENT, Literal: []byte("calc")},
				{Type: tokens.LPAREN, Literal: []byte("(")},
				{Type: tokens.NUMBER, Literal: []byte("100")},
				{Type: tokens.PERCENTAGE, Literal: []byte("%")},
				{Type: tokens.MINUS, Literal: []byte("-")},
				{Type: tokens.NUMBER, Literal: []byte("20")},
				{Type: tokens.IDENT, Literal: []byte("px")},
				{Type: tokens.RPAREN, Literal: []byte(")")},
				{Type: tokens.SEMICOLON, Literal: []byte(";")},
				{Type: tokens.IDENT, Literal: []byte("height")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.NUMBER, Literal: []byte("100")},
				{Type: tokens.IDENT, Literal: []byte("vh")},
				{Type: tokens.SEMICOLON, Literal: []byte(";")},
				{Type: tokens.RBRACE, Literal: []byte("}")},
			},
		},
		{
			name:  "Escaped characters in identifiers",
			input: ".foo\\.bar { color: red; }",
			expected: []tokens.Token{
				{Type: tokens.DOT, Literal: []byte(".")},
				{Type: tokens.IDENT, Literal: []byte("foo\\.bar")},
				{Type: tokens.LBRACE, Literal: []byte("{")},
				{Type: tokens.IDENT, Literal: []byte("color")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.IDENT, Literal: []byte("red")},
				{Type: tokens.SEMICOLON, Literal: []byte(";")},
				{Type: tokens.RBRACE, Literal: []byte("}")},
			},
		},
		{
			name:  "Escaped characters in attribute selectors",
			input: "a[href=\"foo\\\"bar\"] { color: blue; }",
			expected: []tokens.Token{
				{Type: tokens.IDENT, Literal: []byte("a")},
				{Type: tokens.LBRACKET, Literal: []byte("[")},
				{Type: tokens.IDENT, Literal: []byte("href")},
				{Type: tokens.EQUALS, Literal: []byte("=")},
				{Type: tokens.STRING, Literal: []byte("\"foo\\\"bar\"")},
				{Type: tokens.RBRACKET, Literal: []byte("]")},
				{Type: tokens.LBRACE, Literal: []byte("{")},
				{Type: tokens.IDENT, Literal: []byte("color")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.IDENT, Literal: []byte("blue")},
				{Type: tokens.SEMICOLON, Literal: []byte(";")},
				{Type: tokens.RBRACE, Literal: []byte("}")},
			},
		},
		{
			name: "Unicode and special character identifiers",
			input: `
#☃ { color: skyblue; }
.günther, #π_value, .こんにちは {
    color: blue;
}
.--custom-prop, .-moz-custom {
    value: 123;
}
#\26 ABC { /* Escaped ASCII */
    color: red;
}
`,
			expected: []tokens.Token{
				{Type: tokens.HASH, Literal: []byte("#")},
				{Type: tokens.IDENT, Literal: []byte("☃")},
				{Type: tokens.LBRACE, Literal: []byte("{")},
				{Type: tokens.IDENT, Literal: []byte("color")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.IDENT, Literal: []byte("skyblue")},
				{Type: tokens.SEMICOLON, Literal: []byte(";")},
				{Type: tokens.RBRACE, Literal: []byte("}")},
				{Type: tokens.DOT, Literal: []byte(".")},
				{Type: tokens.IDENT, Literal: []byte("günther")},
				{Type: tokens.COMMA, Literal: []byte(",")},
				{Type: tokens.HASH, Literal: []byte("#")},
				{Type: tokens.IDENT, Literal: []byte("π_value")},
				{Type: tokens.COMMA, Literal: []byte(",")},
				{Type: tokens.DOT, Literal: []byte(".")},
				{Type: tokens.IDENT, Literal: []byte("こんにちは")},
				{Type: tokens.LBRACE, Literal: []byte("{")},
				{Type: tokens.IDENT, Literal: []byte("color")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.IDENT, Literal: []byte("blue")},
				{Type: tokens.SEMICOLON, Literal: []byte(";")},
				{Type: tokens.RBRACE, Literal: []byte("}")},

				{Type: tokens.DOT, Literal: []byte(".")},
				{Type: tokens.IDENT, Literal: []byte("--custom-prop")},
				{Type: tokens.COMMA, Literal: []byte(",")},
				{Type: tokens.DOT, Literal: []byte(".")},
				{Type: tokens.IDENT, Literal: []byte("-moz-custom")},
				{Type: tokens.LBRACE, Literal: []byte("{")},
				{Type: tokens.IDENT, Literal: []byte("value")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.NUMBER, Literal: []byte("123")},
				{Type: tokens.SEMICOLON, Literal: []byte(";")},
				{Type: tokens.RBRACE, Literal: []byte("}")},

				{Type: tokens.HASH, Literal: []byte("#")},
				{Type: tokens.IDENT, Literal: []byte("\\26 ABC")},
				{Type: tokens.LBRACE, Literal: []byte("{")},
				{Type: tokens.COMMENT, Literal: []byte("/* Escaped ASCII */")},
				{Type: tokens.IDENT, Literal: []byte("color")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.IDENT, Literal: []byte("red")},
				{Type: tokens.SEMICOLON, Literal: []byte(";")},
				{Type: tokens.RBRACE, Literal: []byte("}")},
			},
		},
		{
			name:  "CSS custom properties (variables)",
			input: ":root { --custom-color: #ff00ff; } .foo { color: var(--custom-color); }",
			expected: []tokens.Token{
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.IDENT, Literal: []byte("root")},
				{Type: tokens.LBRACE, Literal: []byte("{")},
				{Type: tokens.IDENT, Literal: []byte("--custom-color")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.COLOR, Literal: []byte("#ff00ff")},
				{Type: tokens.SEMICOLON, Literal: []byte(";")},
				{Type: tokens.RBRACE, Literal: []byte("}")},

				{Type: tokens.DOT, Literal: []byte(".")},
				{Type: tokens.IDENT, Literal: []byte("foo")},
				{Type: tokens.LBRACE, Literal: []byte("{")},
				{Type: tokens.IDENT, Literal: []byte("color")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.IDENT, Literal: []byte("var")},
				{Type: tokens.LPAREN, Literal: []byte("(")},
				{Type: tokens.IDENT, Literal: []byte("--custom-color")},
				{Type: tokens.RPAREN, Literal: []byte(")")},
				{Type: tokens.SEMICOLON, Literal: []byte(";")},
				{Type: tokens.RBRACE, Literal: []byte("}")},
			},
		},
		{
			name: "Complex property values",
			input: `
            .gradient {
                background-image: linear-gradient(45deg, #ff0000, #00ff00),
                                  radial-gradient(circle, #0000ff, #ffff00);
                font: bold 12px/14px "Helvetica", sans-serif;
            }`,
			expected: []tokens.Token{
				{Type: tokens.DOT, Literal: []byte(".")},
				{Type: tokens.IDENT, Literal: []byte("gradient")},
				{Type: tokens.LBRACE, Literal: []byte("{")},
				{Type: tokens.IDENT, Literal: []byte("background-image")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.IDENT, Literal: []byte("linear-gradient")},
				{Type: tokens.LPAREN, Literal: []byte("(")},
				{Type: tokens.NUMBER, Literal: []byte("45")},
				{Type: tokens.IDENT, Literal: []byte("deg")},
				{Type: tokens.COMMA, Literal: []byte(",")},
				{Type: tokens.COLOR, Literal: []byte("#ff0000")},
				{Type: tokens.COMMA, Literal: []byte(",")},
				{Type: tokens.COLOR, Literal: []byte("#00ff00")},
				{Type: tokens.RPAREN, Literal: []byte(")")},
				{Type: tokens.COMMA, Literal: []byte(",")},
				{Type: tokens.IDENT, Literal: []byte("radial-gradient")},
				{Type: tokens.LPAREN, Literal: []byte("(")},
				{Type: tokens.IDENT, Literal: []byte("circle")},
				{Type: tokens.COMMA, Literal: []byte(",")},
				{Type: tokens.COLOR, Literal: []byte("#0000ff")},
				{Type: tokens.COMMA, Literal: []byte(",")},
				{Type: tokens.COLOR, Literal: []byte("#ffff00")},
				{Type: tokens.RPAREN, Literal: []byte(")")},
				{Type: tokens.SEMICOLON, Literal: []byte(";")},
				{Type: tokens.IDENT, Literal: []byte("font")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.IDENT, Literal: []byte("bold")},
				{Type: tokens.NUMBER, Literal: []byte("12")},
				{Type: tokens.IDENT, Literal: []byte("px")},
				{Type: tokens.DIVIDE, Literal: []byte("/")},
				{Type: tokens.NUMBER, Literal: []byte("14")},
				{Type: tokens.IDENT, Literal: []byte("px")},
				{Type: tokens.STRING, Literal: []byte("\"Helvetica\"")},
				{Type: tokens.COMMA, Literal: []byte(",")},
				{Type: tokens.IDENT, Literal: []byte("sans-serif")},
				{Type: tokens.SEMICOLON, Literal: []byte(";")},
				{Type: tokens.RBRACE, Literal: []byte("}")},
			},
		},
		{
			name:   "CSS Hacks and Legacy Browser Support",
			ignore: true,
			input: `
    /* Pseudo-element hack */
    _::-webkit-full-page-media, _:future, :root .foo {
        color: red;
    }

    /* Pseudo-class hack */
    _:-webkit-full-screen, :root .bar {
        display: block;
    }

    /* Media query hack */
    @media screen and (min-width:0\0) {
        .baz { zoom: 1; }
    }

    /* Property value hacks */
    .hack1 { property: value\9; }
    .hack2 { property: value \9; }
    .hack3 { property: value\0; }
    .hack4 { property: value \0; }

    /* Combination of hacks */
    _:-ms-lang(x), _:-webkit-full-screen, .multi-hack {
        *display: inline;
        _height: 1%;
    }
    `,
			expected: []tokens.Token{
				{Type: tokens.COMMENT, Literal: []byte("/* Pseudo-element hack */")},
				{Type: tokens.IDENT, Literal: []byte("_")},
				{Type: tokens.DBLCOLON, Literal: []byte("::")},
				{Type: tokens.IDENT, Literal: []byte("-webkit-full-page-media")},
				{Type: tokens.COMMA, Literal: []byte(",")},
				{Type: tokens.IDENT, Literal: []byte("_")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.IDENT, Literal: []byte("future")},
				{Type: tokens.COMMA, Literal: []byte(",")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.IDENT, Literal: []byte("root")},
				{Type: tokens.DOT, Literal: []byte(".")},
				{Type: tokens.IDENT, Literal: []byte("foo")},
				{Type: tokens.LBRACE, Literal: []byte("{")},
				{Type: tokens.IDENT, Literal: []byte("color")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.IDENT, Literal: []byte("red")},
				{Type: tokens.SEMICOLON, Literal: []byte(";")},
				{Type: tokens.RBRACE, Literal: []byte("}")},

				{Type: tokens.COMMENT, Literal: []byte("/* Pseudo-class hack */")},
				{Type: tokens.IDENT, Literal: []byte("_")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.IDENT, Literal: []byte("-webkit-full-screen")},
				{Type: tokens.COMMA, Literal: []byte(",")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.IDENT, Literal: []byte("root")},
				{Type: tokens.DOT, Literal: []byte(".")},
				{Type: tokens.IDENT, Literal: []byte("bar")},
				{Type: tokens.LBRACE, Literal: []byte("{")},
				{Type: tokens.IDENT, Literal: []byte("display")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.IDENT, Literal: []byte("block")},
				{Type: tokens.SEMICOLON, Literal: []byte(";")},
				{Type: tokens.RBRACE, Literal: []byte("}")},

				{Type: tokens.COMMENT, Literal: []byte("/* Media query hack */")},
				{Type: tokens.AT, Literal: []byte("@")},
				{Type: tokens.IDENT, Literal: []byte("media")},
				{Type: tokens.IDENT, Literal: []byte("screen")},
				{Type: tokens.IDENT, Literal: []byte("and")},
				{Type: tokens.LPAREN, Literal: []byte("(")},
				{Type: tokens.IDENT, Literal: []byte("min-width")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.NUMBER, Literal: []byte("0")},
				{Type: tokens.IDENT, Literal: []byte("\\0")},
				{Type: tokens.RPAREN, Literal: []byte(")")},
				{Type: tokens.LBRACE, Literal: []byte("{")},
				{Type: tokens.DOT, Literal: []byte(".")},
				{Type: tokens.IDENT, Literal: []byte("baz")},
				{Type: tokens.LBRACE, Literal: []byte("{")},
				{Type: tokens.IDENT, Literal: []byte("zoom")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.NUMBER, Literal: []byte("1")},
				{Type: tokens.SEMICOLON, Literal: []byte(";")},
				{Type: tokens.RBRACE, Literal: []byte("}")},
				{Type: tokens.RBRACE, Literal: []byte("}")},

				{Type: tokens.COMMENT, Literal: []byte("/* Property value hacks */")},
				{Type: tokens.DOT, Literal: []byte(".")},
				{Type: tokens.IDENT, Literal: []byte("hack1")},
				{Type: tokens.LBRACE, Literal: []byte("{")},
				{Type: tokens.IDENT, Literal: []byte("property")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.IDENT, Literal: []byte("value\\9")},
				{Type: tokens.SEMICOLON, Literal: []byte(";")},
				{Type: tokens.RBRACE, Literal: []byte("}")},

				{Type: tokens.DOT, Literal: []byte(".")},
				{Type: tokens.IDENT, Literal: []byte("hack2")},
				{Type: tokens.LBRACE, Literal: []byte("{")},
				{Type: tokens.IDENT, Literal: []byte("property")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.IDENT, Literal: []byte("value\\9")},
				{Type: tokens.SEMICOLON, Literal: []byte(";")},
				{Type: tokens.RBRACE, Literal: []byte("}")},

				{Type: tokens.DOT, Literal: []byte(".")},
				{Type: tokens.IDENT, Literal: []byte("hack3")},
				{Type: tokens.LBRACE, Literal: []byte("{")},
				{Type: tokens.IDENT, Literal: []byte("property")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.IDENT, Literal: []byte("value\\0")},
				{Type: tokens.SEMICOLON, Literal: []byte(";")},
				{Type: tokens.RBRACE, Literal: []byte("}")},

				{Type: tokens.DOT, Literal: []byte(".")},
				{Type: tokens.IDENT, Literal: []byte("hack4")},
				{Type: tokens.LBRACE, Literal: []byte("{")},
				{Type: tokens.IDENT, Literal: []byte("property")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.IDENT, Literal: []byte("value\\0")},
				{Type: tokens.SEMICOLON, Literal: []byte(";")},
				{Type: tokens.RBRACE, Literal: []byte("}")},

				{Type: tokens.COMMENT, Literal: []byte("/* Combination of hacks */")},
				{Type: tokens.IDENT, Literal: []byte("_")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.IDENT, Literal: []byte("-ms-lang")},
				{Type: tokens.LPAREN, Literal: []byte("(")},
				{Type: tokens.IDENT, Literal: []byte("x")},
				{Type: tokens.RPAREN, Literal: []byte(")")},
				{Type: tokens.COMMA, Literal: []byte(",")},
				{Type: tokens.IDENT, Literal: []byte("_")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.IDENT, Literal: []byte("-webkit-full-screen")},
				{Type: tokens.COMMA, Literal: []byte(",")},
				{Type: tokens.DOT, Literal: []byte(".")},
				{Type: tokens.IDENT, Literal: []byte("multi-hack")},
				{Type: tokens.LBRACE, Literal: []byte("{")},
				{Type: tokens.ASTERISK, Literal: []byte("*")},
				{Type: tokens.IDENT, Literal: []byte("display")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.IDENT, Literal: []byte("inline")},
				{Type: tokens.SEMICOLON, Literal: []byte(";")},
				{Type: tokens.IDENT, Literal: []byte("_height")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.NUMBER, Literal: []byte("1")},
				{Type: tokens.PERCENTAGE, Literal: []byte("%")},
				{Type: tokens.SEMICOLON, Literal: []byte(";")},
				{Type: tokens.RBRACE, Literal: []byte("}")},
			},
		},
	}

	for _, tt := range tests {
		if tt.ignore {
			continue
		}
		t.Run(tt.name, func(t *testing.T) {
			tokens := lexer.Lex(strings.NewReader(tt.input))

			if len(tokens)-1 != len(tt.expected) {
				t.Fatalf("Token count mismatch. Expected %d tokens, got %d", len(tt.expected), len(tokens))
			}

			for i, expected := range tt.expected {
				tok := tokens[i]
				if tok.Type != expected.Type {
					t.Errorf("tests[%d] - tokentype wrong. expected=%q, got=%q %s",
						i, expected.Type, tok.Type, string(tok.Literal))
				}
				if !bytesEqual(tok.Literal, expected.Literal) {
					t.Errorf("tests[%d] - literal wrong. expected=%q, got=%q",
						i, string(expected.Literal), string(tok.Literal))
				}
				if tok.Line == 0 || tok.Column == 0 {
					t.Errorf("tests[%d] - line or column not set. got line=%d, column=%d", i, tok.Line, tok.Column)
				}
			}
		})
	}
}

func TestIllegalCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []tokens.Token
	}{
		{
			name:  "Invalid hex color",
			input: ".invalid { color: #1234ZZ; }",
			expected: []tokens.Token{
				{Type: tokens.DOT, Literal: []byte(".")},
				{Type: tokens.IDENT, Literal: []byte("invalid")},
				{Type: tokens.LBRACE, Literal: []byte("{")},
				{Type: tokens.IDENT, Literal: []byte("color")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.HASH, Literal: []byte("#")},
				{Type: tokens.NUMBER, Literal: []byte("1234")},
				{Type: tokens.IDENT, Literal: []byte("ZZ")},
				{Type: tokens.SEMICOLON, Literal: []byte(";")},
				{Type: tokens.RBRACE, Literal: []byte("}")},
			},
		},
		{
			name:  "Invalid unit combination",
			input: ".invalid-unit { width: 50+px; }",
			expected: []tokens.Token{
				{Type: tokens.DOT, Literal: []byte(".")},
				{Type: tokens.IDENT, Literal: []byte("invalid-unit")},
				{Type: tokens.LBRACE, Literal: []byte("{")},
				{Type: tokens.IDENT, Literal: []byte("width")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.NUMBER, Literal: []byte("50")},
				{Type: tokens.PLUS, Literal: []byte("+")},
				{Type: tokens.IDENT, Literal: []byte("px")},
				{Type: tokens.SEMICOLON, Literal: []byte(";")},
				{Type: tokens.RBRACE, Literal: []byte("}")},
			},
		},
		{
			name:  "Invalid percentage",
			input: ".invalid-percentage { height: 100vh%; }",
			expected: []tokens.Token{
				{Type: tokens.DOT, Literal: []byte(".")},
				{Type: tokens.IDENT, Literal: []byte("invalid-percentage")},
				{Type: tokens.LBRACE, Literal: []byte("{")},
				{Type: tokens.IDENT, Literal: []byte("height")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.NUMBER, Literal: []byte("100")},
				{Type: tokens.IDENT, Literal: []byte("vh")},
				{Type: tokens.PERCENTAGE, Literal: []byte("%")},
				{Type: tokens.SEMICOLON, Literal: []byte(";")},
				{Type: tokens.RBRACE, Literal: []byte("}")},
			},
		},
		{
			name:  "Invalid property",
			input: ".invalid-property { colo r: red; }",
			expected: []tokens.Token{
				{Type: tokens.DOT, Literal: []byte(".")},
				{Type: tokens.IDENT, Literal: []byte("invalid-property")},
				{Type: tokens.LBRACE, Literal: []byte("{")},
				{Type: tokens.IDENT, Literal: []byte("colo")},
				{Type: tokens.IDENT, Literal: []byte("r")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.IDENT, Literal: []byte("red")},
				{Type: tokens.SEMICOLON, Literal: []byte(";")},
				{Type: tokens.RBRACE, Literal: []byte("}")},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := lexer.Lex(strings.NewReader(tt.input))

			if len(tokens)-1 != len(tt.expected) {
				t.Fatalf("Token count mismatch. Expected %d tokens, got %d", len(tt.expected), len(tokens))
			}

			for i, expected := range tt.expected {
				got := tokens[i]
				if got.Type != expected.Type {
					t.Errorf("Token %d: expected type %v, got %v", i, expected.Type, got.Type)
				}
				if !bytesEqual(got.Literal, expected.Literal) {
					t.Errorf("Token %d: expected literal %q, got %q", i, string(expected.Literal), string(got.Literal))
				}
				// Optionally, check line and column numbers if your parallel lexer preserves this information
				if got.Line != expected.Line || got.Column != expected.Column {
					t.Errorf("Token %d: expected position line %d, column %d; got line %d, column %d",
						i, expected.Line, expected.Column, got.Line, got.Column)
				}
			}
		})
	}
}

func TestFrameworks(t *testing.T) {
	tests := []struct {
		name     string
		filepath string
		expected int
	}{
		{
			name:     "Can Lex Bootstrap css without ILLEGALS",
			filepath: filepath.Join("..", "..", "test-data", "frameworks", "bootstrap.css"),
			expected: 0,
		},
		{
			name:     "Can Lex Bulma css without ILLEGALS",
			filepath: filepath.Join("..", "..", "test-data", "frameworks", "bulma.css"),
			expected: 0,
		},
		{
			name:     "Can Lex Foundation css without ILLEGALS",
			filepath: filepath.Join("..", "..", "test-data", "frameworks", "foundation.css"),
			expected: 0,
		},
		{
			name:     "Can Lex Materialize css without ILLEGALS",
			filepath: filepath.Join("..", "..", "test-data", "frameworks", "materialize.css"),
			expected: 0,
		},
		{
			name:     "Can Lex Spectre css without ILLEGALS",
			filepath: filepath.Join("..", "..", "test-data", "frameworks", "spectre.css"),
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input, err := os.Open(tt.filepath)
			if err != nil {
				t.Fatalf("Could not open the file %s: %v", tt.filepath, err)
			}
			defer input.Close()

			toks := lexer.Lex(input)
			illegalCount := 0
			for _, tok := range toks {
				if tok.Type == tokens.ILLEGAL {
					illegalCount++
					t.Errorf("Found an ILLEGAL token: %v %s  %d:%d", tok.Type, string(tok.Literal), tok.Line, tok.Column)
				}
			}

			if illegalCount != tt.expected {
				t.Errorf("Expected %d ILLEGAL tokens, but found %d", tt.expected, illegalCount)
			}
		})
	}
}

func BenchmarkFrameworks(b *testing.B) {
	frameworks := []struct {
		name string
		path string
	}{
		{"Bootstrap", filepath.Join("..", "..", "test-data", "frameworks", "bootstrap.css")},
		{"Bulma", filepath.Join("..", "..", "test-data", "frameworks", "bulma.css")},
		{"Foundation", filepath.Join("..", "..", "test-data", "frameworks", "foundation.css")},
		{"Materialize", filepath.Join("..", "..", "test-data", "frameworks", "materialize.css")},
		{"Spectre", filepath.Join("..", "..", "test-data", "frameworks", "spectre.css")},
	}

	for _, fw := range frameworks {
		content, err := os.ReadFile(fw.path)
		if err != nil {
			b.Fatalf("Could not read the file %s: %v", fw.path, err)
		}

		b.Run(fw.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				reader := bytes.NewReader(content)
				tokens := lexer.Lex(reader)
				_ = tokens
			}
		})
	}
}

func bytesEqual(a, b []byte) bool {
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
