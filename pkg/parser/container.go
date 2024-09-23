package parser

import (
	"fmt"
	"strings"

	"github.com/aledsdavies/pristinecss/pkg/tokens"
)

const (
	Container AtType = "container"
)

func init() {
	RegisterAt(Container, visitContainerAtRule, func() AtRule { return &ContainerAtRule{} })
}

type ContainerAtRule struct {
	Name        []byte
	Query       ContainerQuery
	Declarations []Node
}

type ContainerQuery struct {
	Conditions []ContainerCondition
}

type ContainerCondition struct {
	Features []ContainerFeature
}

type ContainerFeature struct {
	Name  []byte
	Value []byte
}

func (r *ContainerAtRule) Type() NodeType   { return NodeAtRule }
func (r *ContainerAtRule) AtType() AtType   { return Container }
func (r *ContainerAtRule) String() string {
	var sb strings.Builder
	sb.WriteString("ContainerAtRule{\n")
	if r.Name != nil {
		sb.WriteString(fmt.Sprintf("  Name: %q,\n", r.Name))
	}
	sb.WriteString("  Query: ")
	sb.WriteString(r.Query.String())
	sb.WriteString(",\n")
	sb.WriteString("  Declarations: [\n")
	for _, decl := range r.Declarations {
		sb.WriteString(indentLines(decl.String(), 4))
		sb.WriteString(",\n")
	}
	sb.WriteString("  ]\n")
	sb.WriteString("}")
	return sb.String()
}

func (cq ContainerQuery) String() string {
	var sb strings.Builder
	sb.WriteString("ContainerQuery{\n")
	sb.WriteString("    Conditions: [\n")
	for _, cond := range cq.Conditions {
		sb.WriteString(indentLines(cond.String(), 6))
		sb.WriteString(",\n")
	}
	sb.WriteString("    ]\n")
	sb.WriteString("  }")
	return sb.String()
}

func (cc ContainerCondition) String() string {
	var sb strings.Builder
	sb.WriteString("ContainerCondition{\n")
	sb.WriteString("      Features: [\n")
	for _, feat := range cc.Features {
		sb.WriteString(indentLines(feat.String(), 8))
		sb.WriteString(",\n")
	}
	sb.WriteString("      ]\n")
	sb.WriteString("    }")
	return sb.String()
}

func (cf ContainerFeature) String() string {
	return fmt.Sprintf("ContainerFeature{Name: %q, Value: %q}", cf.Name, cf.Value)
}

func visitContainerAtRule(pv *ParseVisitor, node AtRule) {
	c := node.(*ContainerAtRule)
	pv.advance() // Consume 'container'

	// Parse optional container name
	if pv.currentTokenIs(tokens.IDENT) {
		c.Name = pv.currentToken.Literal
		pv.advance()
	}

	// Parse container query
	c.Query = *parseContainerQuery(pv)

	if !pv.consume(tokens.LBRACE, "Expected '{' after @container") {
		return
	}

	// Parse declarations
	for !pv.currentTokenIs(tokens.RBRACE) && !pv.currentTokenIs(tokens.EOF) {
		switch pv.currentToken.Type {
		case tokens.IDENT:
			declaration := &Declaration{
				Key: pv.currentToken.Literal,
			}
			visitDeclaration(pv, declaration)
			c.Declarations = append(c.Declarations, declaration)
		case tokens.DOT, tokens.HASH, tokens.COLON, tokens.DBLCOLON:
			selector := &Selector{
				Selectors: make([]SelectorValue, 0),
				Rules:     make([]Node, 0),
			}
			visitSelector(pv, selector)
			c.Declarations = append(c.Declarations, selector)
		default:
			pv.addError("Unexpected token in @container rule", pv.currentToken)
			pv.advance()
		}

		if pv.currentTokenIs(tokens.SEMICOLON) {
			pv.advance()
		}
	}

	pv.consume(tokens.RBRACE, "Expected '}' to close @container rule")
}

func parseContainerQuery(pv *ParseVisitor) *ContainerQuery {
	query := &ContainerQuery{
		Conditions: make([]ContainerCondition, 0),
	}

	for !pv.currentTokenIs(tokens.LBRACE) && !pv.currentTokenIs(tokens.EOF) {
		condition := parseContainerCondition(pv)
		if condition != nil {
			query.Conditions = append(query.Conditions, *condition)
		}

		if pv.currentTokenIs(tokens.IDENT) && string(pv.currentToken.Literal) == "and" {
			pv.advance() // Consume 'and'
		} else {
			break
		}
	}

	return query
}

func parseContainerCondition(pv *ParseVisitor) *ContainerCondition {
	condition := &ContainerCondition{
		Features: make([]ContainerFeature, 0),
	}

	if !pv.consume(tokens.LPAREN, "Expected '(' for container condition") {
		return nil
	}

	for !pv.currentTokenIs(tokens.RPAREN) && !pv.currentTokenIs(tokens.EOF) {
		feature := parseContainerFeature(pv)
		if feature != nil {
			condition.Features = append(condition.Features, *feature)
		}

		if pv.currentTokenIs(tokens.IDENT) && string(pv.currentToken.Literal) == "and" {
			pv.advance() // Consume 'and'
		}
	}

	if !pv.consume(tokens.RPAREN, "Expected ')' to close container condition") {
		return nil
	}

	return condition
}

func parseContainerFeature(pv *ParseVisitor) *ContainerFeature {
	feature := &ContainerFeature{}

	if !pv.currentTokenIs(tokens.IDENT) {
		pv.addError("Expected identifier for container feature", pv.currentToken)
		return nil
	}

	feature.Name = pv.currentToken.Literal
	pv.advance()

	if !pv.consume(tokens.COLON, "Expected ':' after container feature name") {
		return nil
	}

	if pv.currentTokenIs(tokens.IDENT) || pv.currentTokenIs(tokens.NUMBER) {
		feature.Value = pv.currentToken.Literal
		pv.advance()

		// Handle units like 'px'
		if pv.currentTokenIs(tokens.IDENT) {
			feature.Value = append(feature.Value, pv.currentToken.Literal...)
			pv.advance()
		}
	} else {
		pv.addError("Expected value for container feature", pv.currentToken)
		return nil
	}

	return feature
}
