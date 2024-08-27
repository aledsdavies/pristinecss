package parser

import (
	"fmt"
	"github.com/aledsdavies/pristinecss/pkg/tokens"
)

const (
    NodeAtRule NodeType = "AT_RULE"
)

func init() {
	RegisterNodeType(NodeAtRule, visitAt)
}


type AtRule interface {
	Node
	AtType() AtType
}

type AtHandler func(*ParseVisitor, AtRule)
type AtInit func() AtRule

func (pv *ParseVisitor) getAtRule() Node {
	pv.advance() // Consume the @
	if !pv.currentTokenIs(tokens.IDENT) {
		pv.addError("Expected identifier after @", pv.currentToken)
		return nil
	}

	keyword := string(pv.currentToken.Literal)
	initFn, exists := keywordRegistry[keyword]
	if !exists {
		pv.addError(fmt.Sprintf("Unknown at-rule: @%s", keyword), pv.currentToken)
		return nil
	}

	return initFn()
}

func visitAt(pv *ParseVisitor, node Node) {
	atRule, ok := node.(AtRule)
	if !ok {
		pv.addError(fmt.Sprintf("Expected AtRule, got %T", node), pv.currentToken)
		return
	}

	handler := GetAtHandler(atRule)
	handler(pv, atRule)
}
