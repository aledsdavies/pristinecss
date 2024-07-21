package parser_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aledsdavies/pristinecss/parser"
)

func TestLexerPositiveCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []parser.Token
		ignore   bool
	}{
		{
			name:  "Comments",
			input: "/* This is a comment */",
			expected: []parser.Token{
				{Type: parser.COMMENT, Literal: []byte("/* This is a comment */")},
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
			expected: []parser.Token{
				{Type: parser.ASTERISK, Literal: []byte("*")},
				{Type: parser.LBRACE, Literal: []byte("{")},
				{Type: parser.IDENT, Literal: []byte("margin")},
				{Type: parser.COLON, Literal: []byte(":")},
				{Type: parser.NUMBER, Literal: []byte("0")},
				{Type: parser.SEMICOLON, Literal: []byte(";")},
				{Type: parser.IDENT, Literal: []byte("padding")},
				{Type: parser.COLON, Literal: []byte(":")},
				{Type: parser.NUMBER, Literal: []byte("0")},
				{Type: parser.SEMICOLON, Literal: []byte(";")},
				{Type: parser.RBRACE, Literal: []byte("}")},

				{Type: parser.ASTERISK, Literal: []byte("*")},
				{Type: parser.COMMA, Literal: []byte(",")},
				{Type: parser.ASTERISK, Literal: []byte("*")},
				{Type: parser.DBLCOLON, Literal: []byte("::")},
				{Type: parser.IDENT, Literal: []byte("before")},
				{Type: parser.COMMA, Literal: []byte(",")},
				{Type: parser.ASTERISK, Literal: []byte("*")},
				{Type: parser.DBLCOLON, Literal: []byte("::")},
				{Type: parser.IDENT, Literal: []byte("after")},
				{Type: parser.LBRACE, Literal: []byte("{")},
				{Type: parser.IDENT, Literal: []byte("box-sizing")},
				{Type: parser.COLON, Literal: []byte(":")},
				{Type: parser.IDENT, Literal: []byte("border-box")},
				{Type: parser.SEMICOLON, Literal: []byte(";")},
				{Type: parser.RBRACE, Literal: []byte("}")},

				{Type: parser.IDENT, Literal: []byte("div")},
				{Type: parser.ASTERISK, Literal: []byte("*")},
				{Type: parser.IDENT, Literal: []byte("p")},
				{Type: parser.LBRACE, Literal: []byte("{")},
				{Type: parser.IDENT, Literal: []byte("color")},
				{Type: parser.COLON, Literal: []byte(":")},
				{Type: parser.IDENT, Literal: []byte("red")},
				{Type: parser.SEMICOLON, Literal: []byte(";")},
				{Type: parser.RBRACE, Literal: []byte("}")},

				{Type: parser.ASTERISK, Literal: []byte("*")},
				{Type: parser.LBRACKET, Literal: []byte("[")},
				{Type: parser.IDENT, Literal: []byte("lang")},
				{Type: parser.STARTS_WITH, Literal: []byte("^=")},
				{Type: parser.IDENT, Literal: []byte("en")},
				{Type: parser.RBRACKET, Literal: []byte("]")},

				{Type: parser.LBRACE, Literal: []byte("{")},
				{Type: parser.IDENT, Literal: []byte("color")},
				{Type: parser.COLON, Literal: []byte(":")},
				{Type: parser.IDENT, Literal: []byte("green")},
				{Type: parser.SEMICOLON, Literal: []byte(";")},
				{Type: parser.RBRACE, Literal: []byte("}")},

				{Type: parser.ASTERISK, Literal: []byte("*")},
				{Type: parser.DOT, Literal: []byte(".")},
				{Type: parser.IDENT, Literal: []byte("warning")},
				{Type: parser.LBRACE, Literal: []byte("{")},
				{Type: parser.IDENT, Literal: []byte("color")},
				{Type: parser.COLON, Literal: []byte(":")},
				{Type: parser.IDENT, Literal: []byte("yellow")},
				{Type: parser.SEMICOLON, Literal: []byte(";")},
				{Type: parser.RBRACE, Literal: []byte("}")},

				{Type: parser.ASTERISK, Literal: []byte("*")},
				{Type: parser.HASH, Literal: []byte("#")},
				{Type: parser.IDENT, Literal: []byte("myid")},
				{Type: parser.LBRACE, Literal: []byte("{")},
				{Type: parser.IDENT, Literal: []byte("font-weight")},
				{Type: parser.COLON, Literal: []byte(":")},
				{Type: parser.IDENT, Literal: []byte("bold")},
				{Type: parser.SEMICOLON, Literal: []byte(";")},
				{Type: parser.RBRACE, Literal: []byte("}")},
			},
		},
		{
			name:  "Combinators",
			input: "div > p + ul ~ span { color: red; }",
			expected: []parser.Token{
				{Type: parser.IDENT, Literal: []byte("div")},
				{Type: parser.GREATER, Literal: []byte(">")},
				{Type: parser.IDENT, Literal: []byte("p")},
				{Type: parser.PLUS, Literal: []byte("+")},
				{Type: parser.IDENT, Literal: []byte("ul")},
				{Type: parser.TILDE, Literal: []byte("~")},
				{Type: parser.IDENT, Literal: []byte("span")},
				{Type: parser.LBRACE, Literal: []byte("{")},
				{Type: parser.IDENT, Literal: []byte("color")},
				{Type: parser.COLON, Literal: []byte(":")},
				{Type: parser.IDENT, Literal: []byte("red")},
				{Type: parser.SEMICOLON, Literal: []byte(";")},
				{Type: parser.RBRACE, Literal: []byte("}")},
			},
		},
		{
			name:  "Pseudo-elements",
			input: "p::first-line { text-transform: uppercase; }",
			expected: []parser.Token{
				{Type: parser.IDENT, Literal: []byte("p")},
				{Type: parser.DBLCOLON, Literal: []byte("::")},
				{Type: parser.IDENT, Literal: []byte("first-line")},
				{Type: parser.LBRACE, Literal: []byte("{")},
				{Type: parser.IDENT, Literal: []byte("text-transform")},
				{Type: parser.COLON, Literal: []byte(":")},
				{Type: parser.IDENT, Literal: []byte("uppercase")},
				{Type: parser.SEMICOLON, Literal: []byte(";")},
				{Type: parser.RBRACE, Literal: []byte("}")},
			},
		},
		{
			name:  "Multiple Selectors",
			input: "h1, h2, h3 { font-family: sans-serif; }",
			expected: []parser.Token{
				{Type: parser.IDENT, Literal: []byte("h1")},
				{Type: parser.COMMA, Literal: []byte(",")},
				{Type: parser.IDENT, Literal: []byte("h2")},
				{Type: parser.COMMA, Literal: []byte(",")},
				{Type: parser.IDENT, Literal: []byte("h3")},
				{Type: parser.LBRACE, Literal: []byte("{")},
				{Type: parser.IDENT, Literal: []byte("font-family")},
				{Type: parser.COLON, Literal: []byte(":")},
				{Type: parser.IDENT, Literal: []byte("sans-serif")},
				{Type: parser.SEMICOLON, Literal: []byte(";")},
				{Type: parser.RBRACE, Literal: []byte("}")},
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
			expected: []parser.Token{
				{Type: parser.COLON, Literal: []byte(":")},
				{Type: parser.IDENT, Literal: []byte("root")},
				{Type: parser.LBRACE, Literal: []byte("{")},
				{Type: parser.IDENT, Literal: []byte("--main-color")},
				{Type: parser.COLON, Literal: []byte(":")},
				{Type: parser.IDENT, Literal: []byte("blue")},
				{Type: parser.SEMICOLON, Literal: []byte(";")},
				{Type: parser.RBRACE, Literal: []byte("}")},

				{Type: parser.IDENT, Literal: []byte("li")},
				{Type: parser.COLON, Literal: []byte(":")},
				{Type: parser.IDENT, Literal: []byte("nth-child")},
				{Type: parser.LPAREN, Literal: []byte("(")},
				{Type: parser.NUMBER, Literal: []byte("2")},
				{Type: parser.IDENT, Literal: []byte("n")},
				{Type: parser.PLUS, Literal: []byte("+")},
				{Type: parser.NUMBER, Literal: []byte("1")},
				{Type: parser.RPAREN, Literal: []byte(")")},
				{Type: parser.LBRACE, Literal: []byte("{")},
				{Type: parser.IDENT, Literal: []byte("background")},
				{Type: parser.COLON, Literal: []byte(":")},
				{Type: parser.IDENT, Literal: []byte("lightgray")},
				{Type: parser.SEMICOLON, Literal: []byte(";")},
				{Type: parser.RBRACE, Literal: []byte("}")},

				{Type: parser.IDENT, Literal: []byte("tr")},
				{Type: parser.COLON, Literal: []byte(":")},
				{Type: parser.IDENT, Literal: []byte("nth-child")},
				{Type: parser.LPAREN, Literal: []byte("(")},
				{Type: parser.IDENT, Literal: []byte("odd")},
				{Type: parser.RPAREN, Literal: []byte(")")},
				{Type: parser.LBRACE, Literal: []byte("{")},
				{Type: parser.IDENT, Literal: []byte("background-color")},
				{Type: parser.COLON, Literal: []byte(":")},
				{Type: parser.COLOR, Literal: []byte("#f2f2f2")},
				{Type: parser.SEMICOLON, Literal: []byte(";")},
				{Type: parser.RBRACE, Literal: []byte("}")},

				{Type: parser.IDENT, Literal: []byte("div")},
				{Type: parser.COLON, Literal: []byte(":")},
				{Type: parser.IDENT, Literal: []byte("nth-child")},
				{Type: parser.LPAREN, Literal: []byte("(")},
				{Type: parser.IDENT, Literal: []byte("-n")},
				{Type: parser.PLUS, Literal: []byte("+")},
				{Type: parser.NUMBER, Literal: []byte("3")},
				{Type: parser.RPAREN, Literal: []byte(")")},
				{Type: parser.LBRACE, Literal: []byte("{")},
				{Type: parser.IDENT, Literal: []byte("font-weight")},
				{Type: parser.COLON, Literal: []byte(":")},
				{Type: parser.IDENT, Literal: []byte("bold")},
				{Type: parser.SEMICOLON, Literal: []byte(";")},
				{Type: parser.RBRACE, Literal: []byte("}")},
			},
		},
		{
			name: "Pseudo-Elements",
			input: `
        p::first-line { font-weight: bold; }
        div::before { content: ""; display: block; }
    `,
			expected: []parser.Token{
				{Type: parser.IDENT, Literal: []byte("p")},
				{Type: parser.DBLCOLON, Literal: []byte("::")},
				{Type: parser.IDENT, Literal: []byte("first-line")},
				{Type: parser.LBRACE, Literal: []byte("{")},
				{Type: parser.IDENT, Literal: []byte("font-weight")},
				{Type: parser.COLON, Literal: []byte(":")},
				{Type: parser.IDENT, Literal: []byte("bold")},
				{Type: parser.SEMICOLON, Literal: []byte(";")},
				{Type: parser.RBRACE, Literal: []byte("}")},

				{Type: parser.IDENT, Literal: []byte("div")},
				{Type: parser.DBLCOLON, Literal: []byte("::")},
				{Type: parser.IDENT, Literal: []byte("before")},
				{Type: parser.LBRACE, Literal: []byte("{")},
				{Type: parser.IDENT, Literal: []byte("content")},
				{Type: parser.COLON, Literal: []byte(":")},
				{Type: parser.STRING, Literal: []byte("\"\"")},
				{Type: parser.SEMICOLON, Literal: []byte(";")},
				{Type: parser.IDENT, Literal: []byte("display")},
				{Type: parser.COLON, Literal: []byte(":")},
				{Type: parser.IDENT, Literal: []byte("block")},
				{Type: parser.SEMICOLON, Literal: []byte(";")},
				{Type: parser.RBRACE, Literal: []byte("}")},
			},
		},
		{
			name:  "Media Query",
			input: "@media screen and (max-width: 600px) { body { font-size: 14px; } }",
			expected: []parser.Token{
				{Type: parser.AT, Literal: []byte("@")},
				{Type: parser.IDENT, Literal: []byte("media")},
				{Type: parser.IDENT, Literal: []byte("screen")},
				{Type: parser.IDENT, Literal: []byte("and")},
				{Type: parser.LPAREN, Literal: []byte("(")},
				{Type: parser.IDENT, Literal: []byte("max-width")},
				{Type: parser.COLON, Literal: []byte(":")},
				{Type: parser.NUMBER, Literal: []byte("600")},
				{Type: parser.IDENT, Literal: []byte("px")},
				{Type: parser.RPAREN, Literal: []byte(")")},
				{Type: parser.LBRACE, Literal: []byte("{")},
				{Type: parser.IDENT, Literal: []byte("body")},
				{Type: parser.LBRACE, Literal: []byte("{")},
				{Type: parser.IDENT, Literal: []byte("font-size")},
				{Type: parser.COLON, Literal: []byte(":")},
				{Type: parser.NUMBER, Literal: []byte("14")},
				{Type: parser.IDENT, Literal: []byte("px")},
				{Type: parser.SEMICOLON, Literal: []byte(";")},
				{Type: parser.RBRACE, Literal: []byte("}")},
				{Type: parser.RBRACE, Literal: []byte("}")},
			},
		},
		{
			name:  "Complex Selectors",
			input: "#main > .article p:first-child { color: #ff0000; }",
			expected: []parser.Token{
				{Type: parser.HASH, Literal: []byte("#")},
				{Type: parser.IDENT, Literal: []byte("main")},
				{Type: parser.GREATER, Literal: []byte(">")},
				{Type: parser.DOT, Literal: []byte(".")},
				{Type: parser.IDENT, Literal: []byte("article")},
				{Type: parser.IDENT, Literal: []byte("p")},
				{Type: parser.COLON, Literal: []byte(":")},
				{Type: parser.IDENT, Literal: []byte("first-child")},
				{Type: parser.LBRACE, Literal: []byte("{")},
				{Type: parser.IDENT, Literal: []byte("color")},
				{Type: parser.COLON, Literal: []byte(":")},
				{Type: parser.COLOR, Literal: []byte("#ff0000")},
				{Type: parser.SEMICOLON, Literal: []byte(";")},
				{Type: parser.RBRACE, Literal: []byte("}")},
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
			expected: []parser.Token{
				{Type: parser.IDENT, Literal: []byte("main")},
				{Type: parser.LBRACE, Literal: []byte("{")},
				{Type: parser.IDENT, Literal: []byte("color")},
				{Type: parser.COLON, Literal: []byte(":")},
				{Type: parser.COLOR, Literal: []byte("#ff0000")},
				{Type: parser.EXCLAMATION, Literal: []byte("!")},
				{Type: parser.IDENT, Literal: []byte("important")},
				{Type: parser.SEMICOLON, Literal: []byte(";")},
				{Type: parser.RBRACE, Literal: []byte("}")},

				{Type: parser.IDENT, Literal: []byte("p")},
				{Type: parser.LBRACE, Literal: []byte("{")},
				{Type: parser.IDENT, Literal: []byte("margin")},
				{Type: parser.COLON, Literal: []byte(":")},
				{Type: parser.NUMBER, Literal: []byte("10")},
				{Type: parser.IDENT, Literal: []byte("px")},
				{Type: parser.EXCLAMATION, Literal: []byte("!")},
				{Type: parser.IDENT, Literal: []byte("important")},
				{Type: parser.SEMICOLON, Literal: []byte(";")},
				{Type: parser.RBRACE, Literal: []byte("}")},

				{Type: parser.IDENT, Literal: []byte("div")},
				{Type: parser.LBRACE, Literal: []byte("{")},
				{Type: parser.IDENT, Literal: []byte("padding")},
				{Type: parser.COLON, Literal: []byte(":")},
				{Type: parser.NUMBER, Literal: []byte("5")},
				{Type: parser.IDENT, Literal: []byte("px")},
				{Type: parser.EXCLAMATION, Literal: []byte("!")},
				{Type: parser.IDENT, Literal: []byte("important")},
				{Type: parser.SEMICOLON, Literal: []byte(";")},
				{Type: parser.RBRACE, Literal: []byte("}")},

				{Type: parser.IDENT, Literal: []byte("span")},
				{Type: parser.LBRACE, Literal: []byte("{")},
				{Type: parser.IDENT, Literal: []byte("font-weight")},
				{Type: parser.COLON, Literal: []byte(":")},
				{Type: parser.IDENT, Literal: []byte("bold")},
				{Type: parser.EXCLAMATION, Literal: []byte("!")},
				{Type: parser.IDENT, Literal: []byte("important")},
				{Type: parser.SEMICOLON, Literal: []byte(";")},
				{Type: parser.RBRACE, Literal: []byte("}")},

				{Type: parser.DOT, Literal: []byte(".")},
				{Type: parser.IDENT, Literal: []byte("custom")},
				{Type: parser.LBRACE, Literal: []byte("{")},
				{Type: parser.IDENT, Literal: []byte("border")},
				{Type: parser.COLON, Literal: []byte(":")},
				{Type: parser.NUMBER, Literal: []byte("1")},
				{Type: parser.IDENT, Literal: []byte("px")},
				{Type: parser.IDENT, Literal: []byte("solid")},
				{Type: parser.IDENT, Literal: []byte("black")},
				{Type: parser.EXCLAMATION, Literal: []byte("!")},
				{Type: parser.IDENT, Literal: []byte("important")},
				{Type: parser.RBRACE, Literal: []byte("}")},
			},
		},
		{
			name:  "Attribute Selector",
			input: "a[href^=\"https://\"] { color: green; }",
			expected: []parser.Token{
				{Type: parser.IDENT, Literal: []byte("a")},
				{Type: parser.LBRACKET, Literal: []byte("[")},
				{Type: parser.IDENT, Literal: []byte("href")},
				{Type: parser.STARTS_WITH, Literal: []byte("^=")},
				{Type: parser.STRING, Literal: []byte("\"https://\"")},
				{Type: parser.RBRACKET, Literal: []byte("]")},

				{Type: parser.LBRACE, Literal: []byte("{")},
				{Type: parser.IDENT, Literal: []byte("color")},
				{Type: parser.COLON, Literal: []byte(":")},
				{Type: parser.IDENT, Literal: []byte("green")},
				{Type: parser.SEMICOLON, Literal: []byte(";")},
				{Type: parser.RBRACE, Literal: []byte("}")},
			},
		},
		{
			name:  "Keyframes",
			input: "@keyframes fadeIn { 0% { opacity: 0; } 100% { opacity: 1; } }",
			expected: []parser.Token{
				{Type: parser.AT, Literal: []byte("@")},
				{Type: parser.IDENT, Literal: []byte("keyframes")},
				{Type: parser.IDENT, Literal: []byte("fadeIn")},
				{Type: parser.LBRACE, Literal: []byte("{")},
				{Type: parser.NUMBER, Literal: []byte("0")},
				{Type: parser.PERCENTAGE, Literal: []byte("%")},
				{Type: parser.LBRACE, Literal: []byte("{")},
				{Type: parser.IDENT, Literal: []byte("opacity")},
				{Type: parser.COLON, Literal: []byte(":")},
				{Type: parser.NUMBER, Literal: []byte("0")},
				{Type: parser.SEMICOLON, Literal: []byte(";")},
				{Type: parser.RBRACE, Literal: []byte("}")},
				{Type: parser.NUMBER, Literal: []byte("100")},
				{Type: parser.PERCENTAGE, Literal: []byte("%")},
				{Type: parser.LBRACE, Literal: []byte("{")},
				{Type: parser.IDENT, Literal: []byte("opacity")},
				{Type: parser.COLON, Literal: []byte(":")},
				{Type: parser.NUMBER, Literal: []byte("1")},
				{Type: parser.SEMICOLON, Literal: []byte(";")},
				{Type: parser.RBRACE, Literal: []byte("}")},
				{Type: parser.RBRACE, Literal: []byte("}")},
			},
		},
		{
			name:  "CSS Variables",
			input: ":root { --main-color: blue; } body { color: var(--main-color); }",
			expected: []parser.Token{
				{Type: parser.COLON, Literal: []byte(":")},
				{Type: parser.IDENT, Literal: []byte("root")},
				{Type: parser.LBRACE, Literal: []byte("{")},
				{Type: parser.IDENT, Literal: []byte("--main-color")},
				{Type: parser.COLON, Literal: []byte(":")},
				{Type: parser.IDENT, Literal: []byte("blue")},
				{Type: parser.SEMICOLON, Literal: []byte(";")},
				{Type: parser.RBRACE, Literal: []byte("}")},

				{Type: parser.IDENT, Literal: []byte("body")},
				{Type: parser.LBRACE, Literal: []byte("{")},
				{Type: parser.IDENT, Literal: []byte("color")},
				{Type: parser.COLON, Literal: []byte(":")},
				{Type: parser.IDENT, Literal: []byte("var")},
				{Type: parser.LPAREN, Literal: []byte("(")},
				{Type: parser.IDENT, Literal: []byte("--main-color")},
				{Type: parser.RPAREN, Literal: []byte(")")},
				{Type: parser.SEMICOLON, Literal: []byte(";")},
				{Type: parser.RBRACE, Literal: []byte("}")},
			},
		},
		{
			name:  "Calc Function",
			input: "div { width: calc(100% - 20px); height: 100vh; }",
			expected: []parser.Token{
				{Type: parser.IDENT, Literal: []byte("div")},
				{Type: parser.LBRACE, Literal: []byte("{")},
				{Type: parser.IDENT, Literal: []byte("width")},
				{Type: parser.COLON, Literal: []byte(":")},
				{Type: parser.IDENT, Literal: []byte("calc")},
				{Type: parser.LPAREN, Literal: []byte("(")},
				{Type: parser.NUMBER, Literal: []byte("100")},
				{Type: parser.PERCENTAGE, Literal: []byte("%")},
				{Type: parser.MINUS, Literal: []byte("-")},
				{Type: parser.NUMBER, Literal: []byte("20")},
				{Type: parser.IDENT, Literal: []byte("px")},
				{Type: parser.RPAREN, Literal: []byte(")")},
				{Type: parser.SEMICOLON, Literal: []byte(";")},
				{Type: parser.IDENT, Literal: []byte("height")},
				{Type: parser.COLON, Literal: []byte(":")},
				{Type: parser.NUMBER, Literal: []byte("100")},
				{Type: parser.IDENT, Literal: []byte("vh")},
				{Type: parser.SEMICOLON, Literal: []byte(";")},
				{Type: parser.RBRACE, Literal: []byte("}")},
			},
		},
		{
			name:  "Escaped characters in identifiers",
			input: ".foo\\.bar { color: red; }",
			expected: []parser.Token{
				{Type: parser.DOT, Literal: []byte(".")},
				{Type: parser.IDENT, Literal: []byte("foo\\.bar")},
				{Type: parser.LBRACE, Literal: []byte("{")},
				{Type: parser.IDENT, Literal: []byte("color")},
				{Type: parser.COLON, Literal: []byte(":")},
				{Type: parser.IDENT, Literal: []byte("red")},
				{Type: parser.SEMICOLON, Literal: []byte(";")},
				{Type: parser.RBRACE, Literal: []byte("}")},
			},
		},
		{
			name:  "Escaped characters in attribute selectors",
			input: "a[href=\"foo\\\"bar\"] { color: blue; }",
			expected: []parser.Token{
				{Type: parser.IDENT, Literal: []byte("a")},
				{Type: parser.LBRACKET, Literal: []byte("[")},
				{Type: parser.IDENT, Literal: []byte("href")},
				{Type: parser.EQUALS, Literal: []byte("=")},
				{Type: parser.STRING, Literal: []byte("\"foo\\\"bar\"")},
				{Type: parser.RBRACKET, Literal: []byte("]")},
				{Type: parser.LBRACE, Literal: []byte("{")},
				{Type: parser.IDENT, Literal: []byte("color")},
				{Type: parser.COLON, Literal: []byte(":")},
				{Type: parser.IDENT, Literal: []byte("blue")},
				{Type: parser.SEMICOLON, Literal: []byte(";")},
				{Type: parser.RBRACE, Literal: []byte("}")},
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
			expected: []parser.Token{
				{Type: parser.HASH, Literal: []byte("#")},
				{Type: parser.IDENT, Literal: []byte("☃")},
				{Type: parser.LBRACE, Literal: []byte("{")},
				{Type: parser.IDENT, Literal: []byte("color")},
				{Type: parser.COLON, Literal: []byte(":")},
				{Type: parser.IDENT, Literal: []byte("skyblue")},
				{Type: parser.SEMICOLON, Literal: []byte(";")},
				{Type: parser.RBRACE, Literal: []byte("}")},
				{Type: parser.DOT, Literal: []byte(".")},
				{Type: parser.IDENT, Literal: []byte("günther")},
				{Type: parser.COMMA, Literal: []byte(",")},
				{Type: parser.HASH, Literal: []byte("#")},
				{Type: parser.IDENT, Literal: []byte("π_value")},
				{Type: parser.COMMA, Literal: []byte(",")},
				{Type: parser.DOT, Literal: []byte(".")},
				{Type: parser.IDENT, Literal: []byte("こんにちは")},
				{Type: parser.LBRACE, Literal: []byte("{")},
				{Type: parser.IDENT, Literal: []byte("color")},
				{Type: parser.COLON, Literal: []byte(":")},
				{Type: parser.IDENT, Literal: []byte("blue")},
				{Type: parser.SEMICOLON, Literal: []byte(";")},
				{Type: parser.RBRACE, Literal: []byte("}")},

				{Type: parser.DOT, Literal: []byte(".")},
				{Type: parser.IDENT, Literal: []byte("--custom-prop")},
				{Type: parser.COMMA, Literal: []byte(",")},
				{Type: parser.DOT, Literal: []byte(".")},
				{Type: parser.IDENT, Literal: []byte("-moz-custom")},
				{Type: parser.LBRACE, Literal: []byte("{")},
				{Type: parser.IDENT, Literal: []byte("value")},
				{Type: parser.COLON, Literal: []byte(":")},
				{Type: parser.NUMBER, Literal: []byte("123")},
				{Type: parser.SEMICOLON, Literal: []byte(";")},
				{Type: parser.RBRACE, Literal: []byte("}")},

				{Type: parser.HASH, Literal: []byte("#")},
				{Type: parser.IDENT, Literal: []byte("\\26 ABC")},
				{Type: parser.LBRACE, Literal: []byte("{")},
				{Type: parser.COMMENT, Literal: []byte("/* Escaped ASCII */")},
				{Type: parser.IDENT, Literal: []byte("color")},
				{Type: parser.COLON, Literal: []byte(":")},
				{Type: parser.IDENT, Literal: []byte("red")},
				{Type: parser.SEMICOLON, Literal: []byte(";")},
				{Type: parser.RBRACE, Literal: []byte("}")},
			},
		},
		{
			name:  "CSS custom properties (variables)",
			input: ":root { --custom-color: #ff00ff; } .foo { color: var(--custom-color); }",
			expected: []parser.Token{
				{Type: parser.COLON, Literal: []byte(":")},
				{Type: parser.IDENT, Literal: []byte("root")},
				{Type: parser.LBRACE, Literal: []byte("{")},
				{Type: parser.IDENT, Literal: []byte("--custom-color")},
				{Type: parser.COLON, Literal: []byte(":")},
				{Type: parser.COLOR, Literal: []byte("#ff00ff")},
				{Type: parser.SEMICOLON, Literal: []byte(";")},
				{Type: parser.RBRACE, Literal: []byte("}")},

				{Type: parser.DOT, Literal: []byte(".")},
				{Type: parser.IDENT, Literal: []byte("foo")},
				{Type: parser.LBRACE, Literal: []byte("{")},
				{Type: parser.IDENT, Literal: []byte("color")},
				{Type: parser.COLON, Literal: []byte(":")},
				{Type: parser.IDENT, Literal: []byte("var")},
				{Type: parser.LPAREN, Literal: []byte("(")},
				{Type: parser.IDENT, Literal: []byte("--custom-color")},
				{Type: parser.RPAREN, Literal: []byte(")")},
				{Type: parser.SEMICOLON, Literal: []byte(";")},
				{Type: parser.RBRACE, Literal: []byte("}")},
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
			expected: []parser.Token{
				{Type: parser.DOT, Literal: []byte(".")},
				{Type: parser.IDENT, Literal: []byte("gradient")},
				{Type: parser.LBRACE, Literal: []byte("{")},
				{Type: parser.IDENT, Literal: []byte("background-image")},
				{Type: parser.COLON, Literal: []byte(":")},
				{Type: parser.IDENT, Literal: []byte("linear-gradient")},
				{Type: parser.LPAREN, Literal: []byte("(")},
				{Type: parser.NUMBER, Literal: []byte("45")},
				{Type: parser.IDENT, Literal: []byte("deg")},
				{Type: parser.COMMA, Literal: []byte(",")},
				{Type: parser.COLOR, Literal: []byte("#ff0000")},
				{Type: parser.COMMA, Literal: []byte(",")},
				{Type: parser.COLOR, Literal: []byte("#00ff00")},
				{Type: parser.RPAREN, Literal: []byte(")")},
				{Type: parser.COMMA, Literal: []byte(",")},
				{Type: parser.IDENT, Literal: []byte("radial-gradient")},
				{Type: parser.LPAREN, Literal: []byte("(")},
				{Type: parser.IDENT, Literal: []byte("circle")},
				{Type: parser.COMMA, Literal: []byte(",")},
				{Type: parser.COLOR, Literal: []byte("#0000ff")},
				{Type: parser.COMMA, Literal: []byte(",")},
				{Type: parser.COLOR, Literal: []byte("#ffff00")},
				{Type: parser.RPAREN, Literal: []byte(")")},
				{Type: parser.SEMICOLON, Literal: []byte(";")},
				{Type: parser.IDENT, Literal: []byte("font")},
				{Type: parser.COLON, Literal: []byte(":")},
				{Type: parser.IDENT, Literal: []byte("bold")},
				{Type: parser.NUMBER, Literal: []byte("12")},
				{Type: parser.IDENT, Literal: []byte("px")},
				{Type: parser.DIVIDE, Literal: []byte("/")},
				{Type: parser.NUMBER, Literal: []byte("14")},
				{Type: parser.IDENT, Literal: []byte("px")},
				{Type: parser.STRING, Literal: []byte("\"Helvetica\"")},
				{Type: parser.COMMA, Literal: []byte(",")},
				{Type: parser.IDENT, Literal: []byte("sans-serif")},
				{Type: parser.SEMICOLON, Literal: []byte(";")},
				{Type: parser.RBRACE, Literal: []byte("}")},
			},
		},
		{
			name: "CSS Hacks and Legacy Browser Support",
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
			expected: []parser.Token{
				{Type: parser.COMMENT, Literal: []byte("/* Pseudo-element hack */")},
				{Type: parser.IDENT, Literal: []byte("_")},
				{Type: parser.DBLCOLON, Literal: []byte("::")},
				{Type: parser.IDENT, Literal: []byte("-webkit-full-page-media")},
				{Type: parser.COMMA, Literal: []byte(",")},
				{Type: parser.IDENT, Literal: []byte("_")},
				{Type: parser.COLON, Literal: []byte(":")},
				{Type: parser.IDENT, Literal: []byte("future")},
				{Type: parser.COMMA, Literal: []byte(",")},
				{Type: parser.COLON, Literal: []byte(":")},
				{Type: parser.IDENT, Literal: []byte("root")},
				{Type: parser.DOT, Literal: []byte(".")},
				{Type: parser.IDENT, Literal: []byte("foo")},
				{Type: parser.LBRACE, Literal: []byte("{")},
				{Type: parser.IDENT, Literal: []byte("color")},
				{Type: parser.COLON, Literal: []byte(":")},
				{Type: parser.IDENT, Literal: []byte("red")},
				{Type: parser.SEMICOLON, Literal: []byte(";")},
				{Type: parser.RBRACE, Literal: []byte("}")},

				{Type: parser.COMMENT, Literal: []byte("/* Pseudo-class hack */")},
				{Type: parser.IDENT, Literal: []byte("_")},
				{Type: parser.COLON, Literal: []byte(":")},
				{Type: parser.IDENT, Literal: []byte("-webkit-full-screen")},
				{Type: parser.COMMA, Literal: []byte(",")},
				{Type: parser.COLON, Literal: []byte(":")},
				{Type: parser.IDENT, Literal: []byte("root")},
				{Type: parser.DOT, Literal: []byte(".")},
				{Type: parser.IDENT, Literal: []byte("bar")},
				{Type: parser.LBRACE, Literal: []byte("{")},
				{Type: parser.IDENT, Literal: []byte("display")},
				{Type: parser.COLON, Literal: []byte(":")},
				{Type: parser.IDENT, Literal: []byte("block")},
				{Type: parser.SEMICOLON, Literal: []byte(";")},
				{Type: parser.RBRACE, Literal: []byte("}")},

				{Type: parser.COMMENT, Literal: []byte("/* Media query hack */")},
				{Type: parser.AT, Literal: []byte("@")},
				{Type: parser.IDENT, Literal: []byte("media")},
				{Type: parser.IDENT, Literal: []byte("screen")},
				{Type: parser.IDENT, Literal: []byte("and")},
				{Type: parser.LPAREN, Literal: []byte("(")},
				{Type: parser.IDENT, Literal: []byte("min-width")},
				{Type: parser.COLON, Literal: []byte(":")},
				{Type: parser.NUMBER, Literal: []byte("0")},
				{Type: parser.IDENT, Literal: []byte("\\0")},
				{Type: parser.RPAREN, Literal: []byte(")")},
				{Type: parser.LBRACE, Literal: []byte("{")},
				{Type: parser.DOT, Literal: []byte(".")},
				{Type: parser.IDENT, Literal: []byte("baz")},
				{Type: parser.LBRACE, Literal: []byte("{")},
				{Type: parser.IDENT, Literal: []byte("zoom")},
				{Type: parser.COLON, Literal: []byte(":")},
				{Type: parser.NUMBER, Literal: []byte("1")},
				{Type: parser.SEMICOLON, Literal: []byte(";")},
				{Type: parser.RBRACE, Literal: []byte("}")},
				{Type: parser.RBRACE, Literal: []byte("}")},

				{Type: parser.COMMENT, Literal: []byte("/* Property value hacks */")},
				{Type: parser.DOT, Literal: []byte(".")},
				{Type: parser.IDENT, Literal: []byte("hack1")},
				{Type: parser.LBRACE, Literal: []byte("{")},
				{Type: parser.IDENT, Literal: []byte("property")},
				{Type: parser.COLON, Literal: []byte(":")},
				{Type: parser.IDENT, Literal: []byte("value\\9")},
				{Type: parser.SEMICOLON, Literal: []byte(";")},
				{Type: parser.RBRACE, Literal: []byte("}")},

				{Type: parser.DOT, Literal: []byte(".")},
				{Type: parser.IDENT, Literal: []byte("hack2")},
				{Type: parser.LBRACE, Literal: []byte("{")},
				{Type: parser.IDENT, Literal: []byte("property")},
				{Type: parser.COLON, Literal: []byte(":")},
				{Type: parser.IDENT, Literal: []byte("value\\9")},
				{Type: parser.SEMICOLON, Literal: []byte(";")},
				{Type: parser.RBRACE, Literal: []byte("}")},

				{Type: parser.DOT, Literal: []byte(".")},
				{Type: parser.IDENT, Literal: []byte("hack3")},
				{Type: parser.LBRACE, Literal: []byte("{")},
				{Type: parser.IDENT, Literal: []byte("property")},
				{Type: parser.COLON, Literal: []byte(":")},
				{Type: parser.IDENT, Literal: []byte("value\\0")},
				{Type: parser.SEMICOLON, Literal: []byte(";")},
				{Type: parser.RBRACE, Literal: []byte("}")},

				{Type: parser.DOT, Literal: []byte(".")},
				{Type: parser.IDENT, Literal: []byte("hack4")},
				{Type: parser.LBRACE, Literal: []byte("{")},
				{Type: parser.IDENT, Literal: []byte("property")},
				{Type: parser.COLON, Literal: []byte(":")},
				{Type: parser.IDENT, Literal: []byte("value\\0")},
				{Type: parser.SEMICOLON, Literal: []byte(";")},
				{Type: parser.RBRACE, Literal: []byte("}")},

				{Type: parser.COMMENT, Literal: []byte("/* Combination of hacks */")},
				{Type: parser.IDENT, Literal: []byte("_")},
				{Type: parser.COLON, Literal: []byte(":")},
				{Type: parser.IDENT, Literal: []byte("-ms-lang")},
				{Type: parser.LPAREN, Literal: []byte("(")},
				{Type: parser.IDENT, Literal: []byte("x")},
				{Type: parser.RPAREN, Literal: []byte(")")},
				{Type: parser.COMMA, Literal: []byte(",")},
				{Type: parser.IDENT, Literal: []byte("_")},
				{Type: parser.COLON, Literal: []byte(":")},
				{Type: parser.IDENT, Literal: []byte("-webkit-full-screen")},
				{Type: parser.COMMA, Literal: []byte(",")},
				{Type: parser.DOT, Literal: []byte(".")},
				{Type: parser.IDENT, Literal: []byte("multi-hack")},
				{Type: parser.LBRACE, Literal: []byte("{")},
				{Type: parser.ASTERISK, Literal: []byte("*")},
				{Type: parser.IDENT, Literal: []byte("display")},
				{Type: parser.COLON, Literal: []byte(":")},
				{Type: parser.IDENT, Literal: []byte("inline")},
				{Type: parser.SEMICOLON, Literal: []byte(";")},
				{Type: parser.IDENT, Literal: []byte("_height")},
				{Type: parser.COLON, Literal: []byte(":")},
				{Type: parser.NUMBER, Literal: []byte("1")},
				{Type: parser.PERCENTAGE, Literal: []byte("%")},
				{Type: parser.SEMICOLON, Literal: []byte(";")},
				{Type: parser.RBRACE, Literal: []byte("}")},
			},
		},
	}

	for _, tt := range tests {
		if tt.ignore {
			continue
		}

		t.Run(tt.name, func(t *testing.T) {
			l := parser.Read(strings.NewReader(tt.input))
			defer l.Release()
			for i, expected := range tt.expected {

				tok := l.NextToken()
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
			// Check for unexpected additional tokens
			if token := l.NextToken(); token.Type != parser.EOF {
				t.Errorf("Expected EOF, got %v", token.Type)
			}
		})
	}
}

func TestIllegalCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []parser.Token
	}{
		{
			name:  "Invalid hex color",
			input: ".invalid { color: #1234ZZ; }",
			expected: []parser.Token{
				{Type: parser.DOT, Literal: []byte(".")},
				{Type: parser.IDENT, Literal: []byte("invalid")},
				{Type: parser.LBRACE, Literal: []byte("{")},
				{Type: parser.IDENT, Literal: []byte("color")},
				{Type: parser.COLON, Literal: []byte(":")},
				{Type: parser.HASH, Literal: []byte("#")},
				{Type: parser.NUMBER, Literal: []byte("1234")},
				{Type: parser.IDENT, Literal: []byte("ZZ")},
				{Type: parser.SEMICOLON, Literal: []byte(";")},
				{Type: parser.RBRACE, Literal: []byte("}")},
			},
		},
		{
			name:  "Invalid unit combination",
			input: ".invalid-unit { width: 50+px; }",
			expected: []parser.Token{
				{Type: parser.DOT, Literal: []byte(".")},
				{Type: parser.IDENT, Literal: []byte("invalid-unit")},
				{Type: parser.LBRACE, Literal: []byte("{")},
				{Type: parser.IDENT, Literal: []byte("width")},
				{Type: parser.COLON, Literal: []byte(":")},
				{Type: parser.NUMBER, Literal: []byte("50")},
				{Type: parser.PLUS, Literal: []byte("+")},
				{Type: parser.IDENT, Literal: []byte("px")},
				{Type: parser.SEMICOLON, Literal: []byte(";")},
				{Type: parser.RBRACE, Literal: []byte("}")},
			},
		},
		{
			name:  "Invalid percentage",
			input: ".invalid-percentage { height: 100vh%; }",
			expected: []parser.Token{
				{Type: parser.DOT, Literal: []byte(".")},
				{Type: parser.IDENT, Literal: []byte("invalid-percentage")},
				{Type: parser.LBRACE, Literal: []byte("{")},
				{Type: parser.IDENT, Literal: []byte("height")},
				{Type: parser.COLON, Literal: []byte(":")},
				{Type: parser.NUMBER, Literal: []byte("100")},
				{Type: parser.IDENT, Literal: []byte("vh")},
				{Type: parser.PERCENTAGE, Literal: []byte("%")},
				{Type: parser.SEMICOLON, Literal: []byte(";")},
				{Type: parser.RBRACE, Literal: []byte("}")},
			},
		},
		{
			name:  "Invalid property",
			input: ".invalid-property { colo r: red; }",
			expected: []parser.Token{
				{Type: parser.DOT, Literal: []byte(".")},
				{Type: parser.IDENT, Literal: []byte("invalid-property")},
				{Type: parser.LBRACE, Literal: []byte("{")},
				{Type: parser.IDENT, Literal: []byte("colo")},
				{Type: parser.IDENT, Literal: []byte("r")},
				{Type: parser.COLON, Literal: []byte(":")},
				{Type: parser.IDENT, Literal: []byte("red")},
				{Type: parser.SEMICOLON, Literal: []byte(";")},
				{Type: parser.RBRACE, Literal: []byte("}")},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := parser.Read(strings.NewReader(tt.input))
			defer l.Release()
			for i, expected := range tt.expected {
				got := l.NextToken()
				if got.Type != expected.Type {
					t.Errorf("Token %d: expected type %v, got %v", i, expected.Type, got.Type)
				}
				if !bytesEqual(got.Literal, expected.Literal) {
					t.Errorf("Token %d: expected literal %q, got %q", i, string(expected.Literal), string(got.Literal))
				}
			}
			// Check for unexpected additional tokens
			if token := l.NextToken(); token.Type != parser.EOF {
				t.Errorf("Expected EOF, got %v", token.Type)
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
			filepath: filepath.Join("..", "test-data", "frameworks", "bootstrap.css"),
			expected: 0,
		},
		{
			name:     "Can Lex Bulma css without ILLEGALS",
			filepath: filepath.Join("..", "test-data", "frameworks", "bulma.css"),
			expected: 0,
		},
		{
			name:     "Can Lex Foundation css without ILLEGALS",
			filepath: filepath.Join("..", "test-data", "frameworks", "foundation.css"),
			expected: 0,
		},
		{
			name:     "Can Lex Materialize css without ILLEGALS",
			filepath: filepath.Join("..", "test-data", "frameworks", "materialize.css"),
			expected: 0,
		},
		{
			name:     "Can Lex Spectre css without ILLEGALS",
			filepath: filepath.Join("..", "test-data", "frameworks", "spectre.css"),
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

			l := parser.Read(input)
			defer l.Release()
			illegalCount := 0
			for tok := l.NextToken(); tok.Type != parser.EOF; tok = l.NextToken() {
				if tok.Type == parser.ILLEGAL {
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
		{"Bootstrap", filepath.Join("..", "test-data", "frameworks", "bootstrap.css")},
		{"Bulma", filepath.Join("..", "test-data", "frameworks", "bulma.css")},
		{"Foundation", filepath.Join("..", "test-data", "frameworks", "foundation.css")},
		{"Materialize", filepath.Join("..", "test-data", "frameworks", "materialize.css")},
		{"Spectre", filepath.Join("..", "test-data", "frameworks", "spectre.css")},
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
				l := parser.Read(reader)
				defer l.Release()
				for tok := l.NextToken(); tok.Type != parser.EOF; tok = l.NextToken() {
					// Do nothing, just lex
				}
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
