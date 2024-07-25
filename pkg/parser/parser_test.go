package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aledsdavies/pristinecss/pkg/lexer"
	"github.com/google/go-cmp/cmp"
)

// Benchmark for parsing various CSS frameworks
func BenchmarkParseFrameworks(b *testing.B) {
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
		content, err := os.Open(fw.path)
		if err != nil {
			b.Fatalf("Could not read the file %s: %v", fw.path, err)
		}
		tokens := lexer.Lex(content)

		b.Run(fw.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				Parse(tokens)
			}
		})
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *Stylesheet
	}{
		{
			name:  "Empty stylesheet",
			input: "",
			expected: &Stylesheet{
				Rules: []Node{},
			},
		},
		{
			name:  "Comment only",
			input: "/* This is a comment */",
			expected: &Stylesheet{
				Rules: []Node{
					&Comment{Text: []byte(" This is a comment ")},
				},
			},
		},
		{
			name:  "Simple element selector",
			input: "div { color: blue; }",
			expected: &Stylesheet{
				Rules: []Node{
					&Selector{
						Selectors: []SelectorValue{{Type: Element, Value: []byte("div")}},
						Rules: []Node{
							&Declaration{Key: []byte("color"), Value: [][]byte{[]byte("blue")}},
						},
					},
				},
			},
		},
		{
			name:  "Class selector",
			input: ".highlight { background-color: yellow; }",
			expected: &Stylesheet{
				Rules: []Node{
					&Selector{
						Selectors: []SelectorValue{{Type: Class, Value: []byte(".highlight")}},
						Rules: []Node{
							&Declaration{Key: []byte("background-color"), Value: [][]byte{[]byte("yellow")}},
						},
					},
				},
			},
		},
		{
			name:  "ID selector",
			input: "#main { font-size: 16px; }",
			expected: &Stylesheet{
				Rules: []Node{
					&Selector{
						Selectors: []SelectorValue{{Type: ID, Value: []byte("#main")}},
						Rules: []Node{
							&Declaration{Key: []byte("font-size"), Value: [][]byte{[]byte("16"), []byte("px")}},
						},
					},
				},
			},
		},
		{
			name:  "Attribute selector",
			input: "[type='text'] { border: 1px solid gray; }",
			expected: &Stylesheet{
				Rules: []Node{
					&Selector{
						Selectors: []SelectorValue{{Type: Attribute, Value: []byte("[type='text']")}},
						Rules: []Node{
							&Declaration{Key: []byte("border"), Value: [][]byte{[]byte("1"), []byte("px"), []byte("solid"), []byte("gray")}},
						},
					},
				},
			},
		},
		{
			name:  "Compound selector",
			input: "div.container { max-width: 1200px; }",
			expected: &Stylesheet{
				Rules: []Node{
					&Selector{
						Selectors: []SelectorValue{
							{Type: Element, Value: []byte("div")},
							{Type: Class, Value: []byte(".container")},
						},
						Rules: []Node{
							&Declaration{Key: []byte("max-width"), Value: [][]byte{[]byte("1200"), []byte("px")}},
						},
					},
				},
			},
		},
		{
			name:  "Multiple selectors",
			input: "h1, h2, h3 { font-family: sans-serif; }",
			expected: &Stylesheet{
				Rules: []Node{
					&Selector{
						Selectors: []SelectorValue{
							{Type: Element, Value: []byte("h1")},
							{Type: Combinator, Value: []byte(",")},
							{Type: Element, Value: []byte("h2")},
							{Type: Combinator, Value: []byte(",")},
							{Type: Element, Value: []byte("h3")},
						},
						Rules: []Node{
							&Declaration{Key: []byte("font-family"), Value: [][]byte{[]byte("sans-serif")}},
						},
					},
				},
			},
		},
		{
			name:  "Descendant combinator",
			input: "article p { line-height: 1.5; }",
			expected: &Stylesheet{
				Rules: []Node{
					&Selector{
						Selectors: []SelectorValue{
							{Type: Element, Value: []byte("article")},
							{Type: Element, Value: []byte("p")},
						},
						Rules: []Node{
							&Declaration{Key: []byte("line-height"), Value: [][]byte{[]byte("1.5")}},
						},
					},
				},
			},
		},
		{
			name:  "Child combinator",
			input: "ul > li { list-style-type: square; }",
			expected: &Stylesheet{
				Rules: []Node{
					&Selector{
						Selectors: []SelectorValue{
							{Type: Element, Value: []byte("ul")},
							{Type: Combinator, Value: []byte(">")},
							{Type: Element, Value: []byte("li")},
						},
						Rules: []Node{
							&Declaration{Key: []byte("list-style-type"), Value: [][]byte{[]byte("square")}},
						},
					},
				},
			},
		},
		{
			name:  "Adjacent sibling combinator",
			input: "h1 + p { font-weight: bold; }",
			expected: &Stylesheet{
				Rules: []Node{
					&Selector{
						Selectors: []SelectorValue{
							{Type: Element, Value: []byte("h1")},
							{Type: Combinator, Value: []byte("+")},
							{Type: Element, Value: []byte("p")},
						},
						Rules: []Node{
							&Declaration{Key: []byte("font-weight"), Value: [][]byte{[]byte("bold")}},
						},
					},
				},
			},
		},
		{
			name:  "General sibling combinator",
			input: "h1 ~ p { margin-top: 1em; }",
			expected: &Stylesheet{
				Rules: []Node{
					&Selector{
						Selectors: []SelectorValue{
							{Type: Element, Value: []byte("h1")},
							{Type: Combinator, Value: []byte("~")},
							{Type: Element, Value: []byte("p")},
						},
						Rules: []Node{
							&Declaration{Key: []byte("margin-top"), Value: [][]byte{[]byte("1"), []byte("em")}},
						},
					},
				},
			},
		},
		{
			name:  "Complex selector",
			input: "div.container > ul.list li:first-child { color: red; }",
			expected: &Stylesheet{
				Rules: []Node{
					&Selector{
						Selectors: []SelectorValue{
							{Type: Element, Value: []byte("div")},
							{Type: Class, Value: []byte(".container")},
							{Type: Combinator, Value: []byte(">")},
							{Type: Element, Value: []byte("ul")},
							{Type: Class, Value: []byte(".list")},
							{Type: Element, Value: []byte("li")},
							{Type: Pseudo, Value: []byte(":first-child")},
						},
						Rules: []Node{
							&Declaration{Key: []byte("color"), Value: [][]byte{[]byte("red")}},
						},
					},
				},
			},
		},
		{
			name:  "Pseudo-element selector",
			input: "p::first-line { color: red; font-variant: small-caps; }",
			expected: &Stylesheet{
				Rules: []Node{
					&Selector{
						Selectors: []SelectorValue{
							{Type: Element, Value: []byte("p")},
							{Type: Pseudo, Value: []byte("::first-line")},
						},
						Rules: []Node{
							&Declaration{Key: []byte("color"), Value: [][]byte{[]byte("red")}},
							&Declaration{Key: []byte("font-variant"), Value: [][]byte{[]byte("small-caps")}},
						},
					},
				},
			},
		},
		{
			name:  "Pseudo-class and pseudo-element selector",
			input: "a:hover::before { content: '→'; margin-right: 5px; }",
			expected: &Stylesheet{
				Rules: []Node{
					&Selector{
						Selectors: []SelectorValue{
							{Type: Element, Value: []byte("a")},
							{Type: Pseudo, Value: []byte(":hover")},
							{Type: Pseudo, Value: []byte("::before")},
						},
						Rules: []Node{
							&Declaration{Key: []byte("content"), Value: [][]byte{[]byte("'→'")}},
							&Declaration{Key: []byte("margin-right"), Value: [][]byte{[]byte("5"), []byte("px")}},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := lexer.Lex(strings.NewReader(tt.input))
			result, errors := Parse(tokens)
			if len(errors) > 0 {
				t.Errorf("Unexpected errors: %v", errors)
			}
			diffStylesheet(t, tt.expected, result)
		})
	}
}

// Custom comparer function
func stylesheetComparer() cmp.Option {
	return cmp.Comparer(func(x, y *Stylesheet) bool {
		return x.String() == y.String()
	})
}

// Custom diff function
func diffStylesheet(t *testing.T, expected, actual *Stylesheet) {
	t.Helper()

	expectedStr := expected.String()
	actualStr := actual.String()

	if expectedStr == actualStr {
		return
	}

	expectedLines := strings.Split(expectedStr, "\n")
	actualLines := strings.Split(actualStr, "\n")

	diff := []string{}

	i, j := 0, 0
	for i < len(expectedLines) && j < len(actualLines) {
		if expectedLines[i] == actualLines[j] {
			diff = append(diff, "  "+expectedLines[i])
			i++
			j++
		} else {
			// Find the next matching line
			nextMatch := findNextMatch(expectedLines[i:], actualLines[j:])

			// Add removed lines
			for k := 0; k < nextMatch.expectedOffset; k++ {
				diff = append(diff, "-"+expectedLines[i+k])
			}

			// Add added lines
			for k := 0; k < nextMatch.actualOffset; k++ {
				diff = append(diff, "+"+actualLines[j+k])
			}

			i += nextMatch.expectedOffset
			j += nextMatch.actualOffset
		}
	}

	// Handle remaining lines
	for ; i < len(expectedLines); i++ {
		diff = append(diff, "-"+expectedLines[i])
	}
	for ; j < len(actualLines); j++ {
		diff = append(diff, "+"+actualLines[j])
	}

	// Adjust indentation for diff lines
	for i, line := range diff {
		indent := countLeadingSpaces(line[1:]) // Ignore the first character (-, +, or space)
		if line[0] == '-' || line[0] == '+' {
			diff[i] = line[:1] + strings.Repeat(" ", indent) + line[1+indent:]
		}
	}

	t.Errorf("CSS mismatch (-want +got):\n%s", strings.Join(diff, "\n"))
}

func countLeadingSpaces(s string) int {
	return len(s) - len(strings.TrimLeft(s, " "))
}

type matchOffset struct {
	expectedOffset int
	actualOffset   int
}

func findNextMatch(expected, actual []string) matchOffset {
	for i := 0; i < len(expected); i++ {
		for j := 0; j < len(actual); j++ {
			if expected[i] == actual[j] {
				return matchOffset{i, j}
			}
		}
	}
	return matchOffset{len(expected), len(actual)}
}
