package parser

import (
	"fmt"

	"github.com/aledsdavies/pristinecss/pkg/tokens"
)

type ParseError struct {
	Message string
	Line    int
	Column  int
	Token   tokens.Token
}

func (e ParseError) Error() string {
	return fmt.Sprintf("line %d, column %d: %s (token: %s)\n", e.Line, e.Column, e.Message, e.Token.Type)
}

func Parse(tokens []tokens.Token) (*Stylesheet, []ParseError) {
	stylesheet := &Stylesheet{Rules: make([]Node, 0)}
	visitor := NewParseVisitor(tokens)
	stylesheet.Accept(visitor)
	return stylesheet, visitor.errors
}

var _ Visitor = (*ParseVisitor)(nil)

type ParseVisitor struct {
	tokens       []tokens.Token
	position     int
	currentToken tokens.Token
	nextToken    tokens.Token
	errors       []ParseError
}

func NewParseVisitor(tokens []tokens.Token) *ParseVisitor {
	pv := &ParseVisitor{
		tokens:   tokens,
		position: 0,
		errors:   make([]ParseError, 0),
	}
	pv.advance() // Load the first token
	pv.advance() // Load the second token (now in nextToken)
	return pv
}

func (pv *ParseVisitor) advance() {
	pv.currentToken = pv.nextToken
	if pv.position < len(pv.tokens) {
		pv.nextToken = pv.tokens[pv.position]
		pv.position++
	} else {
		pv.nextToken = tokens.Token{Type: tokens.EOF}
	}
}

func (pv *ParseVisitor) currentTokenIs(tokenType tokens.TokenType) bool {
	return pv.currentToken.Type == tokenType
}

func (pv *ParseVisitor) nextTokenIs(tokenType tokens.TokenType) bool {
	return pv.nextToken.Type == tokenType
}

func (pv *ParseVisitor) consume(tokenType tokens.TokenType, errorMessage string) bool {
	if pv.currentTokenIs(tokenType) {
		pv.advance()
		return true
	}
	pv.addError(errorMessage, pv.currentToken)
	return false
}

func (pv *ParseVisitor) addError(message string, token tokens.Token) {
	pv.errors = append(pv.errors, ParseError{
		Message: message,
		Line:    token.Line,
		Column:  token.Column,
		Token:   token,
	})
}

func (pv *ParseVisitor) skipToNextRule() {
	for !pv.currentTokenIs(tokens.EOF) && !pv.currentTokenIs(tokens.RBRACE) {
		pv.advance()
	}
	if pv.currentTokenIs(tokens.RBRACE) {
		pv.advance() // Consume the '}'
	}
}

func (pv *ParseVisitor) skipToNextSemicolonOrBrace() {
	for !pv.currentTokenIs(tokens.SEMICOLON) && !pv.currentTokenIs(tokens.RBRACE) && !pv.currentTokenIs(tokens.EOF) {
		pv.advance()
	}
	if pv.currentTokenIs(tokens.SEMICOLON) {
		pv.advance() // Consume the semicolon
	}
}
