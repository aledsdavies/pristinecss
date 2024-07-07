package parser

import (
	"bufio"
	"io"
	"strings"
)

type ParseResult struct {
	StyleSheet  *StyleSheet
	Errors []*ParseError
	ContextInfo map[string]string
}

func (p *ParseResult) HasErrors() bool {
    return len(p.Errors) > 1
}

type ParseError struct {
	Type     string
	Severity string
	Message  string
	Line     int
	Column   int
	Snippet  string
}

type ContextInfo map[string]string

func Parse(reader io.Reader, contextInfo ...ContextInfo) *ParseResult {
	sheet := &StyleSheet{}
	diags := []*ParseError{}

	// Merge all context information
	ctx := make(map[string]string)
	for _, info := range contextInfo {
		for key, value := range info {
			ctx[key] = value
		}
	}

    parseCSS(reader, sheet, diags)

	return &ParseResult{
		StyleSheet:  sheet,
		Errors: diags,
		ContextInfo: ctx,
	}
}


func parseCSS(reader io.Reader, sheet *StyleSheet, diags []*ParseError) {
	scanner := bufio.NewScanner(reader)
	var selector *Selector
	var lineNum int
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "/*") {
			continue // Skip empty lines and comments
		}
		if strings.HasSuffix(line, "{") {
			// Parse selector
			selectorName := strings.TrimSpace(line[:len(line)-1])
			selector = &Selector{
				Type:         TypeElement,
				Name:         selectorName,
				Declarations: make(map[string]string),
			}
		} else if strings.HasSuffix(line, "}") {
			// End of selector block
			sheet.Rules = append(sheet.Rules, selector)
			selector = nil
		} else if selector != nil {
			// Parse declaration
			parts := strings.SplitN(line, ":", 2)
			if len(parts) != 2 {
				diags = append(diags, &ParseError{
					Type:     "Syntax Error",
					Severity: "Error",
					Message:  "Invalid declaration",
					Line:     lineNum,
					Snippet:  line,
				})
				continue
			}
			prop := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(strings.TrimSuffix(parts[1], ";"))
			selector.Declarations[prop] = val
		} else {
			diags = append(diags, &ParseError{
				Type:     "Syntax Error",
				Severity: "Error",
				Message:  "Unexpected content outside of a selector block",
				Line:     lineNum,
				Snippet:  line,
			})
		}
	}

	if err := scanner.Err(); err != nil {
		diags = append(diags, &ParseError{
			Type:     "IO Error",
			Severity: "Error",
			Message:  err.Error(),
		})
	}
}
