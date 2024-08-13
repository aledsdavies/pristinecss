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

func TestBasicTokens(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []tokens.Token
	}{
		{
			name:  "Comments",
			input: "/* This is a comment */",
			expected: []tokens.Token{
				{Type: tokens.COMMENT, Literal: []byte("/* This is a comment */")},
			},
		},
		{
			name:  "Simple element selector",
			input: "div { color: blue; }",
			expected: []tokens.Token{
				{Type: tokens.IDENT, Literal: []byte("div")},
				{Type: tokens.LBRACE, Literal: []byte("{")},
				{Type: tokens.IDENT, Literal: []byte("color")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.IDENT, Literal: []byte("blue")},
				{Type: tokens.SEMICOLON, Literal: []byte(";")},
				{Type: tokens.RBRACE, Literal: []byte("}")},
			},
		},
	}

	runTests(t, tests)
}

func TestSelectors(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []tokens.Token
	}{
		{
			name:  "Class selector",
			input: ".highlight { background-color: yellow; }",
			expected: []tokens.Token{
				{Type: tokens.DOT, Literal: []byte(".")},
				{Type: tokens.IDENT, Literal: []byte("highlight")},
				{Type: tokens.LBRACE, Literal: []byte("{")},
				{Type: tokens.IDENT, Literal: []byte("background-color")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.IDENT, Literal: []byte("yellow")},
				{Type: tokens.SEMICOLON, Literal: []byte(";")},
				{Type: tokens.RBRACE, Literal: []byte("}")},
			},
		},
		{
			name:  "ID selector",
			input: "#main { font-size: 16px; }",
			expected: []tokens.Token{
				{Type: tokens.HASH, Literal: []byte("#")},
				{Type: tokens.IDENT, Literal: []byte("main")},
				{Type: tokens.LBRACE, Literal: []byte("{")},
				{Type: tokens.IDENT, Literal: []byte("font-size")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.NUMBER, Literal: []byte("16")},
				{Type: tokens.IDENT, Literal: []byte("px")},
				{Type: tokens.SEMICOLON, Literal: []byte(";")},
				{Type: tokens.RBRACE, Literal: []byte("}")},
			},
		},
		{
			name:  "Attribute selector",
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
	}

	runTests(t, tests)
}

func TestPseudoSelectors(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []tokens.Token
	}{
		{
			name:  "Pseudo-class",
			input: "a:hover { color: red; }",
			expected: []tokens.Token{
				{Type: tokens.IDENT, Literal: []byte("a")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.IDENT, Literal: []byte("hover")},
				{Type: tokens.LBRACE, Literal: []byte("{")},
				{Type: tokens.IDENT, Literal: []byte("color")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.IDENT, Literal: []byte("red")},
				{Type: tokens.SEMICOLON, Literal: []byte(";")},
				{Type: tokens.RBRACE, Literal: []byte("}")},
			},
		},
		{
			name:  "Pseudo-element",
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
	}

	runTests(t, tests)
}

func TestComplexSelectors(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []tokens.Token
	}{
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
	}

	runTests(t, tests)
}

func TestAtRules(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []tokens.Token
	}{
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
	}

	runTests(t, tests)
}

func TestURIVariations(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []tokens.Token
	}{
		{
			name:  "Simple URL",
			input: ".foo {background-image: url('image.jpg');}",
			expected: []tokens.Token{
                {Type: tokens.DOT, Literal: []byte(".")},
                {Type: tokens.IDENT, Literal: []byte("foo")},
                {Type: tokens.LBRACE, Literal: []byte("{")},
				{Type: tokens.IDENT, Literal: []byte("background-image")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.URI, Literal: []byte("url('image.jpg')")},
				{Type: tokens.SEMICOLON, Literal: []byte(";")},
                {Type: tokens.RBRACE, Literal: []byte("}")},
			},
		},
		{
			name:  "URL with protocol",
			input: "background-image: url('https://example.com/image.png');",
			expected: []tokens.Token{
				{Type: tokens.IDENT, Literal: []byte("background-image")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.URI, Literal: []byte("url('https://example.com/image.png')")},
				{Type: tokens.SEMICOLON, Literal: []byte(";")},
			},
		},
		{
			name:  "URL without quotes",
			input: "background-image: url(image.gif);",
			expected: []tokens.Token{
				{Type: tokens.IDENT, Literal: []byte("background-image")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.URI, Literal: []byte("url(image.gif)")},
				{Type: tokens.SEMICOLON, Literal: []byte(";")},
			},
		},
		{
			name:  "Data URI",
			input: `background-image: url("data:image/svg+xml;utf8,<svg xmlns='http://www.w3.org/2000/svg'></svg>");`,
			expected: []tokens.Token{
				{Type: tokens.IDENT, Literal: []byte("background-image")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.URI, Literal: []byte(`url("data:image/svg+xml;utf8,<svg xmlns='http://www.w3.org/2000/svg'></svg>")`)},
				{Type: tokens.SEMICOLON, Literal: []byte(";")},
			},
		},
		{
			name:  "URL with parentheses",
			input: `background-image: url("https://example.com/image(1).jpg");`,
			expected: []tokens.Token{
				{Type: tokens.IDENT, Literal: []byte("background-image")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.URI, Literal: []byte(`url("https://example.com/image(1).jpg")`)},
				{Type: tokens.SEMICOLON, Literal: []byte(";")},
			},
		},
		{
			name:  "Multiple URLs",
			input: `background: url("image1.jpg"), url('image2.png'), url(image3.gif);`,
			expected: []tokens.Token{
				{Type: tokens.IDENT, Literal: []byte("background")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.URI, Literal: []byte(`url("image1.jpg")`)},
				{Type: tokens.COMMA, Literal: []byte(",")},
				{Type: tokens.URI, Literal: []byte(`url('image2.png')`)},
				{Type: tokens.COMMA, Literal: []byte(",")},
				{Type: tokens.URI, Literal: []byte(`url(image3.gif)`)},
				{Type: tokens.SEMICOLON, Literal: []byte(";")},
			},
		},
	}

	runTests(t, tests)
}

func TestCSSVariables(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []tokens.Token
	}{
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
	}

	runTests(t, tests)
}

func TestCalcFunction(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []tokens.Token
	}{
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
	}

	runTests(t, tests)
}

func TestEscapedCharacters(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []tokens.Token
	}{
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
	}

	runTests(t, tests)
}

func TestUnicodeAndSpecialCharacters(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []tokens.Token
	}{
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
	}

	runTests(t, tests)
}

func TestComplexPropertyValues(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []tokens.Token
	}{
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
			name: "Nested CSS selectors",
			input: `
            .parent {
                color: blue;
                .child {
                    color: red;
                    &:hover {
                        color: green;
                    }
                }
                &__element {
                    background: yellow;
                }
            }`,
			expected: []tokens.Token{
				{Type: tokens.DOT, Literal: []byte(".")},
				{Type: tokens.IDENT, Literal: []byte("parent")},
				{Type: tokens.LBRACE, Literal: []byte("{")},
				{Type: tokens.IDENT, Literal: []byte("color")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.IDENT, Literal: []byte("blue")},
				{Type: tokens.SEMICOLON, Literal: []byte(";")},
				{Type: tokens.DOT, Literal: []byte(".")},
				{Type: tokens.IDENT, Literal: []byte("child")},
				{Type: tokens.LBRACE, Literal: []byte("{")},
				{Type: tokens.IDENT, Literal: []byte("color")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.IDENT, Literal: []byte("red")},
				{Type: tokens.SEMICOLON, Literal: []byte(";")},
				{Type: tokens.AMPERSAND, Literal: []byte("&")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.IDENT, Literal: []byte("hover")},
				{Type: tokens.LBRACE, Literal: []byte("{")},
				{Type: tokens.IDENT, Literal: []byte("color")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.IDENT, Literal: []byte("green")},
				{Type: tokens.SEMICOLON, Literal: []byte(";")},
				{Type: tokens.RBRACE, Literal: []byte("}")},
				{Type: tokens.RBRACE, Literal: []byte("}")},
				{Type: tokens.AMPERSAND, Literal: []byte("&")},
				{Type: tokens.IDENT, Literal: []byte("__element")},
				{Type: tokens.LBRACE, Literal: []byte("{")},
				{Type: tokens.IDENT, Literal: []byte("background")},
				{Type: tokens.COLON, Literal: []byte(":")},
				{Type: tokens.IDENT, Literal: []byte("yellow")},
				{Type: tokens.SEMICOLON, Literal: []byte(";")},
				{Type: tokens.RBRACE, Literal: []byte("}")},
				{Type: tokens.RBRACE, Literal: []byte("}")},
			},
		},
	}
	runTests(t, tests)
}

func runTests(t *testing.T, tests []struct {
	name     string
	input    string
	expected []tokens.Token
}) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := lexer.Lex(strings.NewReader(tt.input))

			if len(tokens)-1 != len(tt.expected) {
				strs := []string{}
				for _, t := range tokens {
					strs = append(strs, string(t.Literal))
				}

				t.Fatalf("Token count mismatch. Expected %d tokens, got %d [%s]", len(tt.expected), len(tokens)-1, strings.Join(strs, " "))
			}

			for i, expected := range tt.expected {
				tok := tokens[i]
				if tok.Type != expected.Type {
					t.Errorf("Token %d: expected type %v, got %v", i, expected.Type, tok.Type)
				}
				if !bytesEqual(tok.Literal, expected.Literal) {
					t.Errorf("Token %d: expected literal %q, got %q", i, string(expected.Literal), string(tok.Literal))
				}
				if tok.Line == 0 || tok.Column == 0 {
					t.Errorf("Token %d: line or column not set. got line=%d, column=%d", i, tok.Line, tok.Column)
				}
			}
		})
	}
}

func TestLexerIllegalCases(t *testing.T) {
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
