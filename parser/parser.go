package parser

import (
	"fmt"
	"io"
)

type ParseError struct {
	Message  string
	Line     int
	Column   int
	TokenLit string
}

func (e ParseError) Error() string {
	return fmt.Sprintf("line %d, column %d: %s (token: %s)", e.Line, e.Column, e.Message, e.TokenLit)
}

type Parser struct {
	lexer      *lexer
	token      Token
	stylesheet *Stylesheet
	errors     []ParseError
}

func (p *Parser) addError(err error) {
	parseErr, ok := err.(ParseError)
	if !ok {
		parseErr = ParseError{
			Message:  err.Error(),
			Line:     p.token.Line,
			Column:   p.token.Column,
			TokenLit: string(p.token.Literal),
		}
	}
	p.errors = append(p.errors, parseErr)
}

func Parse(input io.Reader) (*Stylesheet, []ParseError) {
	lexer := Read(input)
	p := &Parser{
		lexer:      lexer,
		stylesheet: NewStylesheet(),
	}

	for p.token = p.lexer.NextToken(); p.token.Type != EOF; p.token = p.lexer.NextToken() {
		p.parse()
	}

	return p.stylesheet, p.errors
}

func (p *Parser) parse() error {
	switch p.stylesheet.Type() {
	case NodeStylesheet:
		p.parseStylesheet()
	default:
		// Ignore unknown tokens for now
		return nil
	}
	return nil
}

func (p *Parser) parseStylesheet() {
	switch p.token.Type {
	case COMMENT:
		comment := &Comment{Text: p.token.Literal}
		p.stylesheet.Rules = append(p.stylesheet.Rules, comment)
	default:
		// Ignore unknown tokens for now
	}
}

func (p *Parser) parseSelector() *Selector {
	selector := &Selector{}
	selectorValue := SelectorValue{Type: Element, Value: p.token.Literal}
	selector.Selectors = append(selector.Selectors, selectorValue)

	return selector
}
