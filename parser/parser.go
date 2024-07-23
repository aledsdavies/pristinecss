package parser

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"sync"
)

type ParseError struct {
	Message  string
	Line     int
	Column   int
	TokenLit []byte
}

func (e ParseError) Error() string {
	return fmt.Sprintf("line %d, column %d: %s (token: %s)", e.Line, e.Column, e.Message, e.TokenLit)
}

type Parser struct {
	lexer             *lexer
	tokens            []Token
	position          int
	bufferSize        int
	stylesheet        *Stylesheet
	errors            []ParseError
	selectorValues    []SelectorValue
	declarationValues [][]byte
	classNameBuilder  bytes.Buffer
}

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
)

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
	selectorPool.Put(s)
}

func getSelectorValue() *SelectorValue {
	return selectorValuePool.Get().(*SelectorValue)
}

func putSelectorValue(sv *SelectorValue) {
	selectorValuePool.Put(sv)
}

func getDeclaration() *Declaration {
	return declarationPool.Get().(*Declaration)
}

func putDeclaration(d *Declaration) {
	declarationPool.Put(d)
}

func Parse(input io.Reader) (*Stylesheet, []ParseError) {
	lexer := Read(input)
	var buffSize = 20
	p := &Parser{
		lexer:             lexer,
		tokens:            make([]Token, buffSize),
		position:          0,
		bufferSize:        buffSize,
		stylesheet:        NewStylesheet(),
		errors:            make([]ParseError, 0, 10),
		selectorValues:    make([]SelectorValue, 0, 10),
		declarationValues: make([][]byte, 0, 10),
	}

	for i := 0; i < buffSize; i++ {
		p.tokens[i] = p.lexer.NextToken()
	}

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

func (p *Parser) currentToken() Token {
	return p.tokens[p.position]
}

func (p *Parser) peekToken(offset int) Token {
	return p.tokens[(p.position+offset)%p.bufferSize]
}

func (p *Parser) advance() {
	p.position++
	if p.position >= len(p.tokens) {
		newToken := p.lexer.NextToken()
		p.tokens = append(p.tokens, newToken)
	}
}

func (p *Parser) isAtEnd() bool {
	return p.currentToken().Type == EOF
}

func (p *Parser) addError(message string, token Token) {
	err := p.getParseError()
	err.Message = message
	err.Line = token.Line
	err.Column = token.Column
	err.TokenLit = append(err.TokenLit[:0], token.Literal...)
	p.errors = append(p.errors, *err)
	p.putParseError(err)
}

func (p *Parser) parseStylesheet() {
	var lastToken Token
	var repeatCount int
	for !p.isAtEnd() {
		currentToken := p.currentToken()
		if bytes.Equal(currentToken.Literal, lastToken.Literal) {
			repeatCount++
			if repeatCount > 1000 {
				log.Fatalf("Potential infinite loop detected. Parser stuck on token: %v with value %s at position %d:%d", currentToken, currentToken.Literal, currentToken.Line, currentToken.Column)
			}
		} else {
			repeatCount = 0
		}
		lastToken = currentToken

		switch currentToken.Type {
		case COMMENT:
			comment := &Comment{Text: currentToken.Literal}
			p.stylesheet.Rules = append(p.stylesheet.Rules, comment)
			p.advance()
		case DOT, HASH, COLON, DBLCOLON, IDENT, LBRACKET:
			selector := p.parseSelector()
			if len(selector.Selectors) == 0 {
				p.skipToNextRule()
			} else {
				if p.currentToken().Type == LBRACE {
					p.parseDeclarationBlock(selector)
					p.stylesheet.Rules = append(p.stylesheet.Rules, selector)
				} else {
					p.addError("Expected '{' after selector", p.currentToken())
					p.skipToNextRule()
				}
			}
		default:
			p.addError("Unexpected token at stylesheet level", p.currentToken())
			p.skipToNextRule()
		}
	}
}

func (p *Parser) parseSelector() *Selector {
	selector := selectorPool.Get().(*Selector)
	selector.Selectors = selector.Selectors[:0]
	selector.Rules = selector.Rules[:0]
	p.selectorValues = p.selectorValues[:0]

	for !p.isAtEnd() {
		if p.currentToken().Type == LBRACE {
			break
		}

		switch p.currentToken().Type {
		case DOT:
			p.advance()
			if p.currentToken().Type != IDENT && p.currentToken().Type != NUMBER {
				p.addError("Expected identifier or number after '.'", p.currentToken())
				p.advance()
				continue
			}
			p.classNameBuilder.Reset()
			p.classNameBuilder.WriteByte('.')
			p.classNameBuilder.Write(p.currentToken().Literal)
			p.advance()
			for p.currentToken().Type == MINUS || p.currentToken().Type == NUMBER {
				p.classNameBuilder.Write(p.currentToken().Literal)
				p.advance()
			}
			sv := getSelectorValue()
			sv.Type = Class
			sv.Value = append([]byte(nil), p.classNameBuilder.Bytes()...)
			p.selectorValues = append(p.selectorValues, *sv)
			putSelectorValue(sv)

		case HASH:
			p.advance()
			if p.currentToken().Type != IDENT {
				p.addError("Expected identifier after '#'", p.currentToken())
				return selector
			}
			sv := getSelectorValue()
			sv.Type = ID
			sv.Value = append([]byte("#"), p.currentToken().Literal...)
			p.selectorValues = append(p.selectorValues, *sv)
			putSelectorValue(sv)
			p.advance()

		case IDENT:
			sv := getSelectorValue()
			sv.Type = Element
			sv.Value = append([]byte(nil), p.currentToken().Literal...)
			p.selectorValues = append(p.selectorValues, *sv)
			putSelectorValue(sv)
			p.advance()

		case LBRACKET:
			attrSelector := p.parseAttributeSelector()
			if attrSelector != nil {
				p.selectorValues = append(p.selectorValues, *attrSelector)
			}

		case COLON, DBLCOLON:
			literal := p.currentToken().Literal
			p.advance()
			if p.currentToken().Type != IDENT {
				p.addError(fmt.Sprintf("Expected identifier after '%s'", literal), p.currentToken())
				return selector
			}
			sv := getSelectorValue()
			sv.Type = Pseudo
			sv.Value = append(append([]byte(nil), literal...), p.currentToken().Literal...)
			p.selectorValues = append(p.selectorValues, *sv)
			putSelectorValue(sv)
			p.advance()

		case COMMA, GREATER, PLUS, TILDE:
			sv := getSelectorValue()
			sv.Type = Combinator
			sv.Value = append([]byte(nil), p.currentToken().Literal...)
			p.selectorValues = append(p.selectorValues, *sv)
			putSelectorValue(sv)
			p.advance()

		default:
			p.addError("Unexpected token in selector", p.currentToken())
			return selector
		}
	}

	selector.Selectors = append(selector.Selectors, p.selectorValues...)
	return selector
}

func (p *Parser) parseAttributeSelector() *SelectorValue {
	var attrBuilder bytes.Buffer
	attrBuilder.WriteByte('[')

	p.advance()
	for p.currentToken().Type != RBRACKET && !p.isAtEnd() {
		attrBuilder.Write(p.currentToken().Literal)
		p.advance()
	}

	if p.currentToken().Type == RBRACKET {
		attrBuilder.WriteByte(']')
		p.advance()
	} else {
		p.addError("Expected closing bracket for attribute selector", p.currentToken())
		return nil
	}

	sv := getSelectorValue()
	sv.Type = Attribute
	sv.Value = append([]byte(nil), attrBuilder.Bytes()...)
	return sv
}

func (p *Parser) parseDeclarationBlock(selector *Selector) {
	if p.currentToken().Type != LBRACE {
		p.addError("Expected '{' at the start of declaration block", p.currentToken())
		return
	}
	p.advance()

	for p.currentToken().Type != RBRACE && !p.isAtEnd() {
		declaration := p.parseDeclaration()
		if declaration != nil {
			selector.Rules = append(selector.Rules, declaration)
		} else {
			for !p.isAtEnd() && p.currentToken().Type != SEMICOLON && p.currentToken().Type != RBRACE {
				p.advance()
			}
			if p.currentToken().Type == SEMICOLON {
				p.advance()
			}
		}
	}

	if p.currentToken().Type != RBRACE {
		p.addError("Expected '}' at the end of declaration block", p.currentToken())
	} else {
		p.advance()
	}
}

func (p *Parser) parseDeclaration() *Declaration {
	if p.currentToken().Type != IDENT {
		p.addError("Expected property name", p.currentToken())
		return nil
	}

	declaration := declarationPool.Get().(*Declaration)
	declaration.Key = append([]byte(nil), p.currentToken().Literal...)
	p.advance()

	if p.currentToken().Type != COLON {
		p.addError("Expected ':' after property name", p.currentToken())
		putDeclaration(declaration)
		return nil
	}
	p.advance()

	p.declarationValues = p.declarationValues[:0]
	for p.currentToken().Type != SEMICOLON && p.currentToken().Type != RBRACE && !p.isAtEnd() {
		p.declarationValues = append(p.declarationValues, append([]byte(nil), p.currentToken().Literal...))
		p.advance()
	}

	if p.currentToken().Type == SEMICOLON {
		p.advance()
	}

	declaration.Value = append([][]byte(nil), p.declarationValues...)
	return declaration
}

func (p *Parser) skipToNextRule() {
	for !p.isAtEnd() {
		if p.currentToken().Type == RBRACE {
			p.advance()
			return
		}
		p.advance()
	}
}
