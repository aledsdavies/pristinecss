package parser

import (
	"fmt"
	"strings"

	"github.com/aledsdavies/pristinecss/pkg/tokens"
)

var _ Node = (*Declaration)(nil)

type Declaration struct {
	Key   []byte
	Value [][]byte
}

func (c *Declaration) Type() NodeType   { return NodeDeclaration }
func (d *Declaration) Accept(v Visitor) { v.VisitDeclaration(d) }
func (d *Declaration) String() string {
	var sb strings.Builder
	sb.WriteString("Declaration{\n")
	sb.WriteString(fmt.Sprintf("  Key: %q,\n", d.Key))
	sb.WriteString("  Value: [\n")
	for _, v := range d.Value {
		sb.WriteString(fmt.Sprintf("    %q,\n", v))
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
		case tokens.IDENT, tokens.STRING, tokens.HASH:
			d.Value = append(d.Value, pv.currentToken.Literal)
		case tokens.NUMBER:
			number := pv.currentToken.Literal
			if pv.nextTokenIs(tokens.PERCENTAGE) || isUnit(pv.nextToken.Literal) {
				pv.advance()
				number = append(number, pv.currentToken.Literal...)
			}
			d.Value = append(d.Value, number)
		case tokens.LPAREN:
			value := pv.parseFunction()
			if value != nil {
				d.Value = append(d.Value, value...)
			}
		default:
			pv.addError("Unexpected token in declaration value", pv.currentToken)
			pv.skipToNextSemicolonOrBrace()
			return
		}
		pv.advance()
	}
}

func (pv *ParseVisitor) parseFunction() [][]byte {
	var function [][]byte
	function = append(function, pv.currentToken.Literal) // '('
	pv.advance()

	for !pv.currentTokenIs(tokens.RPAREN) && !pv.currentTokenIs(tokens.EOF) {
		switch pv.currentToken.Type {
		case tokens.IDENT, tokens.STRING, tokens.HASH, tokens.COMMA:
			function = append(function, pv.currentToken.Literal)
		case tokens.NUMBER:
			number := pv.currentToken.Literal
			if pv.nextTokenIs(tokens.PERCENTAGE) || isUnit(pv.nextToken.Literal) {
				pv.advance()
				number = append(number, pv.currentToken.Literal...)
			}

			function = append(function, number)
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
