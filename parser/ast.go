package parser

import (
	"fmt"
	"strings"
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
)

type Node interface {
	Type() NodeType
	String() string
}

type Stylesheet struct {
	// Comments
	// At Rules
	// Selectors

	Rules []Node
}

func (s *Stylesheet) Type() NodeType { return NodeStylesheet }

func NewStylesheet() *Stylesheet {
	return &Stylesheet{Rules: []Node{}}
}

func (s *Stylesheet) String() string {
	var sb strings.Builder
	sb.WriteString("Stylesheet{\n")
	for _, rule := range s.Rules {
		sb.WriteString("  " + indentLines(rule.String(), 2) + "\n")
	}
	sb.WriteString("}")
	return sb.String()
}

type Comment struct {
	Text []byte
}

func (c *Comment) Type() NodeType { return NodeComment }

func (c *Comment) String() string {
	return fmt.Sprintf("Comment{Text: %q}", string(c.Text))
}

type Selector struct {
    Selectors []SelectorValue

	// Comment
	// At Rules
	// Selectors (nested)
	// Declerations
	Rules []Node
}

func (c *Selector) Type() NodeType { return NodeSelector }

func (s *Selector) String() string {
	var sb strings.Builder
	sb.WriteString("Selector{\n")
	sb.WriteString("  Selectors: [")
	for i, sel := range s.Selectors {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(sel.String())
	}
	sb.WriteString("]\n")
	if len(s.Rules) > 0 {
		sb.WriteString("  Rules: [\n")
		for _, rule := range s.Rules {
			sb.WriteString("    " + indentLines(rule.String(), 4) + "\n")
		}
		sb.WriteString("  ]\n")
	} else {
		sb.WriteString("  Rules: []\n")
	}
	sb.WriteString("}")
	return sb.String()
}

type SelectorType int

const (
	Element SelectorType = iota
	Class
	ID
	Attribute
	Combinator
)

type SelectorValue struct {
	Type  SelectorType
	Value []byte
}

func (sv SelectorValue) String() string {
	return fmt.Sprintf("{Type: %s, Value: %q}", selectorTypeToString(sv.Type), string(sv.Value))
}

type Decleration struct {
	Key   []rune
	Value [][]byte
}

func (c *Decleration) Type() NodeType { return NodeDeclaration }

func selectorTypeToString(st SelectorType) string {
	switch st {
	case Element:
		return "Element"
	case Class:
		return "Class"
	case ID:
		return "ID"
	case Attribute:
		return "Attribute"
	default:
		return fmt.Sprintf("Unknown(%d)", st)
	}
}

func indentLines(s string, spaces int) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		if i > 0 {
			lines[i] = strings.Repeat(" ", spaces) + line
		}
	}
	return strings.Join(lines, "\n")
}
