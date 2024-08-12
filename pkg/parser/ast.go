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

type Visitor interface {
	VisitStylesheet(*Stylesheet)
	VisitSelector(*Selector)
	VisitDeclaration(*Declaration)
	VisitAtRule(*AtRule)
	VisitComment(*Comment)
}

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

func (s *Stylesheet) Type() NodeType   { return NodeStylesheet }
func (s *Stylesheet) Accept(v Visitor) { v.VisitStylesheet(s) }

func (s *Stylesheet) String() string {
	var sb strings.Builder
	sb.WriteString("Stylesheet{\n")
	for _, rule := range s.Rules {
		sb.WriteString(indentLines(rule.String(), 2) + "\n")
	}
	sb.WriteString("}")
	return sb.String()
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

var _ Node = (*Selector)(nil)

type Selector struct {
	Selectors []SelectorValue

	// Comment
	// At Rules
	// Selectors (nested)
	// Declerations
	Rules []Node
}

func (c *Selector) Type() NodeType   { return NodeSelector }
func (s *Selector) Accept(v Visitor) { v.VisitSelector(s) }
func (s *Selector) String() string {
	var sb strings.Builder
	sb.WriteString("Selector{\n")
	sb.WriteString("  Selectors: [\n")
	for _, sel := range s.Selectors {
		sb.WriteString("    " + sel.String() + ",\n")
	}
	sb.WriteString("  ]\n")
	if len(s.Rules) > 0 {
		sb.WriteString("  Rules: [\n")
		for _, rule := range s.Rules {
			sb.WriteString(indentLines(rule.String(), 4) + "\n")
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
	Pseudo
	Combinator
)

type SelectorValue struct {
	Type  SelectorType
	Value []byte
}

func (sv SelectorValue) String() string {
	return fmt.Sprintf("{Type: %s, Value: %q}", selectorTypeToString(sv.Type), sv.Value)
}

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

type AtType int

const (
	MEDIA AtType = iota
    KEYFRAMES
	// Add more at-rule types here as needed
)

type AtQuery interface {
	AtType() AtType
	String() string
}

type AtRule struct {
	Name  []byte
	Query AtQuery
	Rules []Node // For nested rules within the at-rule
}

func (a *AtRule) Type() NodeType   { return NodeAtRule }
func (a *AtRule) Accept(v Visitor) { v.VisitAtRule(a) }
func (a *AtRule) String() string {
	var sb strings.Builder
	sb.WriteString("AtRule{\n")
	sb.WriteString(fmt.Sprintf("  Name: %q,\n", a.Name))
	if a.Query != nil {
		sb.WriteString("  Query: ")
		sb.WriteString(indentLines(a.Query.String(), 4))
		sb.WriteString(",\n")
	}
	if len(a.Rules) > 0 {
		sb.WriteString("  Rules: [\n")
		for _, rule := range a.Rules {
			sb.WriteString(indentLines(rule.String(), 4) + ",\n")
		}
		sb.WriteString("  ],\n")
	}
	sb.WriteString("}")
	return sb.String()
}

type MediaQuery struct {
	Queries []MediaQueryExpression
}

func (mq MediaQuery) AtType() AtType { return MEDIA }

func (mq MediaQuery) String() string {
	var sb strings.Builder
	sb.WriteString("MediaQuery{\n")
	sb.WriteString("  Queries: [\n")
	for _, query := range mq.Queries {
		sb.WriteString(indentLines(query.String(), 4) + ",\n")
	}
	sb.WriteString("  ]\n")
	sb.WriteString("}")
	return sb.String()
}

type MediaQueryExpression struct {
	MediaType []byte
	Not       bool
	Only      bool
	Features  []MediaFeature
}

func (mqe MediaQueryExpression) String() string {
	var sb strings.Builder
	sb.WriteString("MediaQueryExpression{\n")
	sb.WriteString(fmt.Sprintf("  MediaType: %q,\n", mqe.MediaType))
	sb.WriteString(fmt.Sprintf("  Not: %v,\n", mqe.Not))
	sb.WriteString(fmt.Sprintf("  Only: %v,\n", mqe.Only))
	sb.WriteString("  Features: [\n")
	for _, feature := range mqe.Features {
		sb.WriteString(indentLines(feature.String(), 4) + ",\n")
	}
	sb.WriteString("  ]\n")
	sb.WriteString("}")
	return sb.String()
}

type MediaFeature struct {
	Name  []byte
	Value []byte
}

func (mf MediaFeature) String() string {
	var valueStr string
	if mf.Value != nil {
		valueStr = fmt.Sprintf("%q", mf.Value)
	} else {
		valueStr = "nil"
	}
	return fmt.Sprintf("MediaFeature{Name: %q, Value: %s}", mf.Name, valueStr)
}

type KeyframesRule struct {
	Name  []byte
	Stops []KeyframeStop
}

func (kr KeyframesRule) AtType() AtType { return KEYFRAMES }

func (kr KeyframesRule) String() string {
	var sb strings.Builder
	sb.WriteString("KeyframesRule{\n")
	sb.WriteString(fmt.Sprintf("  Name: %q,\n", kr.Name))
	sb.WriteString("  Stops: [\n")
	for _, stop := range kr.Stops {
		sb.WriteString(indentLines(stop.String(), 4) + ",\n")
	}
	sb.WriteString("  ]\n")
	sb.WriteString("}")
	return sb.String()
}

type KeyframeStop struct {
	Selectors [][]byte // Could be percentages or "from"/"to"
	Rules     []Node   // Declarations for this keyframe stop
}

func (ks KeyframeStop) String() string {
	var sb strings.Builder
	sb.WriteString("KeyframeStop{\n")
	sb.WriteString("  Selectors: [")
	for i, selector := range ks.Selectors {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(fmt.Sprintf("%q", selector))
	}
	sb.WriteString("],\n")
	sb.WriteString("  Rules: [\n")
	for _, rule := range ks.Rules {
		sb.WriteString(indentLines(rule.String(), 4) + ",\n")
	}
	sb.WriteString("  ]\n")
	sb.WriteString("}")
	return sb.String()
}

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
	case Pseudo:
		return "Pseudo"
	case Combinator:
		return "Combinator"
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
