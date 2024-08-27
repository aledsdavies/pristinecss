package parser

import (
	"strings"

	"github.com/aledsdavies/pristinecss/pkg/tokens"
)


const (
	NodeStylesheet NodeType = "stylesheet"
)

func init() {
	RegisterNodeType(NodeStylesheet, visitStylesheet)
}

var _ Node = (*Stylesheet)(nil)

type Stylesheet struct {
	Rules []Node
}

func NewStylesheet() *Stylesheet {
	return &Stylesheet{Rules: []Node{}}
}

func (s *Stylesheet) Type() NodeType { return NodeStylesheet }
func (s *Stylesheet) String() string {
	var sb strings.Builder
	sb.WriteString("Stylesheet{\n")
	if s.Rules != nil {
		for _, rule := range s.Rules {
			if rule != nil {
				sb.WriteString(indentLines(rule.String(), 2))
				sb.WriteString(",\n")
			}
		}
	}
	sb.WriteString("}")
	return sb.String()
}

func visitStylesheet(pv *ParseVisitor, node Node) {
	s := node.(*Stylesheet)
	for !pv.currentTokenIs(tokens.EOF) {
		var childNode Node
		switch pv.currentToken.Type {
		case tokens.COMMENT:
			childNode = &Comment{Text: pv.currentToken.Literal}
			visitComment(pv, childNode)
		case tokens.DOT, tokens.HASH, tokens.COLON, tokens.DBLCOLON, tokens.IDENT, tokens.LBRACKET:
			childNode = &Selector{
				Selectors: make([]SelectorValue, 0),
				Rules:     make([]Node, 0),
			}
			visitSelector(pv, childNode)
		case tokens.AT:
			childNode = pv.getAtRule()
			visitAt(pv, childNode)
		default:
			pv.addError("Unexpected token at stylesheet level", pv.currentToken)
			pv.advance()
			continue
		}

		if node != nil {
			s.Rules = append(s.Rules, childNode)
		}
	}
}
