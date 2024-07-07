package parser_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aledsdavies/pristinecss/parser"
	"github.com/google/go-cmp/cmp"
)

// TestParseStyleSheetFromFile tests CSS parsing from file paths.
func TestParseStyleSheetFromFile(t *testing.T) {
	tests := []struct {
		name     string
		filepath string
		expected *parser.StyleSheet
	}{
		{
			name:     "Can Parse Empty CSS File",
			filepath: filepath.Join("..", "test-data", "empty.css"),
			expected: &parser.StyleSheet{
			},
		},
		{
			name:     "Can Parse Attribute Selectors",
			filepath: filepath.Join("..", "test-data", "element_selector.css"),
			expected: &parser.StyleSheet{
				Rules: []parser.Node{
					&parser.Selector{
						Type:         parser.TypeElement,
						Name:         "p",
						Declarations: map[string]string{"color": "blue"},
					},
				},
			},
		},
//003   class_selector.css
//004   element_selector.css
//005   id_selector.css
//007   psudo_element_selector.css
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := os.Open(tt.filepath)
			if err != nil {
				t.Fatalf("Failed to read file %s: %v", tt.filepath, err)
			}

			parsedResult := parser.Parse(data)
			if parsedResult.HasErrors() {
				t.Fatalf("Parsing errors in test %s: %+v", tt.name, parsedResult.Errors)
			}

			if diff := cmp.Diff(tt.expected, parsedResult.StyleSheet); diff != "" {
				t.Errorf("Test %s failed: (-want +got):\n%s", tt.name, diff)
			}
		})
	}
}
