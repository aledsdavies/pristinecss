package parser

import (
	"fmt"
	"strings"

	"github.com/aledsdavies/pristinecss/pkg/tokens"
)

var _ Node = (*Declaration)(nil)

type Declaration struct {
	Key   []byte
	Value []Value
}

func (d *Declaration) Type() NodeType   { return NodeDeclaration }
func (d *Declaration) Accept(v Visitor) { v.VisitDeclaration(d) }
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
    sb.WriteString("}")
    return sb.String()
}

func (pv *ParseVisitor) VisitDeclaration(d *Declaration) {
    pv.advance() // Consume property name
    if !pv.consume(tokens.COLON, "Expected ':' after property name") {
        pv.skipToNextSemicolonOrBrace()
        return
    }

    for !pv.currentTokenIs(tokens.SEMICOLON) && !pv.currentTokenIs(tokens.RBRACE) && !pv.currentTokenIs(tokens.EOF) {
        switch pv.currentToken.Type {
        case tokens.IDENT, tokens.HASH, tokens.URI, tokens.STRING, tokens.NUMBER:
            d.Value = append(d.Value, pv.parseValue())
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
