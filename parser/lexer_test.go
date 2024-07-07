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
				{Type: parser.COMMENT, Literal: " This is a comment "},
			},
		},
		{
			name:  "Media Query",
			input: "@media screen and (max-width: 600px) { body { font-size: 14px; } }",
			expected: []parser.Token{
				{Type: parser.AT, Literal: "@"},
				{Type: parser.IDENT, Literal: "media"},
				{Type: parser.IDENT, Literal: "screen"},
				{Type: parser.IDENT, Literal: "and"},
				{Type: parser.LPAREN, Literal: "("},
				{Type: parser.IDENT, Literal: "max-width"},
				{Type: parser.COLON, Literal: ":"},
				{Type: parser.NUMBER, Literal: "600"},
				{Type: parser.UNIT, Literal: "px"},
				{Type: parser.RPAREN, Literal: ")"},
				{Type: parser.LBRACE, Literal: "{"},
				{Type: parser.IDENT, Literal: "body"},
				{Type: parser.LBRACE, Literal: "{"},
				{Type: parser.IDENT, Literal: "font-size"},
				{Type: parser.COLON, Literal: ":"},
				{Type: parser.NUMBER, Literal: "14"},
				{Type: parser.UNIT, Literal: "px"},
				{Type: parser.SEMICOLON, Literal: ";"},
				{Type: parser.RBRACE, Literal: "}"},
				{Type: parser.RBRACE, Literal: "}"},
			},
		},
		{
			name:  "Complex Selectors",
			input: "#main > .article p:first-child { color: #ff0000; }",
			expected: []parser.Token{
				{Type: parser.SELECTOR, Literal: "#main"},
				{Type: parser.GREATER, Literal: ">"},
				{Type: parser.SELECTOR, Literal: ".article"},
				{Type: parser.SELECTOR, Literal: "p:first-child"},
				{Type: parser.LBRACE, Literal: "{"},
				{Type: parser.IDENT, Literal: "color"},
				{Type: parser.COLON, Literal: ":"},
				{Type: parser.COLOR, Literal: "#ff0000"},
				{Type: parser.SEMICOLON, Literal: ";"},
				{Type: parser.RBRACE, Literal: "}"},
			},
		},
		{
			name:  "Attribute Selector",
			input: "a[href^=\"https://\"] { color: green; }",
			expected: []parser.Token{
				{Type: parser.SELECTOR, Literal: "a[href^=\"https://\"]"},
				{Type: parser.LBRACE, Literal: "{"},
				{Type: parser.IDENT, Literal: "color"},
				{Type: parser.COLON, Literal: ":"},
				{Type: parser.IDENT, Literal: "green"},
				{Type: parser.SEMICOLON, Literal: ";"},
				{Type: parser.RBRACE, Literal: "}"},
			},
		},
		{
			name:  "Keyframes",
			input: "@keyframes fadeIn { 0% { opacity: 0; } 100% { opacity: 1; } }",
			expected: []parser.Token{
				{Type: parser.AT, Literal: "@"},
				{Type: parser.IDENT, Literal: "keyframes"},
				{Type: parser.IDENT, Literal: "fadeIn"},
				{Type: parser.LBRACE, Literal: "{"},
				{Type: parser.NUMBER, Literal: "0"},
				{Type: parser.UNIT, Literal: "%"},
				{Type: parser.LBRACE, Literal: "{"},
				{Type: parser.IDENT, Literal: "opacity"},
				{Type: parser.COLON, Literal: ":"},
				{Type: parser.NUMBER, Literal: "0"},
				{Type: parser.SEMICOLON, Literal: ";"},
				{Type: parser.RBRACE, Literal: "}"},
				{Type: parser.NUMBER, Literal: "100"},
				{Type: parser.UNIT, Literal: "%"},
				{Type: parser.LBRACE, Literal: "{"},
				{Type: parser.IDENT, Literal: "opacity"},
				{Type: parser.COLON, Literal: ":"},
				{Type: parser.NUMBER, Literal: "1"},
				{Type: parser.SEMICOLON, Literal: ";"},
				{Type: parser.RBRACE, Literal: "}"},
				{Type: parser.RBRACE, Literal: "}"},
			},
		},
		{
			name:  "CSS Variables",
			input: ":root { --main-color: blue; } body { color: var(--main-color); }",
			expected: []parser.Token{
				{Type: parser.SELECTOR, Literal: ":root"},
				{Type: parser.LBRACE, Literal: "{"},
				{Type: parser.IDENT, Literal: "--main-color"},
				{Type: parser.COLON, Literal: ":"},
				{Type: parser.IDENT, Literal: "blue"},
				{Type: parser.SEMICOLON, Literal: ";"},
				{Type: parser.RBRACE, Literal: "}"},
				{Type: parser.SELECTOR, Literal: "body"},
				{Type: parser.LBRACE, Literal: "{"},
				{Type: parser.IDENT, Literal: "color"},
				{Type: parser.COLON, Literal: ":"},
				{Type: parser.IDENT, Literal: "var"},
				{Type: parser.LPAREN, Literal: "("},
				{Type: parser.IDENT, Literal: "--main-color"},
				{Type: parser.RPAREN, Literal: ")"},
				{Type: parser.SEMICOLON, Literal: ";"},
				{Type: parser.RBRACE, Literal: "}"},
			},
		},
		{
			name:  "Calc Function",
			input: "div { width: calc(100% - 20px); height: 100vh; }",
			expected: []parser.Token{
				{Type: parser.SELECTOR, Literal: "div"},
				{Type: parser.LBRACE, Literal: "{"},
				{Type: parser.IDENT, Literal: "width"},
				{Type: parser.COLON, Literal: ":"},
				{Type: parser.IDENT, Literal: "calc"},
				{Type: parser.LPAREN, Literal: "("},
				{Type: parser.NUMBER, Literal: "100"},
				{Type: parser.UNIT, Literal: "%"},
				{Type: parser.MINUS, Literal: "-"},
				{Type: parser.NUMBER, Literal: "20"},
				{Type: parser.UNIT, Literal: "px"},
				{Type: parser.RPAREN, Literal: ")"},
				{Type: parser.SEMICOLON, Literal: ";"},
				{Type: parser.IDENT, Literal: "height"},
				{Type: parser.COLON, Literal: ":"},
				{Type: parser.NUMBER, Literal: "100"},
				{Type: parser.UNIT, Literal: "vh"},
				{Type: parser.SEMICOLON, Literal: ";"},
				{Type: parser.RBRACE, Literal: "}"},
			},
		},
		{
			name:  "Escaped characters in identifiers",
			input: ".foo\\.bar { color: red; }",
			expected: []parser.Token{
				{Type: parser.SELECTOR, Literal: ".foo\\.bar"},
				{Type: parser.LBRACE, Literal: "{"},
				{Type: parser.IDENT, Literal: "color"},
				{Type: parser.COLON, Literal: ":"},
				{Type: parser.IDENT, Literal: "red"},
				{Type: parser.SEMICOLON, Literal: ";"},
				{Type: parser.RBRACE, Literal: "}"},
			},
		},
		{
			name:  "Escaped characters in attribute selectors",
			input: "a[href=\"foo\\\"bar\"] { color: blue; }",
			expected: []parser.Token{
				{Type: parser.SELECTOR, Literal: "a[href^=\"foo\\\"bar\"]"},
				{Type: parser.LBRACE, Literal: "{"},
				{Type: parser.IDENT, Literal: "color"},
				{Type: parser.COLON, Literal: ":"},
				{Type: parser.IDENT, Literal: "blue"},
				{Type: parser.SEMICOLON, Literal: ";"},
				{Type: parser.RBRACE, Literal: "}"},
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
			ignore: true,
			expected: []parser.Token{
				{Type: parser.SELECTOR, Literal: "#☃"},
				{Type: parser.LBRACE, Literal: "{"},
				{Type: parser.IDENT, Literal: "color"},
				{Type: parser.COLON, Literal: ":"},
				{Type: parser.IDENT, Literal: "skyblue"},
				{Type: parser.SEMICOLON, Literal: ";"},
				{Type: parser.RBRACE, Literal: "}"},
				{Type: parser.SELECTOR, Literal: ".günther"},
				{Type: parser.COMMA, Literal: ","},
				{Type: parser.SELECTOR, Literal: "#π_value"},
				{Type: parser.COMMA, Literal: ","},
				{Type: parser.SELECTOR, Literal: ".こんにちは"},
				{Type: parser.LBRACE, Literal: "{"},
				{Type: parser.IDENT, Literal: "color"},
				{Type: parser.COLON, Literal: ":"},
				{Type: parser.IDENT, Literal: "blue"},
				{Type: parser.SEMICOLON, Literal: ";"},
				{Type: parser.RBRACE, Literal: "}"},
				{Type: parser.DOT, Literal: "."},
				{Type: parser.IDENT, Literal: "--custom-prop"},
				{Type: parser.COMMA, Literal: ","},
				{Type: parser.DOT, Literal: "."},
				{Type: parser.IDENT, Literal: "-moz-custom"},
				{Type: parser.LBRACE, Literal: "{"},
				{Type: parser.IDENT, Literal: "value"},
				{Type: parser.COLON, Literal: ":"},
				{Type: parser.NUMBER, Literal: "123"},
				{Type: parser.SEMICOLON, Literal: ";"},
				{Type: parser.RBRACE, Literal: "}"},
				{Type: parser.HASH, Literal: "#"},
				{Type: parser.IDENT, Literal: "&ABC"},
				{Type: parser.LBRACE, Literal: "{"},
				{Type: parser.COMMENT, Literal: " Escaped ASCII "},
				{Type: parser.IDENT, Literal: "color"},
				{Type: parser.COLON, Literal: ":"},
				{Type: parser.IDENT, Literal: "red"},
				{Type: parser.SEMICOLON, Literal: ";"},
				{Type: parser.RBRACE, Literal: "}"},
			},
		},
		{
			name:  "CSS custom properties (variables)",
			input: ":root { --custom-color: #ff00ff; } .foo { color: var(--custom-color); }",
			expected: []parser.Token{
				{Type: parser.SELECTOR, Literal: ":root"},
				{Type: parser.LBRACE, Literal: "{"},
				{Type: parser.IDENT, Literal: "--custom-color"},
				{Type: parser.COLON, Literal: ":"},
				{Type: parser.COLOR, Literal: "#ff00ff"},
				{Type: parser.SEMICOLON, Literal: ";"},
				{Type: parser.RBRACE, Literal: "}"},
				{Type: parser.SELECTOR, Literal: ".foo"},
				{Type: parser.LBRACE, Literal: "{"},
				{Type: parser.IDENT, Literal: "color"},
				{Type: parser.COLON, Literal: ":"},
				{Type: parser.IDENT, Literal: "var"},
				{Type: parser.LPAREN, Literal: "("},
				{Type: parser.IDENT, Literal: "--custom-color"},
				{Type: parser.RPAREN, Literal: ")"},
				{Type: parser.SEMICOLON, Literal: ";"},
				{Type: parser.RBRACE, Literal: "}"},
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
				{Type: parser.SELECTOR, Literal: ".gradient"},
				{Type: parser.LBRACE, Literal: "{"},
				{Type: parser.IDENT, Literal: "background-image"},
				{Type: parser.COLON, Literal: ":"},
				{Type: parser.IDENT, Literal: "linear-gradient"},
				{Type: parser.LPAREN, Literal: "("},
				{Type: parser.NUMBER, Literal: "45"},
				{Type: parser.UNIT, Literal: "deg"},
				{Type: parser.COMMA, Literal: ","},
				{Type: parser.COLOR, Literal: "#ff0000"},
				{Type: parser.COMMA, Literal: ","},
				{Type: parser.COLOR, Literal: "#00ff00"},
				{Type: parser.RPAREN, Literal: ")"},
				{Type: parser.COMMA, Literal: ","},
				{Type: parser.IDENT, Literal: "radial-gradient"},
				{Type: parser.LPAREN, Literal: "("},
				{Type: parser.IDENT, Literal: "circle"},
				{Type: parser.COMMA, Literal: ","},
				{Type: parser.COLOR, Literal: "#0000ff"},
				{Type: parser.COMMA, Literal: ","},
				{Type: parser.COLOR, Literal: "#ffff00"},
				{Type: parser.RPAREN, Literal: ")"},
				{Type: parser.SEMICOLON, Literal: ";"},
				{Type: parser.IDENT, Literal: "font"},
				{Type: parser.COLON, Literal: ":"},
				{Type: parser.IDENT, Literal: "bold"},
				{Type: parser.NUMBER, Literal: "12"},
				{Type: parser.UNIT, Literal: "px"},
				{Type: parser.DIVIDE, Literal: "/"},
				{Type: parser.NUMBER, Literal: "14"},
				{Type: parser.UNIT, Literal: "px"},
				{Type: parser.STRING, Literal: "Helvetica"},
				{Type: parser.COMMA, Literal: ","},
				{Type: parser.IDENT, Literal: "sans-serif"},
				{Type: parser.SEMICOLON, Literal: ";"},
				{Type: parser.RBRACE, Literal: "}"},
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
				{Type: parser.COMMENT, Literal: " Pseudo-element hack "},
				{Type: parser.IDENT, Literal: "_"},
				{Type: parser.COLON, Literal: ":"},
				{Type: parser.COLON, Literal: ":"},
				{Type: parser.IDENT, Literal: "-webkit-full-page-media"},
				{Type: parser.COMMA, Literal: ","},
				{Type: parser.IDENT, Literal: "_"},
				{Type: parser.COLON, Literal: ":"},
				{Type: parser.IDENT, Literal: "future"},
				{Type: parser.COMMA, Literal: ","},
				{Type: parser.COLON, Literal: ":"},
				{Type: parser.IDENT, Literal: "root"},
				{Type: parser.IDENT, Literal: ".foo"},
				{Type: parser.LBRACE, Literal: "{"},
				{Type: parser.IDENT, Literal: "color"},
				{Type: parser.COLON, Literal: ":"},
				{Type: parser.IDENT, Literal: "red"},
				{Type: parser.SEMICOLON, Literal: ";"},
				{Type: parser.RBRACE, Literal: "}"},

				{Type: parser.COMMENT, Literal: " Pseudo-class hack "},
				{Type: parser.IDENT, Literal: "_"},
				{Type: parser.COLON, Literal: ":"},
				{Type: parser.IDENT, Literal: "-webkit-full-screen"},
				{Type: parser.COMMA, Literal: ","},
				{Type: parser.COLON, Literal: ":"},
				{Type: parser.IDENT, Literal: "root"},
				{Type: parser.IDENT, Literal: ".bar"},
				{Type: parser.LBRACE, Literal: "{"},
				{Type: parser.IDENT, Literal: "display"},
				{Type: parser.COLON, Literal: ":"},
				{Type: parser.IDENT, Literal: "block"},
				{Type: parser.SEMICOLON, Literal: ";"},
				{Type: parser.RBRACE, Literal: "}"},

				{Type: parser.COMMENT, Literal: " Media query hack "},
				{Type: parser.AT, Literal: "@"},
				{Type: parser.IDENT, Literal: "media"},
				{Type: parser.IDENT, Literal: "screen"},
				{Type: parser.IDENT, Literal: "and"},
				{Type: parser.LPAREN, Literal: "("},
				{Type: parser.IDENT, Literal: "min-width"},
				{Type: parser.COLON, Literal: ":"},
				{Type: parser.NUMBER, Literal: "0"},
				{Type: parser.IDENT, Literal: "\\0"},
				{Type: parser.RPAREN, Literal: ")"},
				{Type: parser.LBRACE, Literal: "{"},
				{Type: parser.IDENT, Literal: ".baz"},
				{Type: parser.LBRACE, Literal: "{"},
				{Type: parser.IDENT, Literal: "zoom"},
				{Type: parser.COLON, Literal: ":"},
				{Type: parser.NUMBER, Literal: "1"},
				{Type: parser.SEMICOLON, Literal: ";"},
				{Type: parser.RBRACE, Literal: "}"},
				{Type: parser.RBRACE, Literal: "}"},

				{Type: parser.COMMENT, Literal: " Property value hacks "},
				{Type: parser.IDENT, Literal: ".hack1"},
				{Type: parser.LBRACE, Literal: "{"},
				{Type: parser.IDENT, Literal: "property"},
				{Type: parser.COLON, Literal: ":"},
				{Type: parser.IDENT, Literal: "value"},
				{Type: parser.IDENT, Literal: "\\9"},
				{Type: parser.SEMICOLON, Literal: ";"},
				{Type: parser.RBRACE, Literal: "}"},

				{Type: parser.IDENT, Literal: ".hack2"},
				{Type: parser.LBRACE, Literal: "{"},
				{Type: parser.IDENT, Literal: "property"},
				{Type: parser.COLON, Literal: ":"},
				{Type: parser.IDENT, Literal: "value"},
				{Type: parser.IDENT, Literal: "\\9"},
				{Type: parser.SEMICOLON, Literal: ";"},
				{Type: parser.RBRACE, Literal: "}"},

				{Type: parser.IDENT, Literal: ".hack3"},
				{Type: parser.LBRACE, Literal: "{"},
				{Type: parser.IDENT, Literal: "property"},
				{Type: parser.COLON, Literal: ":"},
				{Type: parser.IDENT, Literal: "value"},
				{Type: parser.IDENT, Literal: "\\0"},
				{Type: parser.SEMICOLON, Literal: ";"},
				{Type: parser.RBRACE, Literal: "}"},

				{Type: parser.IDENT, Literal: ".hack4"},
				{Type: parser.LBRACE, Literal: "{"},
				{Type: parser.IDENT, Literal: "property"},
				{Type: parser.COLON, Literal: ":"},
				{Type: parser.IDENT, Literal: "value"},
				{Type: parser.IDENT, Literal: "\\0"},
				{Type: parser.SEMICOLON, Literal: ";"},
				{Type: parser.RBRACE, Literal: "}"},

				{Type: parser.COMMENT, Literal: " Combination of hacks "},
				{Type: parser.IDENT, Literal: "_"},
				{Type: parser.COLON, Literal: ":"},
				{Type: parser.IDENT, Literal: "-ms-lang"},
				{Type: parser.LPAREN, Literal: "("},
				{Type: parser.IDENT, Literal: "x"},
				{Type: parser.RPAREN, Literal: ")"},
				{Type: parser.COMMA, Literal: ","},
				{Type: parser.IDENT, Literal: "_"},
				{Type: parser.COLON, Literal: ":"},
				{Type: parser.IDENT, Literal: "-webkit-full-screen"},
				{Type: parser.COMMA, Literal: ","},
				{Type: parser.IDENT, Literal: ".multi-hack"},
				{Type: parser.LBRACE, Literal: "{"},
				{Type: parser.IDENT, Literal: "*display"},
				{Type: parser.COLON, Literal: ":"},
				{Type: parser.IDENT, Literal: "inline"},
				{Type: parser.SEMICOLON, Literal: ";"},
				{Type: parser.IDENT, Literal: "_height"},
				{Type: parser.COLON, Literal: ":"},
				{Type: parser.NUMBER, Literal: "1"},
				{Type: parser.UNIT, Literal: "%"},
				{Type: parser.SEMICOLON, Literal: ";"},
				{Type: parser.RBRACE, Literal: "}"},
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
					t.Errorf("tests[%d] - tokentype wrong. expected=%q, got=%q %s", i, expected.Type, tok.Type, tok.Literal)
				}
				if tok.Literal != expected.Literal {
					t.Errorf("tests[%d] - literal wrong. expected=%q, got=%q", i, expected.Literal, tok.Literal)
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
			input: "color: #1234ZZ;",
			expected: []parser.Token{
				{Type: parser.IDENT, Literal: "color"},
				{Type: parser.COLON, Literal: ":"},
				{Type: parser.COLOR, Literal: "#1234"},
				{Type: parser.IDENT, Literal: "ZZ"},
				{Type: parser.SEMICOLON, Literal: ";"},
			},
		},
		{
			name:  "Invalid unit combination",
			input: "width: 50+px;",
			expected: []parser.Token{
				{Type: parser.IDENT, Literal: "width"},
				{Type: parser.COLON, Literal: ":"},
				{Type: parser.NUMBER, Literal: "50"},
				{Type: parser.PLUS, Literal: "+"},
				{Type: parser.IDENT, Literal: "px"},
				{Type: parser.SEMICOLON, Literal: ";"},
			},
		},
		{
			name:  "Invalid percentage",
			input: "height: 100vh%;",
			expected: []parser.Token{
				{Type: parser.IDENT, Literal: "height"},
				{Type: parser.COLON, Literal: ":"},
				{Type: parser.NUMBER, Literal: "100"},
				{Type: parser.UNIT, Literal: "vh"},
				{Type: parser.ILLEGAL, Literal: "%"},
				{Type: parser.SEMICOLON, Literal: ";"},
			},
		},
		// Add more test cases here
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := parser.NewLexer(strings.NewReader(tt.input))
			for i, expected := range tt.expected {
				got := l.NextToken()
				if got.Type != expected.Type {
					t.Errorf("Token %d: expected type %v, got %v", i, expected.Type, got.Type)
				}
				if got.Literal != expected.Literal {
					t.Errorf("Token %d: expected literal %q, got %q", i, expected.Literal, got.Literal)
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
					t.Errorf("Found an ILLEGAL token: %v", tok)
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
