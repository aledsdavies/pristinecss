package parser

import (
	"fmt"
	"strings"

	"github.com/aledsdavies/pristinecss/pkg/tokens"
)

var _ Node = (*Selector)(nil)

type Selector struct {
	Selectors []SelectorValue

	// Comment
	// At Rules
	// Selectors (nested)
	// Declerations
	Rules []Node
}

func (c *Selector) Type() NodeType   { return NodeSelector }
func (s *Selector) Accept(v Visitor) { v.VisitSelector(s) }
func (s *Selector) String() string {
	var sb strings.Builder
	sb.WriteString("Selector{\n")
	sb.WriteString("  Selectors: [\n")
	for _, sel := range s.Selectors {
		sb.WriteString("    " + sel.String() + ",\n")
	}
	sb.WriteString("  ]\n")
	if len(s.Rules) > 0 {
		sb.WriteString("  Rules: [\n")
		for _, rule := range s.Rules {
			sb.WriteString(indentLines(rule.String(), 4) + "\n")
		}
		sb.WriteString("  ]\n")
	} else {
		sb.WriteString("  Rules: []\n")
	}
	sb.WriteString("}")
	return sb.String()
}

type SelectorType int

const (
	Element SelectorType = iota
	Class
	ID
	Attribute
	Pseudo
	Combinator
)

type SelectorValue struct {
	Type  SelectorType
	Value []byte
}

func (sv SelectorValue) String() string {
	return fmt.Sprintf("{Type: %s, Value: %q}", selectorTypeToString(sv.Type), sv.Value)
}

func (pv *ParseVisitor) VisitSelector(s *Selector) {
	pv.parseSelector(s)
	if !pv.consume(tokens.LBRACE, "Expected '{' after selector") {
		return
	}
	for !pv.currentTokenIs(tokens.RBRACE) && !pv.currentTokenIs(tokens.EOF) {
		if !pv.currentTokenIs(tokens.IDENT) {
			pv.addError("Expected property name", pv.currentToken)
			pv.skipToNextSemicolonOrBrace()
			continue
		}
		declaration := &Declaration{
			Key:   pv.currentToken.Literal,
			Value: make([][]byte, 0),
		}
		declaration.Accept(pv)
		s.Rules = append(s.Rules, declaration)

		if pv.currentTokenIs(tokens.SEMICOLON) {
			pv.advance() // Consume ';'
		}
	}
	pv.consume(tokens.RBRACE, "Expected '}' at the end of declaration block")
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