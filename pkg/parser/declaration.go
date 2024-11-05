package parser

import (
	"fmt"
	"github.com/aledsdavies/pristinecss/pkg/tokens"
	"strings"
)

const (
	NodeDeclaration NodeType = "Declaration"
)

func init() {
	RegisterNodeType(NodeDeclaration, visitDeclaration)
}

var _ Node = (*Declaration)(nil)

type Declaration struct {
	Key       []byte
	Value     []Value
	Important bool
}

func (d *Declaration) Type() NodeType { return NodeDeclaration }
func (d *Declaration) String() string {
	var sb strings.Builder
	sb.WriteString("Declaration{\n")
	sb.WriteString(fmt.Sprintf("  Key: %q,\n", d.Key))
	sb.WriteString("  Value: [\n")
	for _, vl := range d.Value {
		sb.WriteString(indentLines(vl.String(), 4))
		sb.WriteString(",\n")
	}
	sb.WriteString("  ]\n")
	sb.WriteString(fmt.Sprintf("  Important: %v\n", d.Important))
	sb.WriteString("}")
	return sb.String()
}

func visitDeclaration(pv *ParseVisitor, node Node) {
	d := node.(*Declaration)
	pv.advance() // Consume property name
	if !pv.consume(tokens.COLON, "Expected ':' after property name") {
		pv.skipToNextSemicolonOrBrace()
		return
	}
	for !pv.currentTokenIs(tokens.SEMICOLON) && !pv.currentTokenIs(tokens.RBRACE) && !pv.currentTokenIs(tokens.EOF) {
		switch pv.currentToken.Type {
		case tokens.COMMENT:
			comment := &Comment{Text: pv.currentToken.Literal}
			visitComment(pv, comment)
			d.Value = append(d.Value, comment)
		case tokens.IDENT, tokens.HASH, tokens.URI, tokens.STRING, tokens.NUMBER, tokens.COLOR:
			d.Value = append(d.Value, pv.parseValue())
		case tokens.EXCLAMATION:
			if pv.nextTokenIs(tokens.IDENT) && string(pv.nextToken.Literal) == "important" {
				d.Important = true
				pv.advance() // Consume '!'
				pv.advance() // Consume 'important'
			} else {
				pv.addError("Unexpected '!' in declaration value", pv.currentToken)
				pv.skipToNextSemicolonOrBrace()
				return
			}
		case tokens.COMMA:
			// Skip the comma and continue parsing values
			pv.advance()
		default:
			pv.addError("Unexpected token in declaration value", pv.currentToken)
			pv.skipToNextSemicolonOrBrace()
			return
		}
	}
	// Consume the semicolon if present
	if pv.currentTokenIs(tokens.SEMICOLON) {
		pv.advance()
	}
}
