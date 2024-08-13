package parser

import (
	"fmt"
	"strings"

	"github.com/aledsdavies/pristinecss/pkg/tokens"
)

type NodeType int

const (
	NodeStylesheet NodeType = iota
	NodeAtRule
	NodeRuleSet
	NodeSelector
	NodeDeclaration
	NodeFunction
	NodeComment
    NodeKeyframeStop
    NodeValue
)

type Node interface {
	Type() NodeType
	String() string
	Accept(Visitor)
}

var _ Node = (*Stylesheet)(nil)

type Stylesheet struct {
	// Comments
	// At Rules
	// Selectors

	Rules []Node
}

func NewStylesheet() *Stylesheet {
	return &Stylesheet{Rules: []Node{}}
}

func (s *Stylesheet) Type() NodeType { return NodeStylesheet }

func (s *Stylesheet) Accept(v Visitor) { v.VisitStylesheet(s) }

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

func (pv *ParseVisitor) VisitStylesheet(s *Stylesheet) {
	for !pv.currentTokenIs(tokens.EOF) {
		switch pv.currentToken.Type {
		case tokens.COMMENT:
			comment := &Comment{Text: pv.currentToken.Literal}
			comment.Accept(pv)
			s.Rules = append(s.Rules, comment)
			pv.advance()
		case tokens.DOT, tokens.HASH, tokens.COLON, tokens.DBLCOLON, tokens.IDENT, tokens.LBRACKET:
			selector := &Selector{
				Selectors: make([]SelectorValue, 0),
				Rules:     make([]Node, 0),
			}
			selector.Accept(pv)
			s.Rules = append(s.Rules, selector)
		case tokens.AT:
			atRule := pv.parseAtRule()
			s.Rules = append(s.Rules, atRule)
		default:
			pv.addError("Unexpected token at stylesheet level", pv.currentToken)
			pv.advance()
		}
	}
}

var _ Node = (*Comment)(nil)

type Comment struct {
	Text []byte
}

func (c *Comment) Type() NodeType   { return NodeComment }
func (c *Comment) Accept(v Visitor) { v.VisitComment(c) }
func (c *Comment) String() string {
	return fmt.Sprintf("Comment{Text: %q}", string(c.Text))
}

func (pv *ParseVisitor) VisitComment(c *Comment) {
	// Nothing to do here, as comments are simple tokens
}

