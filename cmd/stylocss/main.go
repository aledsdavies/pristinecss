package main

import (
	"github.com/aledsdavies/pristinecss/processor"
)

var templateString = `// Code generated by StyloCSS tool; DO NOT EDIT.
// This file was generated at {{.Timestamp}}.
package {{.PackageName}}

import (
    "embed"
    "context"
    "github.com/aledsdavies/pristinecss"
)

//go:embed {{.CSSPaths}}
var embeddedCss embed.FS

var cssFiles = map[string]string{
{{range $key, $value := .Files}}    "{{$key}}": "{{$value}}",
{{end}}}

var cssClasses = map[string]stylocss.CSSClass{
{{range $key, $value := .Classes}}    "{{$key}}": { Path: "{{$value.Path}}", Class: "{{$value.Class}}" },
{{end}}}

var loader = stylocss.NewLoader(cssFiles, cssClasses)

func Class(ctx context.Context, className string) (string, error) {
    return loader.Class(ctx, className)
}

func LoadedFiles(ctx context.Context) []string {
    return loader.LoadedFiles(ctx)
}`

func main() {
	processor.Process("./", processor.WithVerbose(true))
}

/*
func main() {
	tmpl, err := template.New("cssTemplate").Parse(templateString) // templateString is your template defined earlier
	if err != nil {
		log.Fatalf("Error parsing template: %v", err)
	}

	// Prepare data for the template
	data := struct {
        PackageName string
		CSSPaths string
		Files    map[string]string
		Classes  map[string]stylocss.CSSClass
		Timestamp string
	}{
        PackageName: "css",
		CSSPaths: "styles/reset.css,styles/main.css",
		Files: map[string]string{
			"reset.css": "/d41d8cd98f/reset.css",
			"main.css":  "/d3djskdjifc/main.css",
		},
		Classes: map[string]stylocss.CSSClass {
			"main.button": {Path: "main.css", Class: "button_234mjd"},
			"main.header": {Path: "main.css", Class: "header_233443"},
		},
		Timestamp: time.Now().Format(time.RFC1123),
	}

    // Ensure the build directory exists.
    if err := os.MkdirAll("./build", 0755); err != nil {
        log.Fatalf("Failed to create build directory: %v", err)
    }

	file, err := os.Create("build/output.go")
	if err != nil {
		log.Fatalf("Error creating file: %v", err)
	}
	defer file.Close()

	err = tmpl.Execute(file, data)
	if err != nil {
		log.Fatalf("Error executing template: %v", err)
	}
}
*/
