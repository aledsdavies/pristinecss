package parser

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/aledsdavies/pristinecss/pkg/tokens"
)

var (
	selectorPool = sync.Pool{
		New: func() interface{} { return &Selector{} },
	}
	selectorValuePool = sync.Pool{
		New: func() interface{} { return &SelectorValue{} },
	}
	declarationPool = sync.Pool{
		New: func() interface{} { return &Declaration{} },
	}
	parseErrorPool = sync.Pool{
		New: func() interface{} { return &ParseError{} },
	}
	byteSlicePool = sync.Pool{
		New: func() interface{} { return make([]byte, 0, 64) },
	}
)

func getByteSlice() []byte {
	return byteSlicePool.Get().([]byte)[:0]
}

func putByteSlice(b []byte) {
	byteSlicePool.Put(b)
}

func (p *Parser) getParseError() *ParseError {
	return parseErrorPool.Get().(*ParseError)
}

func (p *Parser) putParseError(err *ParseError) {
	parseErrorPool.Put(err)
}

func getSelector() *Selector {
	return selectorPool.Get().(*Selector)
}

func putSelector(s *Selector) {
	s.Selectors = s.Selectors[:0]
	s.Rules = s.Rules[:0]
	selectorPool.Put(s)
}

func getSelectorValue() *SelectorValue {
	return selectorValuePool.Get().(*SelectorValue)
}

func putSelectorValue(sv *SelectorValue) {
	sv.Value = sv.Value[:0]
	selectorValuePool.Put(sv)
}

func getDeclaration() *Declaration {
	return declarationPool.Get().(*Declaration)
}

func putDeclaration(d *Declaration) {
	d.Key = d.Key[:0]
	d.Value = d.Value[:0]
	declarationPool.Put(d)
}

type ParseError struct {
	Message  string
	Line     int
	Column   int
	TokenLit []byte
}

func (e ParseError) Error() string {
	return fmt.Sprintf("line %d, column %d: %s (tokens. %s)", e.Line, e.Column, e.Message, e.TokenLit)
}

type Parser struct {
	tokens            []tokens.Token
	position          int
	currentToken      tokens.Token
	nextToken         tokens.Token
	stylesheet        *Stylesheet
	errors            []ParseError
	selectorValues    []SelectorValue
	declarationValues [][]byte
	classNameBuilder  bytes.Buffer
}

func Parse(tokens []tokens.Token) (*Stylesheet, []ParseError) {
	p := &Parser{
		tokens:            tokens,
		position:          0,
		stylesheet:        NewStylesheet(),
		errors:            make([]ParseError, 0, 10),
		selectorValues:    make([]SelectorValue, 0, 10),
		declarationValues: make([][]byte, 0, 10),
	}

	p.advance() // Load the first token
	p.advance() // Load the second token (now in nextToken)

	p.parseStylesheet()

	// Create a copy of the errors slice to return
	errors := make([]ParseError, len(p.errors))
	copy(errors, p.errors)

	// Release the errors back to the pool
	for i := range p.errors {
		p.putParseError(&p.errors[i])
	}
	p.errors = p.errors[:0]

	return p.stylesheet, errors
}

func (p *Parser) advance() {
	p.currentToken = p.nextToken
	if p.position < len(p.tokens) {
		p.nextToken = p.tokens[p.position]
		p.position++
	} else {
		p.nextToken = tokens.Token{Type: tokens.EOF}
	}
}

func (p *Parser) currentTokenIs(tokenType tokens.TokenType) bool {
	return p.currentToken.Type == tokenType
}

func (p *Parser) nextTokenIs(tokenType tokens.TokenType) bool {
	return p.nextToken.Type == tokenType
}

func (p *Parser) consume(tokenType tokens.TokenType, errorMessage string) bool {
	if p.currentTokenIs(tokenType) {
		p.advance()
		return true
	}
	p.addError(errorMessage, p.currentToken)
	return false
}

func (p *Parser) addError(message string, token tokens.Token) {
	err := p.getParseError()
	err.Message = message
	err.Line = token.Line
	err.Column = token.Column
	err.TokenLit = append(err.TokenLit[:0], token.Literal...)
	p.errors = append(p.errors, *err)
	p.putParseError(err)
}

func (p *Parser) parseStylesheet() {
	for !p.currentTokenIs(tokens.EOF) {
		switch p.currentToken.Type {
		case tokens.COMMENT:
			p.parseComment()
		case tokens.DOT, tokens.HASH, tokens.COLON, tokens.DBLCOLON, tokens.IDENT, tokens.LBRACKET:
			p.parseRule()
		default:
			p.addError("Unexpected token at stylesheet level", p.currentToken)
			p.advance()
		}
	}
}

func (p *Parser) parseComment() {
	comment := &Comment{Text: p.currentToken.Literal}
	p.stylesheet.Rules = append(p.stylesheet.Rules, comment)
	p.advance()
}

func (p *Parser) parseRule() {
	selector := p.parseSelector()
	if len(selector.Selectors) == 0 {
		p.skipToNextRule()
		return
	}

	p.parseDeclarationBlock(selector)
	p.stylesheet.Rules = append(p.stylesheet.Rules, selector)
}

func (p *Parser) parseSelector() *Selector {
	selector := &Selector{
		Selectors: make([]SelectorValue, 0),
		Rules:     make([]Node, 0),
	}

	for !p.currentTokenIs(tokens.EOF) && !p.currentTokenIs(tokens.LBRACE) {
		switch p.currentToken.Type {
		case tokens.IDENT:
			selector.Selectors = append(selector.Selectors, SelectorValue{
				Type:  Element,
				Value: p.currentToken.Literal,
			})
			p.advance()
		case tokens.DOT:
			if p.nextTokenIs(tokens.IDENT) {
				p.advance() // Consume the dot
				selector.Selectors = append(selector.Selectors, SelectorValue{
					Type:  Class,
					Value: append([]byte("."), p.currentToken.Literal...),
				})
				p.advance() // Consume the identifier
			} else {
				p.addError("Expected identifier after '.'", p.nextToken)
				p.advance() // Skip the dot
			}
		case tokens.HASH:
			if p.nextTokenIs(tokens.IDENT) {
				p.advance() // Consume the hash
				selector.Selectors = append(selector.Selectors, SelectorValue{
					Type:  ID,
					Value: append([]byte("#"), p.currentToken.Literal...),
				})
				p.advance() // Consume the identifier
			} else {
				p.addError("Expected identifier after '#'", p.nextToken)
				p.advance() // Skip the hash
			}
		case tokens.LBRACKET:
			attrSelector := p.parseAttributeSelector()
			if attrSelector != nil {
				selector.Selectors = append(selector.Selectors, *attrSelector)
			}
		case tokens.COLON, tokens.DBLCOLON:
			pseudoSelector := p.parsePseudoSelector()
			if pseudoSelector != nil {
				selector.Selectors = append(selector.Selectors, *pseudoSelector)
			}
		case tokens.COMMA, tokens.GREATER, tokens.PLUS, tokens.TILDE:
			selector.Selectors = append(selector.Selectors, SelectorValue{
				Type:  Combinator,
				Value: p.currentToken.Literal,
			})
			p.advance()
		default:
			p.addError("Unexpected token in selector", p.currentToken)
			p.advance() // Skip unexpected token
		}
	}

	return selector
}

func (p *Parser) parseAttributeSelector() *SelectorValue {
	var attrBuilder bytes.Buffer
	attrBuilder.WriteByte('[')

	p.advance() // Consume '['
	for !p.currentTokenIs(tokens.RBRACKET) && !p.currentTokenIs(tokens.EOF) {
		attrBuilder.Write(p.currentToken.Literal)
		p.advance()
	}

	if p.currentTokenIs(tokens.RBRACKET) {
		attrBuilder.WriteByte(']')
		p.advance() // Consume ']'
		return &SelectorValue{
			Type:  Attribute,
			Value: attrBuilder.Bytes(),
		}
	} else {
		p.addError("Expected closing bracket for attribute selector", p.currentToken)
		return nil
	}
}

func (p *Parser) parsePseudoSelector() *SelectorValue {
	pseudo := p.currentToken.Literal
	p.advance() // Consume the colon(s)
	if p.currentTokenIs(tokens.IDENT) {
		pseudo = append(pseudo, p.currentToken.Literal...)
		p.advance()
		return &SelectorValue{
			Type:  Pseudo,
			Value: pseudo,
		}
	} else {
		p.addError("Expected identifier after pseudo-selector", p.currentToken)
		return nil
	}
}

func (p *Parser) parseDeclarationBlock(selector *Selector) {
	if !p.consume(tokens.LBRACE, "Expected '{' at the start of declaration block") {
		return
	}

	for !p.currentTokenIs(tokens.RBRACE) && !p.currentTokenIs(tokens.EOF) {
		declaration := p.parseDeclaration()
		if declaration != nil {
			selector.Rules = append(selector.Rules, declaration)
		}

		if p.currentTokenIs(tokens.SEMICOLON) {
			p.advance() // Consume ';'
		}
	}

	if !p.consume(tokens.RBRACE, "Expected '}' at the end of declaration block") {
		p.skipToNextRule()
	}
}

func (p *Parser) parseDeclaration() *Declaration {
	if !p.currentTokenIs(tokens.IDENT) {
		p.addError("Expected property name", p.currentToken)
		p.skipToNextSemicolonOrBrace()
		return nil
	}

	declaration := &Declaration{
		Key:   p.currentToken.Literal,
		Value: make([][]byte, 0),
	}
	p.advance() // Consume property name

	if !p.consume(tokens.COLON, "Expected ':' after property name") {
		p.skipToNextSemicolonOrBrace()
		return nil
	}

	// Parse declaration value
	for !p.currentTokenIs(tokens.SEMICOLON) && !p.currentTokenIs(tokens.RBRACE) && !p.currentTokenIs(tokens.EOF) {
		switch p.currentToken.Type {
		case tokens.IDENT, tokens.NUMBER, tokens.PERCENTAGE, tokens.STRING, tokens.HASH:
			declaration.Value = append(declaration.Value, p.currentToken.Literal)
		case tokens.LPAREN:
			value := p.parseFunction()
			if value != nil {
				declaration.Value = append(declaration.Value, value...)
			}
		default:
			p.addError("Unexpected tokens.in declaration value", p.currentToken)
			p.skipToNextSemicolonOrBrace()
			return nil
		}
		p.advance()
	}

	return declaration
}

func (p *Parser) skipToNextSemicolonOrBrace() {
	for !p.currentTokenIs(tokens.SEMICOLON) && !p.currentTokenIs(tokens.RBRACE) && !p.currentTokenIs(tokens.EOF) {
		p.advance()
	}
}

func (p *Parser) parseFunction() [][]byte {
	var function [][]byte
	function = append(function, p.currentToken.Literal) // '('
	p.advance()

	for !p.currentTokenIs(tokens.RPAREN) && !p.currentTokenIs(tokens.EOF) {
		switch p.currentToken.Type {
		case tokens.IDENT, tokens.NUMBER, tokens.PERCENTAGE, tokens.STRING, tokens.HASH, tokens.COMMA:
			function = append(function, p.currentToken.Literal)
		default:
			p.addError("Unexpected tokens.in function", p.currentToken)
			return function
		}
		p.advance()
	}

	if p.consume(tokens.RPAREN, "Expected ')' to close function") {
		function = append(function, p.currentToken.Literal)
	}

	return function
}

func (p *Parser) skipToNextRule() {
	for !p.currentTokenIs(tokens.EOF) && !p.currentTokenIs(tokens.RBRACE) {
		p.advance()
	}
	if p.currentTokenIs(tokens.RBRACE) {
		p.advance() // Consume the '}'
	}
}
