package parser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aledsdavies/pristinecss/pkg/lexer"
)

func TestCanParseFrameworksWithoutError(t *testing.T) {
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
		t.Run(fw.name, func(t *testing.T) {
			file, err := os.Open(fw.path)
			if err != nil {
				t.Fatalf("Could not open the file %s: %v", fw.path, err)
			}
			defer file.Close()

			tokens := lexer.Lex(file)
			_, errors := Parse(tokens)

			if len(errors) > 0 {
				t.Errorf("Parsing %s produced %d errors:", fw.name, len(errors))
				for _, err := range errors {
					t.Errorf("  - %v", err)
				}
			}
		})
	}
}
