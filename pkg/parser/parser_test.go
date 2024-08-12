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
		{"Bootstrap", filepath.Join("..", "..", "test-data", "frameworks", "bootstrap.css")},
		{"Bulma", filepath.Join("..", "..", "test-data", "frameworks", "bulma.css")},
		{"Foundation", filepath.Join("..", "..", "test-data", "frameworks", "foundation.css")},
		{"Materialize", filepath.Join("..", "..", "test-data", "frameworks", "materialize.css")},
		{"Spectre", filepath.Join("..", "..", "test-data", "frameworks", "spectre.css")},
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

func TestBasicSelectors(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *Stylesheet
	}{
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
							&Declaration{Key: []byte("font-size"), Value: [][]byte{[]byte("16px")}},
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
							&Declaration{Key: []byte("border"), Value: [][]byte{[]byte("1px"), []byte("solid"), []byte("gray")}},
						},
					},
				},
			},
		},
	}

	runTests(t, tests)
}

func TestComplexSelectors(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *Stylesheet
	}{
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
							&Declaration{Key: []byte("max-width"), Value: [][]byte{[]byte("1200px")}},
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
	}

	runTests(t, tests)
}

func TestPseudoSelectors(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *Stylesheet
	}{
		{
			name:  "Pseudo-class selector",
			input: "a:hover { color: red; }",
			expected: &Stylesheet{
				Rules: []Node{
					&Selector{
						Selectors: []SelectorValue{
							{Type: Element, Value: []byte("a")},
							{Type: Pseudo, Value: []byte(":hover")},
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
			input: "p::first-line { font-weight: bold; }",
			expected: &Stylesheet{
				Rules: []Node{
					&Selector{
						Selectors: []SelectorValue{
							{Type: Element, Value: []byte("p")},
							{Type: Pseudo, Value: []byte("::first-line")},
						},
						Rules: []Node{
							&Declaration{Key: []byte("font-weight"), Value: [][]byte{[]byte("bold")}},
						},
					},
				},
			},
		},
		{
			name:  "Pseudo-class and pseudo-element selector",
			input: "a:hover::before { content: '→'; }",
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
						},
					},
				},
			},
		},
	}

	runTests(t, tests)
}

func TestMediaQueries(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *Stylesheet
	}{
		{
			name:  "Basic media query",
			input: "@media screen { body { font-size: 16px; } }",
			expected: &Stylesheet{
				Rules: []Node{
					&MediaAtRule{
						Name: []byte("media"),
						Query: MediaQuery{
							Queries: []MediaQueryExpression{
								{
									MediaType: []byte("screen"),
								},
							},
						},
						Rules: []Node{
							&Selector{
								Selectors: []SelectorValue{{Type: Element, Value: []byte("body")}},
								Rules: []Node{
									&Declaration{Key: []byte("font-size"), Value: [][]byte{[]byte("16px")}},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "Media query with condition",
			input: "@media (max-width: 600px) { .container { width: 100%; } }",
			expected: &Stylesheet{
				Rules: []Node{
					&MediaAtRule{
						Name: []byte("media"),
						Query: MediaQuery{
							Queries: []MediaQueryExpression{
								{
									Features: []MediaFeature{
										{Name: []byte("max-width"), Value: []byte("600px")},
									},
								},
							},
						},
						Rules: []Node{
							&Selector{
								Selectors: []SelectorValue{{Type: Class, Value: []byte(".container")}},
								Rules: []Node{
									&Declaration{Key: []byte("width"), Value: [][]byte{[]byte("100%")}},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "Complex media query",
			input: "@media screen and (min-width: 768px) and (max-width: 1024px) { .sidebar { display: none; } }",
			expected: &Stylesheet{
				Rules: []Node{
					&MediaAtRule{
						Name: []byte("media"),
						Query: MediaQuery{
							Queries: []MediaQueryExpression{
								{
									MediaType: []byte("screen"),
									Features: []MediaFeature{
										{Name: []byte("min-width"), Value: []byte("768px")},
										{Name: []byte("max-width"), Value: []byte("1024px")},
									},
								},
							},
						},
						Rules: []Node{
							&Selector{
								Selectors: []SelectorValue{{Type: Class, Value: []byte(".sidebar")}},
								Rules: []Node{
									&Declaration{Key: []byte("display"), Value: [][]byte{[]byte("none")}},
								},
							},
						},
					},
				},
			},
		},
	}
	runTests(t, tests)
}

func TestKeyframes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *Stylesheet
	}{
		{
			name: "Basic @keyframes rule",
			input: `@keyframes slide-in {
                from { transform: translateX(-100%); }
                to { transform: translateX(0); }
            }`,
			expected: &Stylesheet{
				Rules: []Node{
					&KeyframesAtRule{
						Name: []byte("slide-in"),
						Stops: []KeyframeStop{
							{
								Selectors: [][]byte{[]byte("from")},
								Rules: []Node{
									&Declaration{Key: []byte("transform"), Value: [][]byte{[]byte("translateX"), []byte("("), []byte("-100%"), []byte(")")}},
								},
							},
							{
								Selectors: [][]byte{[]byte("to")},
								Rules: []Node{
									&Declaration{Key: []byte("transform"), Value: [][]byte{[]byte("translateX"), []byte("("), []byte("0"), []byte(")")}},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "@keyframes rule with percentages",
			input: `@keyframes color-change {
                0% { background-color: red; }
                50% { background-color: green; }
                100% { background-color: blue; }
            }`,
			expected: &Stylesheet{
				Rules: []Node{
					&KeyframesAtRule{
						Name: []byte("color-change"),
						Stops: []KeyframeStop{
							{
								Selectors: [][]byte{[]byte("0%")},
								Rules: []Node{
									&Declaration{Key: []byte("background-color"), Value: [][]byte{[]byte("red")}},
								},
							},
							{
								Selectors: [][]byte{[]byte("50%")},
								Rules: []Node{
									&Declaration{Key: []byte("background-color"), Value: [][]byte{[]byte("green")}},
								},
							},
							{
								Selectors: [][]byte{[]byte("100%")},
								Rules: []Node{
									&Declaration{Key: []byte("background-color"), Value: [][]byte{[]byte("blue")}},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "@keyframes with multiple selectors per stop",
			input: `@keyframes multi-step {
                0%, 100% { opacity: 0; }
                25%, 75% { opacity: 0.5; }
                50% { opacity: 1; }
            }`,
			expected: &Stylesheet{
				Rules: []Node{
					&KeyframesAtRule{
						Name: []byte("multi-step"),
						Stops: []KeyframeStop{
							{
								Selectors: [][]byte{[]byte("0%"), []byte("100%")},
								Rules: []Node{
									&Declaration{Key: []byte("opacity"), Value: [][]byte{[]byte("0")}},
								},
							},
							{
								Selectors: [][]byte{[]byte("25%"), []byte("75%")},
								Rules: []Node{
									&Declaration{Key: []byte("opacity"), Value: [][]byte{[]byte("0.5")}},
								},
							},
							{
								Selectors: [][]byte{[]byte("50%")},
								Rules: []Node{
									&Declaration{Key: []byte("opacity"), Value: [][]byte{[]byte("1")}},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "@keyframes with vendor prefix",
			input: `@-webkit-keyframes bounce {
                0%, 20%, 50%, 80%, 100% { transform: translateY(0); }
                40% { transform: translateY(-30px); }
                60% { transform: translateY(-15px); }
            }`,
			expected: &Stylesheet{
				Rules: []Node{
					&KeyframesAtRule{
                        WebKitPrefix: true,
						Name: []byte("bounce"),
						Stops: []KeyframeStop{
							{
								Selectors: [][]byte{[]byte("0%"), []byte("20%"), []byte("50%"), []byte("80%"), []byte("100%")},
								Rules: []Node{
									&Declaration{Key: []byte("transform"), Value: [][]byte{[]byte("translateY"), []byte("("), []byte("0"), []byte(")")}},
								},
							},
							{
								Selectors: [][]byte{[]byte("40%")},
								Rules: []Node{
									&Declaration{Key: []byte("transform"), Value: [][]byte{[]byte("translateY"), []byte("("), []byte("-30px"), []byte(")")}},
								},
							},
							{
								Selectors: [][]byte{[]byte("60%")},
								Rules: []Node{
									&Declaration{Key: []byte("transform"), Value: [][]byte{[]byte("translateY"), []byte("("), []byte("-15px"), []byte(")")}},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "@keyframes with multiple properties per stop",
			input: `@keyframes complex-animation {
                from {
                    left: 0;
                    top: 0;
                }
                50% {
                    left: 50%;
                    top: 100px;
                    background-color: blue;
                }
                to {
                    left: 100%;
                    top: 0;
                }
            }`,
			expected: &Stylesheet{
				Rules: []Node{
					&KeyframesAtRule{
						Name: []byte("complex-animation"),
						Stops: []KeyframeStop{
							{
								Selectors: [][]byte{[]byte("from")},
								Rules: []Node{
									&Declaration{Key: []byte("left"), Value: [][]byte{[]byte("0")}},
									&Declaration{Key: []byte("top"), Value: [][]byte{[]byte("0")}},
								},
							},
							{
								Selectors: [][]byte{[]byte("50%")},
								Rules: []Node{
									&Declaration{Key: []byte("left"), Value: [][]byte{[]byte("50%")}},
									&Declaration{Key: []byte("top"), Value: [][]byte{[]byte("100px")}},
									&Declaration{Key: []byte("background-color"), Value: [][]byte{[]byte("blue")}},
								},
							},
							{
								Selectors: [][]byte{[]byte("to")},
								Rules: []Node{
									&Declaration{Key: []byte("left"), Value: [][]byte{[]byte("100%")}},
									&Declaration{Key: []byte("top"), Value: [][]byte{[]byte("0")}},
								},
							},
						},
					},
				},
			},
		},
	}
	runTests(t, tests)
}

func runTests(t *testing.T, tests []struct {
	name     string
	input    string
	expected *Stylesheet
}) {
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