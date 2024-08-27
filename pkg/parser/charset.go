package parser

import (
	"fmt"

	"github.com/aledsdavies/pristinecss/pkg/tokens"
)

const(
    Charset AtType = "charset"
)

func init() {
	RegisterAt(Charset, visitCharsetAtRule, func() AtRule { return &CharsetAtRule{} })
}

type CharsetAtRule struct {
	Charset Value
}

func (r *CharsetAtRule) Type() NodeType   { return NodeAtRule }
func (r *CharsetAtRule) AtType() AtType   { return Charset }
func (r *CharsetAtRule) String() string {
	return fmt.Sprintf("CharsetAtRule{Charset: %q}", r.Charset)
}

func visitCharsetAtRule(pv *ParseVisitor, node AtRule) {
    r := node.(*CharsetAtRule)
	pv.advance() // Consume 'charset'

	if pv.currentTokenIs(tokens.STRING) {
		r.Charset = pv.parseValue()
	} else {
		pv.addError("Expected string after @charset", pv.currentToken)
		return
	}

	pv.consume(tokens.SEMICOLON, "Expected ';' after @charset rule")
}

