package parser

import (
	"fmt"
	"strings"

	"github.com/aledsdavies/pristinecss/pkg/tokens"
)

var _ Node = (*MediaAtRule)(nil)

type MediaAtRule struct {
	Name  []byte
	Query MediaQuery
	Rules []Node
}

func (m *MediaAtRule) Type() NodeType { return NodeAtRule }

func (m *MediaAtRule) AtType() AtType { return AtRuleMedia }

func (m *MediaAtRule) Accept(v Visitor) { v.VisitMediaAtRule(m) }

func (m *MediaAtRule) String() string {
	var sb strings.Builder
	sb.WriteString("MediaAtRule{\n")
	sb.WriteString(fmt.Sprintf("  Name: %q,\n", m.Name))
	sb.WriteString("  Query: ")
	sb.WriteString(indentLines(m.Query.String(), 2))
	sb.WriteString(",\n")
	if len(m.Rules) > 0 {
		sb.WriteString("  Rules: [\n")
		for _, rule := range m.Rules {
			sb.WriteString(indentLines(rule.String(), 4))
			sb.WriteString(",\n")
		}
		sb.WriteString("  ]\n")
	}
	sb.WriteString("}")
	return sb.String()
}

type MediaQuery struct {
	Queries []MediaQueryExpression
}

func (mq MediaQuery) String() string {
	var sb strings.Builder
	sb.WriteString("MediaQuery{\n")
	sb.WriteString("  Queries: [\n")
	for _, query := range mq.Queries {
		sb.WriteString(indentLines(query.String(), 4))
		sb.WriteString(",\n")
	}
	sb.WriteString("  ]\n")
	sb.WriteString("}")
	return sb.String()
}

type MediaQueryExpression struct {
	MediaType []byte
	Not       bool
	Only      bool
	Features  []MediaFeature
}

func (mqe MediaQueryExpression) String() string {
	var sb strings.Builder
	sb.WriteString("MediaQueryExpression{\n")
	sb.WriteString(fmt.Sprintf("  MediaType: %q,\n", mqe.MediaType))
	sb.WriteString(fmt.Sprintf("  Not: %v,\n", mqe.Not))
	sb.WriteString(fmt.Sprintf("  Only: %v,\n", mqe.Only))
	sb.WriteString("  Features: [\n")
	for _, feature := range mqe.Features {
		sb.WriteString(indentLines(feature.String(), 4) + ",\n")
	}
	sb.WriteString("  ]\n")
	sb.WriteString("}")
	return sb.String()
}

type MediaFeature struct {
	Name  []byte
	Value []byte
}

func (mf MediaFeature) String() string {
	var valueStr string
	if mf.Value != nil {
		valueStr = fmt.Sprintf("%q", mf.Value)
	} else {
		valueStr = "nil"
	}
	return fmt.Sprintf("MediaFeature{Name: %q, Value: %s}", mf.Name, valueStr)
}

func (pv *ParseVisitor) VisitMediaAtRule(m *MediaAtRule) {
	pv.advance() // Consume 'media'
	m.Query = *pv.parseMediaQuery()

	if !pv.consume(tokens.LBRACE, "Expected '{' after media query") {
		return
	}

	// Parse the content of the media block
	for !pv.currentTokenIs(tokens.RBRACE) && !pv.currentTokenIs(tokens.EOF) {
		switch pv.currentToken.Type {
		case tokens.COMMENT:
			comment := &Comment{Text: pv.currentToken.Literal}
			comment.Accept(pv)
			m.Rules = append(m.Rules, comment)
		case tokens.DOT, tokens.HASH, tokens.COLON, tokens.DBLCOLON, tokens.IDENT, tokens.LBRACKET:
			selector := &Selector{
				Selectors: make([]SelectorValue, 0),
				Rules:     make([]Node, 0),
			}
			selector.Accept(pv)
			m.Rules = append(m.Rules, selector)
		default:
			pv.addError("Unexpected token in media block", pv.currentToken)
			pv.advance()
		}
	}

	pv.consume(tokens.RBRACE, "Expected '}' to close media block")
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
