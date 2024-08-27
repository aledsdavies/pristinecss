package parser

import (
	"fmt"
	"log/slog"
	"runtime"
	"strings"

	"github.com/aledsdavies/pristinecss/pkg/tokens"
)

const(
    Import AtType = "import"
)

func init() {
	RegisterAt(Import, visitImportAtRule, func() AtRule { return &ImportAtRule{} })
}

type ImportAtRule struct {
	URL      Value
	Layer    Value
	Media    MediaQuery
	Supports SupportsCondition
}

type SupportsCondition interface {
	supportCondition()
}

type SupportsDecleration struct {
	Key   []byte
	Value []Value
}

func (SupportsDecleration) supportCondition() {}

type SupportsFunction struct {
	Name []byte
	Args []byte
}

func (SupportsFunction) supportCondition() {}

type SupportsOperator struct {
	Operator string
}

func (SupportsOperator) supportCondition() {}

type SupportsNot struct {
	Condition SupportsCondition
}

func (SupportsNot) supportCondition() {}

type SupportsGroup struct {
	Conditions []SupportsCondition
}

func (SupportsGroup) supportCondition() {}

func (r *ImportAtRule) Type() NodeType   { return NodeAtRule }
func (r *ImportAtRule) AtType() AtType   { return Import }
func (r *ImportAtRule) String() string {
	var sb strings.Builder
	sb.WriteString("ImportAtRule{\n")
	sb.WriteString(fmt.Sprintf("  URL: %s,\n", r.URL.String()))
	if r.Layer != nil {
		sb.WriteString(fmt.Sprintf("  Layer: %s,\n", r.Layer.String()))
	}
	if len(r.Media.Queries) > 0 {
		sb.WriteString(fmt.Sprintf("  Media: %s,\n", r.Media.String()))
	}
	if r.Supports != nil {
		sb.WriteString("  Supports: ")
		sb.WriteString(supportConditionToString(r.Supports, 1) + ",\n")
	}
	sb.WriteString("}")
	return sb.String()
}

func supportConditionToString(condition SupportsCondition, indentLevel int) string {
	indent := strings.Repeat("  ", indentLevel)

	switch c := condition.(type) {
	case *SupportsDecleration:
		return fmt.Sprintf("%sSupportDecleration{Key: %q, Value: %q}", indent, c.Key, c.Value)
	case *SupportsFunction:
		return fmt.Sprintf("%sSupportsFunction{Name: %q, Args: %q}", indent, c.Name, c.Args)
	case *SupportsOperator:
		return fmt.Sprintf("%sSupportsOperator{Operator: %q}", indent, c.Operator)
	case *SupportsNot:
		return fmt.Sprintf("%sSupportsNot{\n%sCondition: %s\n%s}",
			indent,
			strings.Repeat("  ", indentLevel+1),
			supportConditionToString(c.Condition, 0),
			indent)
	case *SupportsGroup:
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("%sSupportsGroup{\n", indent))
		sb.WriteString(fmt.Sprintf("%sConditions: [\n", strings.Repeat("  ", indentLevel+1)))
		for _, cond := range c.Conditions {
			sb.WriteString(supportConditionToString(cond, indentLevel+2) + ",\n")
		}
		sb.WriteString(fmt.Sprintf("%s]\n", strings.Repeat("  ", indentLevel+1)))
		sb.WriteString(fmt.Sprintf("%s}", indent))
		return sb.String()
	default:
		// Log the error with detailed information
		_, file, line, _ := runtime.Caller(1)
		slog.Error("Unexpected SupportsCondition type",
			"type", fmt.Sprintf("%T", c),
			"value", fmt.Sprintf("%+v", c),
			"file", file,
			"line", line,
		)
		// Panic with an assertion message
		panic(fmt.Sprintf("Assertion failed: unexpected SupportsCondition type %T", c))
	}
}

func visitImportAtRule(pv *ParseVisitor, node AtRule) {
    i := node.(*ImportAtRule)
	pv.advance() // Consume 'import'
	if pv.currentTokenIs(tokens.URI) || pv.currentTokenIs(tokens.STRING) {
		i.URL = pv.parseValue()
	} else {
		pv.addError("Expected string or URI after @import", pv.currentToken)
		return
	}

	for !pv.currentTokenIs(tokens.SEMICOLON) && !pv.currentTokenIs(tokens.EOF) {
		currTok := string(pv.currentToken.Literal)
		switch {
		case pv.currentTokenIs(tokens.IDENT) && currTok == "supports":
            pv.advance() // consume supports
			i.Supports = pv.parseSupportsCondition()
		case pv.currentTokenIs(tokens.IDENT) && currTok == "layer":
			i.Layer = pv.parseValue()
		case pv.currentTokenIs(tokens.IDENT) || pv.currentTokenIs(tokens.LPAREN):
			i.Media = *pv.parseMediaQuery()
		default:
			pv.addError("Unexpected token in @import rule", pv.currentToken)
			return
		}
	}

	pv.consume(tokens.SEMICOLON, "Expected ';' after @import rule")
}

func (pv *ParseVisitor) parseSupportsCondition() SupportsCondition {
	if pv.currentTokenIs(tokens.LPAREN) {
		pv.advance() // Consume '('

		if string(pv.currentToken.Literal) == "not" {
			pv.advance() // Consume 'not'
			notCondition := pv.parseSupportsCondition()
			pv.consume(tokens.RPAREN, "Expected ')' to close not condition")
			return &SupportsNot{Condition: notCondition}
		} else if pv.currentTokenIs(tokens.LPAREN) {
			// Handle nested conditions as a SupportsGroup
			group := &SupportsGroup{}
			for !pv.currentTokenIs(tokens.RPAREN) && !pv.currentTokenIs(tokens.EOF) {
				condition := pv.parseSupportsCondition()
				if condition != nil {
					group.Conditions = append(group.Conditions, condition)
				}
				if pv.currentTokenIs(tokens.IDENT) {
					op := string(pv.currentToken.Literal)
					if op == "and" || op == "or" {
						group.Conditions = append(group.Conditions, &SupportsOperator{Operator: op})
						pv.advance()
					}
				}
			}
			pv.consume(tokens.RPAREN, "Expected ')' to close group condition")
			return group
		} else if pv.currentTokenIs(tokens.IDENT) && pv.nextTokenIs(tokens.LPAREN) {
			condition := pv.parseSupportsFunction()
			pv.consume(tokens.RPAREN, "Expected ')' to close function condition")
			return condition
		} else {
			condition := pv.parseSupportsDeclaration()
			pv.consume(tokens.RPAREN, "Expected ')' to close declaration condition")
			return condition
		}
	}

	pv.addError("Unexpected token in supports condition", pv.currentToken)
	pv.skipToNextSemicolonOrBrace()
	return nil
}

func (pv *ParseVisitor) parseSupportsFunction() SupportsCondition {
	name := pv.currentToken.Literal
	pv.advance() // Consume function name

	if !pv.currentTokenIs(tokens.LPAREN) {
		return nil
	}

	var args []byte
	parenCount := 1 // We've already consumed one '('
	pv.advance()

	for parenCount > 0 && !pv.currentTokenIs(tokens.EOF) {
		if pv.currentTokenIs(tokens.LPAREN) {
			parenCount++
		} else if pv.currentTokenIs(tokens.RPAREN) {
			parenCount--
		}

		if parenCount > 0 {
			args = append(args, pv.currentToken.Literal...)
		}
		pv.advance()
	}

	return &SupportsFunction{Name: name, Args: args}
}

func (pv *ParseVisitor) parseSupportsDeclaration() SupportsCondition {
	declaration := &SupportsDecleration{
		Key: pv.currentToken.Literal,
	}
	pv.advance() // consume the property
	if !pv.currentTokenIs(tokens.COLON) {
		pv.addError("Unexpected token in supports declaration", pv.currentToken)
		return nil
	}
	pv.advance()
	declaration.Value = append(declaration.Value, pv.parseValue())

	return declaration
}
