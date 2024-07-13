package parser

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

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
		content, err := os.ReadFile(fw.path)
		if err != nil {
			b.Fatalf("Could not read the file %s: %v", fw.path, err)
		}

		b.Run(fw.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				reader := bytes.NewReader(content)
				Parse(reader)
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
					&Comment{Text: []byte("/* This is a comment */")},
				},
			},
		},
		{
			name:  "Simple selector without declarations",
			input: "div { }",
			expected: &Stylesheet{
				Rules: []Node{
					&Selector{
						Selectors: []SelectorValue{{Type: Element, Value: []byte("div")}},
						Rules:     []Node{},
					},
				},
			},
		},
		{
			name: "Simple CSS with declarations (ignored for now)",
			input: `
                body {
                    color: red;
                    font-size: 16px;
                }
            `,
			expected: &Stylesheet{
				Rules: []Node{
					&Selector{
						Selectors: []SelectorValue{{Type: Element, Value: []byte("body")}},
						Rules:     []Node{},
					},
				},
			},
		},
		{
			name:  "Multiple selectors",
			input: "div, p { color: blue; }",
			expected: &Stylesheet{
				Rules: []Node{
					&Selector{
						Selectors: []SelectorValue{
							{Type: Element, Value: []byte("div")},
							{Type: Element, Value: []byte("p")},
						},
						Rules: []Node{},
					},
				},
			},
		},
		{
			name:  "Complex selectors",
			input: ".class #id[attr] { margin: 10px; }",
			expected: &Stylesheet{
				Rules: []Node{
					&Selector{
						Selectors: []SelectorValue{
							{Type: Class, Value: []byte(".class")},
							{Type: ID, Value: []byte("#id")},
							{Type: Attribute, Value: []byte("[attr]")},
						},
						Rules: []Node{},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, errors := Parse(strings.NewReader(tt.input))
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
	diff := cmp.Diff(expected, actual, stylesheetComparer())
	if diff != "" {
		t.Errorf("CSS mismatch (-want +got):\n%s", diff)
	}
}
