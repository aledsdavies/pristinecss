package parser_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aledsdavies/pristinecss/parser"
)

func TestPositiveCases(t *testing.T) {
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
				{Type: parser.COMMENT, Literal: []rune("/* This is a comment */")},
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
				{Type: parser.SELECTOR, Literal: []rune("*")},
				{Type: parser.LBRACE, Literal: []rune("{")},
				{Type: parser.IDENT, Literal: []rune("margin")},
				{Type: parser.COLON, Literal: []rune(":")},
				{Type: parser.NUMBER, Literal: []rune("0")},
				{Type: parser.SEMICOLON, Literal: []rune(";")},
				{Type: parser.IDENT, Literal: []rune("padding")},
				{Type: parser.COLON, Literal: []rune(":")},
				{Type: parser.NUMBER, Literal: []rune("0")},
				{Type: parser.SEMICOLON, Literal: []rune(";")},
				{Type: parser.RBRACE, Literal: []rune("}")},

				{Type: parser.SELECTOR, Literal: []rune("*")},
				{Type: parser.COMMA, Literal: []rune(",")},
				{Type: parser.SELECTOR, Literal: []rune("*")},
				{Type: parser.PSEUDO_ELEMENT, Literal: []rune("::before")},
				{Type: parser.COMMA, Literal: []rune(",")},
				{Type: parser.SELECTOR, Literal: []rune("*")},
				{Type: parser.PSEUDO_ELEMENT, Literal: []rune("::after")},
				{Type: parser.LBRACE, Literal: []rune("{")},
				{Type: parser.IDENT, Literal: []rune("box-sizing")},
				{Type: parser.COLON, Literal: []rune(":")},
				{Type: parser.IDENT, Literal: []rune("border-box")},
				{Type: parser.SEMICOLON, Literal: []rune(";")},
				{Type: parser.RBRACE, Literal: []rune("}")},

				{Type: parser.SELECTOR, Literal: []rune("div")},
				{Type: parser.SELECTOR, Literal: []rune("*")},
				{Type: parser.SELECTOR, Literal: []rune("p")},
				{Type: parser.LBRACE, Literal: []rune("{")},
				{Type: parser.IDENT, Literal: []rune("color")},
				{Type: parser.COLON, Literal: []rune(":")},
				{Type: parser.IDENT, Literal: []rune("red")},
				{Type: parser.SEMICOLON, Literal: []rune(";")},
				{Type: parser.RBRACE, Literal: []rune("}")},

				{Type: parser.SELECTOR, Literal: []rune("*[lang^=en]")},
				{Type: parser.LBRACE, Literal: []rune("{")},
				{Type: parser.IDENT, Literal: []rune("color")},
				{Type: parser.COLON, Literal: []rune(":")},
				{Type: parser.IDENT, Literal: []rune("green")},
				{Type: parser.SEMICOLON, Literal: []rune(";")},
				{Type: parser.RBRACE, Literal: []rune("}")},

				{Type: parser.SELECTOR, Literal: []rune("*.warning")},
				{Type: parser.LBRACE, Literal: []rune("{")},
				{Type: parser.IDENT, Literal: []rune("color")},
				{Type: parser.COLON, Literal: []rune(":")},
				{Type: parser.IDENT, Literal: []rune("yellow")},
				{Type: parser.SEMICOLON, Literal: []rune(";")},
				{Type: parser.RBRACE, Literal: []rune("}")},

				{Type: parser.SELECTOR, Literal: []rune("*#myid")},
				{Type: parser.LBRACE, Literal: []rune("{")},
				{Type: parser.IDENT, Literal: []rune("font-weight")},
				{Type: parser.COLON, Literal: []rune(":")},
				{Type: parser.IDENT, Literal: []rune("bold")},
				{Type: parser.SEMICOLON, Literal: []rune(";")},
				{Type: parser.RBRACE, Literal: []rune("}")},
			},
		},
		{
			name:  "Combinators",
			input: "div > p + ul ~ span { color: red; }",
			expected: []parser.Token{
				{Type: parser.SELECTOR, Literal: []rune("div")},
				{Type: parser.GREATER, Literal: []rune(">")},
				{Type: parser.SELECTOR, Literal: []rune("p")},
				{Type: parser.PLUS, Literal: []rune("+")},
				{Type: parser.SELECTOR, Literal: []rune("ul")},
				{Type: parser.TILDE, Literal: []rune("~")},
				{Type: parser.SELECTOR, Literal: []rune("span")},
				{Type: parser.LBRACE, Literal: []rune("{")},
				{Type: parser.IDENT, Literal: []rune("color")},
				{Type: parser.COLON, Literal: []rune(":")},
				{Type: parser.IDENT, Literal: []rune("red")},
				{Type: parser.SEMICOLON, Literal: []rune(";")},
				{Type: parser.RBRACE, Literal: []rune("}")},
			},
		},
		{
			name:  "Pseudo-elements",
			input: "p::first-line { text-transform: uppercase; }",
			expected: []parser.Token{
				{Type: parser.SELECTOR, Literal: []rune("p")},
				{Type: parser.PSEUDO_ELEMENT, Literal: []rune("::first-line")},
				{Type: parser.LBRACE, Literal: []rune("{")},
				{Type: parser.IDENT, Literal: []rune("text-transform")},
				{Type: parser.COLON, Literal: []rune(":")},
				{Type: parser.IDENT, Literal: []rune("uppercase")},
				{Type: parser.SEMICOLON, Literal: []rune(";")},
				{Type: parser.RBRACE, Literal: []rune("}")},
			},
		},
		{
			name:  "Multiple Selectors",
			input: "h1, h2, h3 { font-family: sans-serif; }",
			expected: []parser.Token{
				{Type: parser.SELECTOR, Literal: []rune("h1")},
				{Type: parser.COMMA, Literal: []rune(",")},
				{Type: parser.SELECTOR, Literal: []rune("h2")},
				{Type: parser.COMMA, Literal: []rune(",")},
				{Type: parser.SELECTOR, Literal: []rune("h3")},
				{Type: parser.LBRACE, Literal: []rune("{")},
				{Type: parser.IDENT, Literal: []rune("font-family")},
				{Type: parser.COLON, Literal: []rune(":")},
				{Type: parser.IDENT, Literal: []rune("sans-serif")},
				{Type: parser.SEMICOLON, Literal: []rune(";")},
				{Type: parser.RBRACE, Literal: []rune("}")},
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
				{Type: parser.PSEUDO_CLASS, Literal: []rune(":root")},
				{Type: parser.LBRACE, Literal: []rune("{")},
				{Type: parser.IDENT, Literal: []rune("--main-color")},
				{Type: parser.COLON, Literal: []rune(":")},
				{Type: parser.IDENT, Literal: []rune("blue")},
				{Type: parser.SEMICOLON, Literal: []rune(";")},
				{Type: parser.RBRACE, Literal: []rune("}")},

				{Type: parser.SELECTOR, Literal: []rune("li")},
				{Type: parser.PSEUDO_CLASS, Literal: []rune(":nth-child")},
				{Type: parser.LPAREN, Literal: []rune("(")},
				{Type: parser.NUMBER, Literal: []rune("2")},
				{Type: parser.IDENT, Literal: []rune("n")},
				{Type: parser.PLUS, Literal: []rune("+")},
				{Type: parser.NUMBER, Literal: []rune("1")},
				{Type: parser.RPAREN, Literal: []rune(")")},
				{Type: parser.LBRACE, Literal: []rune("{")},
				{Type: parser.IDENT, Literal: []rune("background")},
				{Type: parser.COLON, Literal: []rune(":")},
				{Type: parser.IDENT, Literal: []rune("lightgray")},
				{Type: parser.SEMICOLON, Literal: []rune(";")},
				{Type: parser.RBRACE, Literal: []rune("}")},

				{Type: parser.SELECTOR, Literal: []rune("tr")},
				{Type: parser.PSEUDO_CLASS, Literal: []rune(":nth-child")},
				{Type: parser.LPAREN, Literal: []rune("(")},
				{Type: parser.IDENT, Literal: []rune("odd")},
				{Type: parser.RPAREN, Literal: []rune(")")},
				{Type: parser.LBRACE, Literal: []rune("{")},
				{Type: parser.IDENT, Literal: []rune("background-color")},
				{Type: parser.COLON, Literal: []rune(":")},
				{Type: parser.COLOR, Literal: []rune("#f2f2f2")},
				{Type: parser.SEMICOLON, Literal: []rune(";")},
				{Type: parser.RBRACE, Literal: []rune("}")},

				{Type: parser.SELECTOR, Literal: []rune("div")},
				{Type: parser.PSEUDO_CLASS, Literal: []rune(":nth-child")},
				{Type: parser.LPAREN, Literal: []rune("(")},
				{Type: parser.IDENT, Literal: []rune("-n")},
				{Type: parser.PLUS, Literal: []rune("+")},
				{Type: parser.NUMBER, Literal: []rune("3")},
				{Type: parser.RPAREN, Literal: []rune(")")},
				{Type: parser.LBRACE, Literal: []rune("{")},
				{Type: parser.IDENT, Literal: []rune("font-weight")},
				{Type: parser.COLON, Literal: []rune(":")},
				{Type: parser.IDENT, Literal: []rune("bold")},
				{Type: parser.SEMICOLON, Literal: []rune(";")},
				{Type: parser.RBRACE, Literal: []rune("}")},
			},
		},
		{
			name: "Pseudo-Elements",
			input: `
        p::first-line { font-weight: bold; }
        div::before { content: ""; display: block; }
    `,
			expected: []parser.Token{
				{Type: parser.SELECTOR, Literal: []rune("p")},
				{Type: parser.PSEUDO_ELEMENT, Literal: []rune("::first-line")},
				{Type: parser.LBRACE, Literal: []rune("{")},
				{Type: parser.IDENT, Literal: []rune("font-weight")},
				{Type: parser.COLON, Literal: []rune(":")},
				{Type: parser.IDENT, Literal: []rune("bold")},
				{Type: parser.SEMICOLON, Literal: []rune(";")},
				{Type: parser.RBRACE, Literal: []rune("}")},

				{Type: parser.SELECTOR, Literal: []rune("div")},
				{Type: parser.PSEUDO_ELEMENT, Literal: []rune("::before")},
				{Type: parser.LBRACE, Literal: []rune("{")},
				{Type: parser.IDENT, Literal: []rune("content")},
				{Type: parser.COLON, Literal: []rune(":")},
				{Type: parser.STRING, Literal: []rune("\"\"")},
				{Type: parser.SEMICOLON, Literal: []rune(";")},
				{Type: parser.IDENT, Literal: []rune("display")},
				{Type: parser.COLON, Literal: []rune(":")},
				{Type: parser.IDENT, Literal: []rune("block")},
				{Type: parser.SEMICOLON, Literal: []rune(";")},
				{Type: parser.RBRACE, Literal: []rune("}")},
			},
		},
		{
			name:  "Media Query",
			input: "@media screen and (max-width: 600px) { body { font-size: 14px; } }",
			expected: []parser.Token{
				{Type: parser.AT_RULE, Literal: []rune("@media")},
				{Type: parser.IDENT, Literal: []rune("screen")},
				{Type: parser.IDENT, Literal: []rune("and")},
				{Type: parser.LPAREN, Literal: []rune("(")},
				{Type: parser.IDENT, Literal: []rune("max-width")},
				{Type: parser.COLON, Literal: []rune(":")},
				{Type: parser.NUMBER, Literal: []rune("600")},
				{Type: parser.UNIT, Literal: []rune("px")},
				{Type: parser.RPAREN, Literal: []rune(")")},
				{Type: parser.LBRACE, Literal: []rune("{")},
				{Type: parser.IDENT, Literal: []rune("body")},
				{Type: parser.LBRACE, Literal: []rune("{")},
				{Type: parser.IDENT, Literal: []rune("font-size")},
				{Type: parser.COLON, Literal: []rune(":")},
				{Type: parser.NUMBER, Literal: []rune("14")},
				{Type: parser.UNIT, Literal: []rune("px")},
				{Type: parser.SEMICOLON, Literal: []rune(";")},
				{Type: parser.RBRACE, Literal: []rune("}")},
				{Type: parser.RBRACE, Literal: []rune("}")},
			},
		},
		{
			name:  "Complex Selectors",
			input: "#main > .article p:first-child { color: #ff0000; }",
			expected: []parser.Token{
				{Type: parser.SELECTOR, Literal: []rune("#main")},
				{Type: parser.GREATER, Literal: []rune(">")},
				{Type: parser.SELECTOR, Literal: []rune(".article")},
				{Type: parser.SELECTOR, Literal: []rune("p")},
				{Type: parser.PSEUDO_CLASS, Literal: []rune(":first-child")},
				{Type: parser.LBRACE, Literal: []rune("{")},
				{Type: parser.IDENT, Literal: []rune("color")},
				{Type: parser.COLON, Literal: []rune(":")},
				{Type: parser.COLOR, Literal: []rune("#ff0000")},
				{Type: parser.SEMICOLON, Literal: []rune(";")},
				{Type: parser.RBRACE, Literal: []rune("}")},
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
				{Type: parser.SELECTOR, Literal: []rune("main")},
				{Type: parser.LBRACE, Literal: []rune("{")},
				{Type: parser.IDENT, Literal: []rune("color")},
				{Type: parser.COLON, Literal: []rune(":")},
				{Type: parser.COLOR, Literal: []rune("#ff0000")},
				{Type: parser.IMPORTANT, Literal: []rune("!important")},
				{Type: parser.SEMICOLON, Literal: []rune(";")},
				{Type: parser.RBRACE, Literal: []rune("}")},

				{Type: parser.SELECTOR, Literal: []rune("p")},
				{Type: parser.LBRACE, Literal: []rune("{")},
				{Type: parser.IDENT, Literal: []rune("margin")},
				{Type: parser.COLON, Literal: []rune(":")},
				{Type: parser.NUMBER, Literal: []rune("10")},
				{Type: parser.UNIT, Literal: []rune("px")},
				{Type: parser.IMPORTANT, Literal: []rune("! important")},
				{Type: parser.SEMICOLON, Literal: []rune(";")},
				{Type: parser.RBRACE, Literal: []rune("}")},

				{Type: parser.SELECTOR, Literal: []rune("div")},
				{Type: parser.LBRACE, Literal: []rune("{")},
				{Type: parser.IDENT, Literal: []rune("padding")},
				{Type: parser.COLON, Literal: []rune(":")},
				{Type: parser.NUMBER, Literal: []rune("5")},
				{Type: parser.UNIT, Literal: []rune("px")},
				{Type: parser.IMPORTANT, Literal: []rune("!   important")},
				{Type: parser.SEMICOLON, Literal: []rune(";")},
				{Type: parser.RBRACE, Literal: []rune("}")},

				{Type: parser.SELECTOR, Literal: []rune("span")},
				{Type: parser.LBRACE, Literal: []rune("{")},
				{Type: parser.IDENT, Literal: []rune("font-weight")},
				{Type: parser.COLON, Literal: []rune(":")},
				{Type: parser.IDENT, Literal: []rune("bold")},
				{Type: parser.IMPORTANT, Literal: []rune("!\nimportant")},
				{Type: parser.SEMICOLON, Literal: []rune(";")},
				{Type: parser.RBRACE, Literal: []rune("}")},

				{Type: parser.SELECTOR, Literal: []rune(".custom")},
				{Type: parser.LBRACE, Literal: []rune("{")},
				{Type: parser.IDENT, Literal: []rune("border")},
				{Type: parser.COLON, Literal: []rune(":")},
				{Type: parser.NUMBER, Literal: []rune("1")},
				{Type: parser.UNIT, Literal: []rune("px")},
				{Type: parser.IDENT, Literal: []rune("solid")},
				{Type: parser.IDENT, Literal: []rune("black")},
				{Type: parser.IMPORTANT, Literal: []rune("!important")},
				{Type: parser.RBRACE, Literal: []rune("}")},
			},
		},
		{
			name:  "Attribute Selector",
			input: "a[href^=\"https://\"] { color: green; }",
			expected: []parser.Token{
				{Type: parser.SELECTOR, Literal: []rune("a[href^=\"https://\"]")},
				{Type: parser.LBRACE, Literal: []rune("{")},
				{Type: parser.IDENT, Literal: []rune("color")},
				{Type: parser.COLON, Literal: []rune(":")},
				{Type: parser.IDENT, Literal: []rune("green")},
				{Type: parser.SEMICOLON, Literal: []rune(";")},
				{Type: parser.RBRACE, Literal: []rune("}")},
			},
		},
		{
			name:  "Keyframes",
			input: "@keyframes fadeIn { 0% { opacity: 0; } 100% { opacity: 1; } }",
			expected: []parser.Token{
				{Type: parser.AT_RULE, Literal: []rune("@keyframes")},
				{Type: parser.IDENT, Literal: []rune("fadeIn")},
				{Type: parser.LBRACE, Literal: []rune("{")},
				{Type: parser.NUMBER, Literal: []rune("0")},
				{Type: parser.UNIT, Literal: []rune("%")},
				{Type: parser.LBRACE, Literal: []rune("{")},
				{Type: parser.IDENT, Literal: []rune("opacity")},
				{Type: parser.COLON, Literal: []rune(":")},
				{Type: parser.NUMBER, Literal: []rune("0")},
				{Type: parser.SEMICOLON, Literal: []rune(";")},
				{Type: parser.RBRACE, Literal: []rune("}")},
				{Type: parser.NUMBER, Literal: []rune("100")},
				{Type: parser.UNIT, Literal: []rune("%")},
				{Type: parser.LBRACE, Literal: []rune("{")},
				{Type: parser.IDENT, Literal: []rune("opacity")},
				{Type: parser.COLON, Literal: []rune(":")},
				{Type: parser.NUMBER, Literal: []rune("1")},
				{Type: parser.SEMICOLON, Literal: []rune(";")},
				{Type: parser.RBRACE, Literal: []rune("}")},
				{Type: parser.RBRACE, Literal: []rune("}")},
			},
		},
		{
			name:  "CSS Variables",
			input: ":root { --main-color: blue; } body { color: var(--main-color); }",
			expected: []parser.Token{
				{Type: parser.PSEUDO_CLASS, Literal: []rune(":root")},
				{Type: parser.LBRACE, Literal: []rune("{")},
				{Type: parser.IDENT, Literal: []rune("--main-color")},
				{Type: parser.COLON, Literal: []rune(":")},
				{Type: parser.IDENT, Literal: []rune("blue")},
				{Type: parser.SEMICOLON, Literal: []rune(";")},
				{Type: parser.RBRACE, Literal: []rune("}")},
				{Type: parser.SELECTOR, Literal: []rune("body")},
				{Type: parser.LBRACE, Literal: []rune("{")},
				{Type: parser.IDENT, Literal: []rune("color")},
				{Type: parser.COLON, Literal: []rune(":")},
				{Type: parser.IDENT, Literal: []rune("var")},
				{Type: parser.LPAREN, Literal: []rune("(")},
				{Type: parser.IDENT, Literal: []rune("--main-color")},
				{Type: parser.RPAREN, Literal: []rune(")")},
				{Type: parser.SEMICOLON, Literal: []rune(";")},
				{Type: parser.RBRACE, Literal: []rune("}")},
			},
		},
		{
			name:  "Calc Function",
			input: "div { width: calc(100% - 20px); height: 100vh; }",
			expected: []parser.Token{
				{Type: parser.SELECTOR, Literal: []rune("div")},
				{Type: parser.LBRACE, Literal: []rune("{")},
				{Type: parser.IDENT, Literal: []rune("width")},
				{Type: parser.COLON, Literal: []rune(":")},
				{Type: parser.IDENT, Literal: []rune("calc")},
				{Type: parser.LPAREN, Literal: []rune("(")},
				{Type: parser.NUMBER, Literal: []rune("100")},
				{Type: parser.UNIT, Literal: []rune("%")},
				{Type: parser.MINUS, Literal: []rune("-")},
				{Type: parser.NUMBER, Literal: []rune("20")},
				{Type: parser.UNIT, Literal: []rune("px")},
				{Type: parser.RPAREN, Literal: []rune(")")},
				{Type: parser.SEMICOLON, Literal: []rune(";")},
				{Type: parser.IDENT, Literal: []rune("height")},
				{Type: parser.COLON, Literal: []rune(":")},
				{Type: parser.NUMBER, Literal: []rune("100")},
				{Type: parser.UNIT, Literal: []rune("vh")},
				{Type: parser.SEMICOLON, Literal: []rune(";")},
				{Type: parser.RBRACE, Literal: []rune("}")},
			},
		},
		{
			name:  "Escaped characters in identifiers",
			input: ".foo\\.bar { color: red; }",
			expected: []parser.Token{
				{Type: parser.SELECTOR, Literal: []rune(".foo\\.bar")},
				{Type: parser.LBRACE, Literal: []rune("{")},
				{Type: parser.IDENT, Literal: []rune("color")},
				{Type: parser.COLON, Literal: []rune(":")},
				{Type: parser.IDENT, Literal: []rune("red")},
				{Type: parser.SEMICOLON, Literal: []rune(";")},
				{Type: parser.RBRACE, Literal: []rune("}")},
			},
		},
		{
			name:  "Escaped characters in attribute selectors",
			input: "a[href=\"foo\\\"bar\"] { color: blue; }",
			expected: []parser.Token{
				{Type: parser.SELECTOR, Literal: []rune("a[href=\"foo\\\"bar\"]")},
				{Type: parser.LBRACE, Literal: []rune("{")},
				{Type: parser.IDENT, Literal: []rune("color")},
				{Type: parser.COLON, Literal: []rune(":")},
				{Type: parser.IDENT, Literal: []rune("blue")},
				{Type: parser.SEMICOLON, Literal: []rune(";")},
				{Type: parser.RBRACE, Literal: []rune("}")},
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
				{Type: parser.SELECTOR, Literal: []rune("#☃")},
				{Type: parser.LBRACE, Literal: []rune("{")},
				{Type: parser.IDENT, Literal: []rune("color")},
				{Type: parser.COLON, Literal: []rune(":")},
				{Type: parser.IDENT, Literal: []rune("skyblue")},
				{Type: parser.SEMICOLON, Literal: []rune(";")},
				{Type: parser.RBRACE, Literal: []rune("}")},
				{Type: parser.SELECTOR, Literal: []rune(".günther")},
				{Type: parser.COMMA, Literal: []rune(",")},
				{Type: parser.SELECTOR, Literal: []rune("#π_value")},
				{Type: parser.COMMA, Literal: []rune(",")},
				{Type: parser.SELECTOR, Literal: []rune(".こんにちは")},
				{Type: parser.LBRACE, Literal: []rune("{")},
				{Type: parser.IDENT, Literal: []rune("color")},
				{Type: parser.COLON, Literal: []rune(":")},
				{Type: parser.IDENT, Literal: []rune("blue")},
				{Type: parser.SEMICOLON, Literal: []rune(";")},
				{Type: parser.RBRACE, Literal: []rune("}")},
				{Type: parser.DOT, Literal: []rune(".")},
				{Type: parser.IDENT, Literal: []rune("--custom-prop")},
				{Type: parser.COMMA, Literal: []rune(",")},
				{Type: parser.DOT, Literal: []rune(".")},
				{Type: parser.IDENT, Literal: []rune("-moz-custom")},
				{Type: parser.LBRACE, Literal: []rune("{")},
				{Type: parser.IDENT, Literal: []rune("value")},
				{Type: parser.COLON, Literal: []rune(":")},
				{Type: parser.NUMBER, Literal: []rune("123")},
				{Type: parser.SEMICOLON, Literal: []rune(";")},
				{Type: parser.RBRACE, Literal: []rune("}")},
				{Type: parser.SELECTOR, Literal: []rune("#\\26 ABC")},
				{Type: parser.LBRACE, Literal: []rune("{")},
				{Type: parser.COMMENT, Literal: []rune("/* Escaped ASCII */")},
				{Type: parser.IDENT, Literal: []rune("color")},
				{Type: parser.COLON, Literal: []rune(":")},
				{Type: parser.IDENT, Literal: []rune("red")},
				{Type: parser.SEMICOLON, Literal: []rune(";")},
				{Type: parser.RBRACE, Literal: []rune("}")},
			},
		},
		{
			name:  "CSS custom properties (variables)",
			input: ":root { --custom-color: #ff00ff; } .foo { color: var(--custom-color); }",
			expected: []parser.Token{
				{Type: parser.PSEUDO_CLASS, Literal: []rune(":root")},
				{Type: parser.LBRACE, Literal: []rune("{")},
				{Type: parser.IDENT, Literal: []rune("--custom-color")},
				{Type: parser.COLON, Literal: []rune(":")},
				{Type: parser.COLOR, Literal: []rune("#ff00ff")},
				{Type: parser.SEMICOLON, Literal: []rune(";")},
				{Type: parser.RBRACE, Literal: []rune("}")},
				{Type: parser.SELECTOR, Literal: []rune(".foo")},
				{Type: parser.LBRACE, Literal: []rune("{")},
				{Type: parser.IDENT, Literal: []rune("color")},
				{Type: parser.COLON, Literal: []rune(":")},
				{Type: parser.IDENT, Literal: []rune("var")},
				{Type: parser.LPAREN, Literal: []rune("(")},
				{Type: parser.IDENT, Literal: []rune("--custom-color")},
				{Type: parser.RPAREN, Literal: []rune(")")},
				{Type: parser.SEMICOLON, Literal: []rune(";")},
				{Type: parser.RBRACE, Literal: []rune("}")},
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
				{Type: parser.SELECTOR, Literal: []rune(".gradient")},
				{Type: parser.LBRACE, Literal: []rune("{")},
				{Type: parser.IDENT, Literal: []rune("background-image")},
				{Type: parser.COLON, Literal: []rune(":")},
				{Type: parser.IDENT, Literal: []rune("linear-gradient")},
				{Type: parser.LPAREN, Literal: []rune("(")},
				{Type: parser.NUMBER, Literal: []rune("45")},
				{Type: parser.UNIT, Literal: []rune("deg")},
				{Type: parser.COMMA, Literal: []rune(",")},
				{Type: parser.COLOR, Literal: []rune("#ff0000")},
				{Type: parser.COMMA, Literal: []rune(",")},
				{Type: parser.COLOR, Literal: []rune("#00ff00")},
				{Type: parser.RPAREN, Literal: []rune(")")},
				{Type: parser.COMMA, Literal: []rune(",")},
				{Type: parser.IDENT, Literal: []rune("radial-gradient")},
				{Type: parser.LPAREN, Literal: []rune("(")},
				{Type: parser.IDENT, Literal: []rune("circle")},
				{Type: parser.COMMA, Literal: []rune(",")},
				{Type: parser.COLOR, Literal: []rune("#0000ff")},
				{Type: parser.COMMA, Literal: []rune(",")},
				{Type: parser.COLOR, Literal: []rune("#ffff00")},
				{Type: parser.RPAREN, Literal: []rune(")")},
				{Type: parser.SEMICOLON, Literal: []rune(";")},
				{Type: parser.IDENT, Literal: []rune("font")},
				{Type: parser.COLON, Literal: []rune(":")},
				{Type: parser.IDENT, Literal: []rune("bold")},
				{Type: parser.NUMBER, Literal: []rune("12")},
				{Type: parser.UNIT, Literal: []rune("px")},
				{Type: parser.DIVIDE, Literal: []rune("/")},
				{Type: parser.NUMBER, Literal: []rune("14")},
				{Type: parser.UNIT, Literal: []rune("px")},
				{Type: parser.STRING, Literal: []rune("\"Helvetica\"")},
				{Type: parser.COMMA, Literal: []rune(",")},
				{Type: parser.IDENT, Literal: []rune("sans-serif")},
				{Type: parser.SEMICOLON, Literal: []rune(";")},
				{Type: parser.RBRACE, Literal: []rune("}")},
			},
		},
		{
			name: "CSS Hacks and Legacy Browser Support",
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
			ignore: true,
			expected: []parser.Token{
				{Type: parser.COMMENT, Literal: []rune(" Pseudo-element hack ")},
				{Type: parser.IDENT, Literal: []rune("_")},
				{Type: parser.COLON, Literal: []rune(":")},
				{Type: parser.COLON, Literal: []rune(":")},
				{Type: parser.IDENT, Literal: []rune("-webkit-full-page-media")},
				{Type: parser.COMMA, Literal: []rune(",")},
				{Type: parser.IDENT, Literal: []rune("_")},
				{Type: parser.COLON, Literal: []rune(":")},
				{Type: parser.IDENT, Literal: []rune("future")},
				{Type: parser.COMMA, Literal: []rune(",")},
				{Type: parser.COLON, Literal: []rune(":")},
				{Type: parser.IDENT, Literal: []rune("root")},
				{Type: parser.IDENT, Literal: []rune(".foo")},
				{Type: parser.LBRACE, Literal: []rune("{")},
				{Type: parser.IDENT, Literal: []rune("color")},
				{Type: parser.COLON, Literal: []rune(":")},
				{Type: parser.IDENT, Literal: []rune("red")},
				{Type: parser.SEMICOLON, Literal: []rune(";")},
				{Type: parser.RBRACE, Literal: []rune("}")},

				{Type: parser.COMMENT, Literal: []rune(" Pseudo-class hack ")},
				{Type: parser.IDENT, Literal: []rune("_")},
				{Type: parser.COLON, Literal: []rune(":")},
				{Type: parser.IDENT, Literal: []rune("-webkit-full-screen")},
				{Type: parser.COMMA, Literal: []rune(",")},
				{Type: parser.COLON, Literal: []rune(":")},
				{Type: parser.IDENT, Literal: []rune("root")},
				{Type: parser.IDENT, Literal: []rune(".bar")},
				{Type: parser.LBRACE, Literal: []rune("{")},
				{Type: parser.IDENT, Literal: []rune("display")},
				{Type: parser.COLON, Literal: []rune(":")},
				{Type: parser.IDENT, Literal: []rune("block")},
				{Type: parser.SEMICOLON, Literal: []rune(";")},
				{Type: parser.RBRACE, Literal: []rune("}")},

				{Type: parser.COMMENT, Literal: []rune(" Media query hack ")},
				{Type: parser.AT_RULE, Literal: []rune("@media")},
				{Type: parser.IDENT, Literal: []rune("screen")},
				{Type: parser.IDENT, Literal: []rune("and")},
				{Type: parser.LPAREN, Literal: []rune("(")},
				{Type: parser.IDENT, Literal: []rune("min-width")},
				{Type: parser.COLON, Literal: []rune(":")},
				{Type: parser.NUMBER, Literal: []rune("0")},
				{Type: parser.IDENT, Literal: []rune("\\0")},
				{Type: parser.RPAREN, Literal: []rune(")")},
				{Type: parser.LBRACE, Literal: []rune("{")},
				{Type: parser.IDENT, Literal: []rune(".baz")},
				{Type: parser.LBRACE, Literal: []rune("{")},
				{Type: parser.IDENT, Literal: []rune("zoom")},
				{Type: parser.COLON, Literal: []rune(":")},
				{Type: parser.NUMBER, Literal: []rune("1")},
				{Type: parser.SEMICOLON, Literal: []rune(";")},
				{Type: parser.RBRACE, Literal: []rune("}")},
				{Type: parser.RBRACE, Literal: []rune("}")},

				{Type: parser.COMMENT, Literal: []rune(" Property value hacks ")},
				{Type: parser.IDENT, Literal: []rune(".hack1")},
				{Type: parser.LBRACE, Literal: []rune("{")},
				{Type: parser.IDENT, Literal: []rune("property")},
				{Type: parser.COLON, Literal: []rune(":")},
				{Type: parser.IDENT, Literal: []rune("value")},
				{Type: parser.IDENT, Literal: []rune("\\9")},
				{Type: parser.SEMICOLON, Literal: []rune(";")},
				{Type: parser.RBRACE, Literal: []rune("}")},

				{Type: parser.IDENT, Literal: []rune(".hack2")},
				{Type: parser.LBRACE, Literal: []rune("{")},
				{Type: parser.IDENT, Literal: []rune("property")},
				{Type: parser.COLON, Literal: []rune(":")},
				{Type: parser.IDENT, Literal: []rune("value")},
				{Type: parser.IDENT, Literal: []rune("\\9")},
				{Type: parser.SEMICOLON, Literal: []rune(";")},
				{Type: parser.RBRACE, Literal: []rune("}")},

				{Type: parser.IDENT, Literal: []rune(".hack3")},
				{Type: parser.LBRACE, Literal: []rune("{")},
				{Type: parser.IDENT, Literal: []rune("property")},
				{Type: parser.COLON, Literal: []rune(":")},
				{Type: parser.IDENT, Literal: []rune("value")},
				{Type: parser.IDENT, Literal: []rune("\\0")},
				{Type: parser.SEMICOLON, Literal: []rune(";")},
				{Type: parser.RBRACE, Literal: []rune("}")},

				{Type: parser.IDENT, Literal: []rune(".hack4")},
				{Type: parser.LBRACE, Literal: []rune("{")},
				{Type: parser.IDENT, Literal: []rune("property")},
				{Type: parser.COLON, Literal: []rune(":")},
				{Type: parser.IDENT, Literal: []rune("value")},
				{Type: parser.IDENT, Literal: []rune("\\0")},
				{Type: parser.SEMICOLON, Literal: []rune(";")},
				{Type: parser.RBRACE, Literal: []rune("}")},

				{Type: parser.COMMENT, Literal: []rune(" Combination of hacks ")},
				{Type: parser.IDENT, Literal: []rune("_")},
				{Type: parser.COLON, Literal: []rune(":")},
				{Type: parser.IDENT, Literal: []rune("-ms-lang")},
				{Type: parser.LPAREN, Literal: []rune("(")},
				{Type: parser.IDENT, Literal: []rune("x")},
				{Type: parser.RPAREN, Literal: []rune(")")},
				{Type: parser.COMMA, Literal: []rune(",")},
				{Type: parser.IDENT, Literal: []rune("_")},
				{Type: parser.COLON, Literal: []rune(":")},
				{Type: parser.IDENT, Literal: []rune("-webkit-full-screen")},
				{Type: parser.COMMA, Literal: []rune(",")},
				{Type: parser.IDENT, Literal: []rune(".multi-hack")},
				{Type: parser.LBRACE, Literal: []rune("{")},
				{Type: parser.IDENT, Literal: []rune("*display")},
				{Type: parser.COLON, Literal: []rune(":")},
				{Type: parser.IDENT, Literal: []rune("inline")},
				{Type: parser.SEMICOLON, Literal: []rune(";")},
				{Type: parser.IDENT, Literal: []rune("_height")},
				{Type: parser.COLON, Literal: []rune(":")},
				{Type: parser.NUMBER, Literal: []rune("1")},
				{Type: parser.UNIT, Literal: []rune("%")},
				{Type: parser.SEMICOLON, Literal: []rune(";")},
				{Type: parser.RBRACE, Literal: []rune("}")},
			},
		},
	}

	for _, tt := range tests {
		if tt.ignore {
			continue
		}

		t.Run(tt.name, func(t *testing.T) {
			l := parser.NewLexer(strings.NewReader(tt.input))
			for i, expected := range tt.expected {

				tok := l.NextToken()
				if tok.Type != expected.Type {
					t.Errorf("tests[%d] - tokentype wrong. expected=%q, got=%q %s",
						i, expected.Type, tok.Type, string(tok.Literal))
				}
				if !runesEqual(tok.Literal, expected.Literal) {
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
				{Type: parser.SELECTOR, Literal: []rune(".invalid")},
				{Type: parser.LBRACE, Literal: []rune("{")},
				{Type: parser.IDENT, Literal: []rune("color")},
				{Type: parser.COLON, Literal: []rune(":")},
				{Type: parser.SELECTOR, Literal: []rune("#1234ZZ")},
				{Type: parser.SEMICOLON, Literal: []rune(";")},
				{Type: parser.RBRACE, Literal: []rune("}")},
			},
		},
		{
			name:  "Invalid unit combination",
			input: ".invalid-unit { width: 50+px; }",
			expected: []parser.Token{
				{Type: parser.SELECTOR, Literal: []rune(".invalid-unit")},
				{Type: parser.LBRACE, Literal: []rune("{")},
				{Type: parser.IDENT, Literal: []rune("width")},
				{Type: parser.COLON, Literal: []rune(":")},
				{Type: parser.NUMBER, Literal: []rune("50")},
				{Type: parser.PLUS, Literal: []rune("+")},
				{Type: parser.UNIT, Literal: []rune("px")},
				{Type: parser.SEMICOLON, Literal: []rune(";")},
				{Type: parser.RBRACE, Literal: []rune("}")},
			},
		},
		{
			name:  "Invalid percentage",
			input: ".invalid-percentage { height: 100vh%; }",
			expected: []parser.Token{
				{Type: parser.SELECTOR, Literal: []rune(".invalid-percentage")},
				{Type: parser.LBRACE, Literal: []rune("{")},
				{Type: parser.IDENT, Literal: []rune("height")},
				{Type: parser.COLON, Literal: []rune(":")},
				{Type: parser.NUMBER, Literal: []rune("100")},
				{Type: parser.UNIT, Literal: []rune("vh")},
				{Type: parser.ILLEGAL, Literal: []rune("%")},
				{Type: parser.SEMICOLON, Literal: []rune(";")},
				{Type: parser.RBRACE, Literal: []rune("}")},
			},
		},
		// You can add more test cases here
		{
			name:  "Invalid property",
			input: ".invalid-property { colo r: red; }",
			expected: []parser.Token{
				{Type: parser.SELECTOR, Literal: []rune(".invalid-property")},
				{Type: parser.LBRACE, Literal: []rune("{")},
				{Type: parser.IDENT, Literal: []rune("colo")},
				{Type: parser.IDENT, Literal: []rune("r")},
				{Type: parser.COLON, Literal: []rune(":")},
				{Type: parser.IDENT, Literal: []rune("red")},
				{Type: parser.SEMICOLON, Literal: []rune(";")},
				{Type: parser.RBRACE, Literal: []rune("}")},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := parser.NewLexer(strings.NewReader(tt.input))
			for i, expected := range tt.expected {
				got := l.NextToken()
				if got.Type != expected.Type {
					t.Errorf("Token %d: expected type %v, got %v", i, expected.Type, got.Type)
				}
				if !runesEqual(got.Literal, expected.Literal) {
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

			l := parser.NewLexer(input)
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
				l := parser.NewLexer(reader)
				for tok := l.NextToken(); tok.Type != parser.EOF; tok = l.NextToken() {
					// Do nothing, just lex
				}
			}
		})
	}
}

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
