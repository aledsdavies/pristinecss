package parser

import (
	"fmt"
	"strings"

	"github.com/aledsdavies/pristinecss/pkg/tokens"
)

type ParseError struct {
	Message  string
	Line     int
	Column   int
	TokenLit []byte
}

func (e ParseError) Error() string {
	return fmt.Sprintf("line %d, column %d: %s (tokens. %s)", e.Line, e.Column, e.Message, e.TokenLit)
}

func Parse(tokens []tokens.Token) (*Stylesheet, []ParseError) {
	stylesheet := &Stylesheet{Rules: make([]Node, 0)}
	visitor := NewParseVisitor(tokens)
	stylesheet.Accept(visitor)
	return stylesheet, visitor.errors
}

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
		Message:  message,
		Line:     token.Line,
		Column:   token.Column,
		TokenLit: token.Literal,
	})
}

func (pv *ParseVisitor) VisitStylesheet(s *Stylesheet) {
	for !pv.currentTokenIs(tokens.EOF) {
		switch pv.currentToken.Type {
		case tokens.COMMENT:
			comment := &Comment{Text: pv.currentToken.Literal}
			comment.Accept(pv)
			s.Rules = append(s.Rules, comment)
			pv.advance()
		case tokens.DOT, tokens.HASH, tokens.COLON, tokens.DBLCOLON, tokens.IDENT, tokens.LBRACKET:
			selector := &Selector{
				Selectors: make([]SelectorValue, 0),
				Rules:     make([]Node, 0),
			}
			selector.Accept(pv)
			s.Rules = append(s.Rules, selector)
		case tokens.AT:
			atRule := &AtRule{
				Name:  pv.nextToken.Literal,
				Query: nil,
				Rules: make([]Node, 0),
			}
			atRule.Accept(pv)
			s.Rules = append(s.Rules, atRule)
		default:
			pv.addError("Unexpected token at stylesheet level", pv.currentToken)
			pv.advance()
		}
	}
}

func (pv *ParseVisitor) VisitSelector(s *Selector) {
	pv.parseSelector(s)
	pv.parseDeclarationBlock(s)
}

func (pv *ParseVisitor) VisitDeclaration(d *Declaration) {
	// This method might not be needed if declarations are always handled within selectors
}

func (pv *ParseVisitor) VisitAtRule(a *AtRule) {
	pv.advance() // Consume '@'
	pv.advance() // Consume the at-rule name

	switch string(a.Name) {
	case "media":
		a.Query = pv.parseMediaQuery()
	case "keyframes":
		a.Query = pv.parseKeyframesRule()
        return
	// Add cases for other at-rules as needed
	default:
		pv.addError("Unsupported at-rule", pv.currentToken)
		return
	}

	if !pv.consume(tokens.LBRACE, "Expected '{' after at-rule query") {
		return
	}

	// Parse the content of the media block
	for !pv.currentTokenIs(tokens.RBRACE) && !pv.currentTokenIs(tokens.EOF) {
		switch pv.currentToken.Type {
		case tokens.COMMENT:
			comment := &Comment{Text: pv.currentToken.Literal}
			comment.Accept(pv)
			a.Rules = append(a.Rules, comment)
		case tokens.DOT, tokens.HASH, tokens.COLON, tokens.DBLCOLON, tokens.IDENT, tokens.LBRACKET:
			selector := &Selector{
				Selectors: make([]SelectorValue, 0),
				Rules:     make([]Node, 0),
			}
			pv.parseSelector(selector)
			pv.parseDeclarationBlock(selector)
			a.Rules = append(a.Rules, selector)
		default:
			pv.addError("Unexpected token in at-rule block", pv.currentToken)
			pv.advance()
		}
	}

	pv.consume(tokens.RBRACE, "Expected '}' to close at-rule block")
}

func (pv *ParseVisitor) VisitComment(c *Comment) {
	// Nothing to do here, as comments are simple tokens
}

func (pv *ParseVisitor) parseSelector(s *Selector) {
	for !pv.currentTokenIs(tokens.EOF) && !pv.currentTokenIs(tokens.LBRACE) {
		switch pv.currentToken.Type {
		case tokens.IDENT:
			s.Selectors = append(s.Selectors, SelectorValue{
				Type:  Element,
				Value: pv.currentToken.Literal,
			})
			pv.advance()
		case tokens.DOT:
			if pv.nextTokenIs(tokens.IDENT) {
				pv.advance() // Consume the dot
				s.Selectors = append(s.Selectors, SelectorValue{
					Type:  Class,
					Value: append([]byte("."), pv.currentToken.Literal...),
				})
				pv.advance() // Consume the identifier
			} else {
				pv.addError("Expected identifier after '.'", pv.nextToken)
				pv.advance() // Skip the dot
			}
		case tokens.HASH:
			if pv.nextTokenIs(tokens.IDENT) {
				pv.advance() // Consume the hash
				s.Selectors = append(s.Selectors, SelectorValue{
					Type:  ID,
					Value: append([]byte("#"), pv.currentToken.Literal...),
				})
				pv.advance() // Consume the identifier
			} else {
				pv.addError("Expected identifier after '#'", pv.nextToken)
				pv.advance() // Skip the hash
			}
		case tokens.LBRACKET:
			attrSelector := pv.parseAttributeSelector()
			if attrSelector != nil {
				s.Selectors = append(s.Selectors, *attrSelector)
			}
		case tokens.COLON, tokens.DBLCOLON:
			pseudoSelector := pv.parsePseudoSelector()
			if pseudoSelector != nil {
				s.Selectors = append(s.Selectors, *pseudoSelector)
			}
		case tokens.COMMA, tokens.GREATER, tokens.PLUS, tokens.TILDE:
			s.Selectors = append(s.Selectors, SelectorValue{
				Type:  Combinator,
				Value: pv.currentToken.Literal,
			})
			pv.advance()
		default:
			pv.addError("Unexpected token in selector", pv.currentToken)
			pv.advance() // Skip unexpected token
		}
	}
}

func (pv *ParseVisitor) parseDeclarationBlock(s *Selector) {
	if !pv.consume(tokens.LBRACE, "Expected '{' after selector") {
		return
	}

	for !pv.currentTokenIs(tokens.RBRACE) && !pv.currentTokenIs(tokens.EOF) {
		declaration := pv.parseDeclaration()
		if declaration != nil {
			s.Rules = append(s.Rules, declaration)
		}

		if pv.currentTokenIs(tokens.SEMICOLON) {
			pv.advance() // Consume ';'
		}
	}

	pv.consume(tokens.RBRACE, "Expected '}' at the end of declaration block")
}

func (pv *ParseVisitor) parseDeclaration() *Declaration {
	if !pv.currentTokenIs(tokens.IDENT) {
		pv.addError("Expected property name", pv.currentToken)
		pv.skipToNextSemicolonOrBrace()
		return nil
	}

	declaration := &Declaration{
		Key:   pv.currentToken.Literal,
		Value: make([][]byte, 0),
	}
	pv.advance() // Consume property name

	if !pv.consume(tokens.COLON, "Expected ':' after property name") {
		pv.skipToNextSemicolonOrBrace()
		return nil
	}

	// Parse declaration value
	for !pv.currentTokenIs(tokens.SEMICOLON) && !pv.currentTokenIs(tokens.RBRACE) && !pv.currentTokenIs(tokens.EOF) {
		switch pv.currentToken.Type {
		case tokens.IDENT, tokens.NUMBER, tokens.PERCENTAGE, tokens.STRING, tokens.HASH:
			declaration.Value = append(declaration.Value, pv.currentToken.Literal)
		case tokens.LPAREN:
			value := pv.parseFunction()
			if value != nil {
				declaration.Value = append(declaration.Value, value...)
			}
		default:
			pv.addError("Unexpected token in declaration value", pv.currentToken)
			pv.skipToNextSemicolonOrBrace()
			return nil
		}
		pv.advance()
	}

	return declaration
}

func (pv *ParseVisitor) parseFunction() [][]byte {
	var function [][]byte
	function = append(function, pv.currentToken.Literal) // '('
	pv.advance()

	for !pv.currentTokenIs(tokens.RPAREN) && !pv.currentTokenIs(tokens.EOF) {
		switch pv.currentToken.Type {
		case tokens.IDENT, tokens.NUMBER, tokens.PERCENTAGE, tokens.STRING, tokens.HASH, tokens.COMMA:
			function = append(function, pv.currentToken.Literal)
		default:
			pv.addError("Unexpected token in function", pv.currentToken)
			return function
		}
		pv.advance()
	}

	if pv.consume(tokens.RPAREN, "Expected ')' to close function") {
		function = append(function, []byte(")"))
	}

	return function
}

func (pv *ParseVisitor) parseMediaQuery() *MediaQuery {
	mediaQuery := &MediaQuery{
		Queries: make([]MediaQueryExpression, 0),
	}

	for !pv.currentTokenIs(tokens.LBRACE) && !pv.currentTokenIs(tokens.EOF) {
		expr := pv.parseMediaQueryExpression()
		mediaQuery.Queries = append(mediaQuery.Queries, expr)

		if pv.currentTokenIs(tokens.COMMA) {
			pv.advance() // Consume comma
		} else {
			break
		}
	}

	return mediaQuery
}

func (pv *ParseVisitor) parseMediaQueryExpression() MediaQueryExpression {
	expr := MediaQueryExpression{
		Features: make([]MediaFeature, 0),
	}

	if pv.currentTokenIs(tokens.IDENT) {
		switch string(pv.currentToken.Literal) {
		case "not":
			expr.Not = true
			pv.advance()
		case "only":
			expr.Only = true
			pv.advance()
		}
	}

	if pv.currentTokenIs(tokens.IDENT) {
		expr.MediaType = pv.currentToken.Literal
		pv.advance()
	}

	for pv.currentTokenIs(tokens.LPAREN) || (pv.currentTokenIs(tokens.IDENT) && string(pv.currentToken.Literal) == "and") {
		if pv.currentTokenIs(tokens.IDENT) && string(pv.currentToken.Literal) == "and" {
			pv.advance() // Consume 'and'
		}
		feature := pv.parseMediaFeature()
		if feature != nil {
			expr.Features = append(expr.Features, *feature)
		}
	}

	return expr
}

func (pv *ParseVisitor) parseMediaFeature() *MediaFeature {
	if !pv.consume(tokens.LPAREN, "Expected '(' for media feature") {
		return nil
	}

	feature := &MediaFeature{}

	if !pv.currentTokenIs(tokens.IDENT) {
		pv.addError("Expected media feature name", pv.currentToken)
		pv.skipToNextRule()
		return nil
	}

	feature.Name = pv.currentToken.Literal
	pv.advance()

	if pv.currentTokenIs(tokens.COLON) {
		pv.advance() // Consume ':'
		if pv.currentTokenIs(tokens.IDENT) || pv.currentTokenIs(tokens.NUMBER) {
			feature.Value = pv.currentToken.Literal
			pv.advance()

			// Handle units like 'px'
			if pv.currentTokenIs(tokens.IDENT) || pv.currentTokenIs(tokens.PERCENTAGE) {
				feature.Value = append(feature.Value, pv.currentToken.Literal...)
				pv.advance()
			}
		} else {
			pv.addError("Expected value for media feature", pv.currentToken)
		}
	}

	if !pv.consume(tokens.RPAREN, "Expected ')' to close media feature") {
		return nil
	}

	return feature
}

func (pv *ParseVisitor) parseAttributeSelector() *SelectorValue {
	var attrBuilder strings.Builder
	attrBuilder.WriteByte('[')

	pv.advance() // Consume '['
	for !pv.currentTokenIs(tokens.RBRACKET) && !pv.currentTokenIs(tokens.EOF) {
		attrBuilder.Write(pv.currentToken.Literal)
		pv.advance()
	}

	if pv.currentTokenIs(tokens.RBRACKET) {
		attrBuilder.WriteByte(']')
		pv.advance() // Consume ']'
		return &SelectorValue{
			Type:  Attribute,
			Value: []byte(attrBuilder.String()),
		}
	} else {
		pv.addError("Expected closing bracket for attribute selector", pv.currentToken)
		return nil
	}
}

func (pv *ParseVisitor) parsePseudoSelector() *SelectorValue {
	pseudo := pv.currentToken.Literal
	pv.advance() // Consume the colon(s)
	if pv.currentTokenIs(tokens.IDENT) {
		pseudo = append(pseudo, pv.currentToken.Literal...)
		pv.advance()
		return &SelectorValue{
			Type:  Pseudo,
			Value: pseudo,
		}
	} else {
		pv.addError("Expected identifier after pseudo-selector", pv.currentToken)
		return nil
	}
}

func (pv *ParseVisitor) parseKeyframesRule() *KeyframesRule {
	keyframesRule := &KeyframesRule{
		Stops: make([]KeyframeStop, 0),
	}

	if !pv.currentTokenIs(tokens.IDENT) {
		pv.addError("Expected keyframes name", pv.currentToken)
		return nil
	}

	keyframesRule.Name = pv.currentToken.Literal
	pv.advance()

	if !pv.consume(tokens.LBRACE, "Expected '{' after @keyframes name") {
		return nil
	}

	for !pv.currentTokenIs(tokens.RBRACE) && !pv.currentTokenIs(tokens.EOF) {
		stop := pv.parseKeyframeStop()
		if stop != nil {
			keyframesRule.Stops = append(keyframesRule.Stops, *stop)
		}
	}

	pv.consume(tokens.RBRACE, "Expected '}' to close @keyframes block")

	return keyframesRule
}

func (pv *ParseVisitor) parseKeyframeStop() *KeyframeStop {
	stop := &KeyframeStop{
		Selectors: make([][]byte, 0),
		Rules:     make([]Node, 0),
	}

	// Parse selectors
	for !pv.currentTokenIs(tokens.LBRACE) && !pv.currentTokenIs(tokens.EOF) {
		switch pv.currentToken.Type {
		case tokens.PERCENTAGE, tokens.IDENT, tokens.NUMBER:
			stop.Selectors = append(stop.Selectors, pv.currentToken.Literal)
			pv.advance()
		case tokens.COMMA:
			pv.advance()
		default:
			pv.addError("Unexpected token in keyframe selector", pv.currentToken)
			pv.advance()
		}
	}

	if !pv.consume(tokens.LBRACE, "Expected '{' after keyframe selector") {
		return nil
	}

	// Parse declarations
	for !pv.currentTokenIs(tokens.RBRACE) && !pv.currentTokenIs(tokens.EOF) {
		declaration := pv.parseDeclaration()
		if declaration != nil {
			stop.Rules = append(stop.Rules, declaration)
		}

		if pv.currentTokenIs(tokens.SEMICOLON) {
			pv.advance() // Consume ';'
		}
	}

	pv.consume(tokens.RBRACE, "Expected '}' at the end of keyframe block")

	return stop
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
}
