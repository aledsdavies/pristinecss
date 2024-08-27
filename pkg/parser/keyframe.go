package parser

import (
	"fmt"
	"strings"

	"github.com/aledsdavies/pristinecss/pkg/tokens"
)

const (
	NodeKeyframeStop NodeType = "KeyframeStop"
	Keyframe         AtType   = "keyframes"
	WebkitKeyframe   AtType   = "-webkit-keyframes"
)

func init() {
	RegisterAt(Keyframe, visitKeyframesAtRule, func() AtRule { return &KeyframesAtRule{} })
	RegisterAt(WebkitKeyframe, visitKeyframesAtRule, func() AtRule { return &KeyframesAtRule{WebKitPrefix: true} })
	RegisterNodeType(NodeKeyframeStop, visitKeyframeStop)
}

var _ Node = (*KeyframesAtRule)(nil)

type KeyframesAtRule struct {
	WebKitPrefix bool
	Name         []byte
	Stops        []KeyframeStop
}

func (k *KeyframesAtRule) Type() NodeType { return NodeAtRule }
func (k *KeyframesAtRule) AtType() AtType { return Keyframe }
func (k *KeyframesAtRule) String() string {
	var sb strings.Builder
	sb.WriteString("KeyframesAtRule{\n")
	if k.WebKitPrefix {
		sb.WriteString("  WebKitPrefix: true,\n")
	}
	sb.WriteString(fmt.Sprintf("  Name: %q,\n", string(k.Name)))
	sb.WriteString("  Stops: [\n")
	for _, stop := range k.Stops {
		sb.WriteString(indentLines(stop.String(), 4))
		sb.WriteString(",\n")
	}
	sb.WriteString("  ]\n")
	sb.WriteString("}")
	return sb.String()
}

type KeyframeStop struct {
	Stops []Value // Could be percentages or "from"/"to"
	Rules []Node  // Declarations for this keyframe stop
}

func (ks *KeyframeStop) Type() NodeType { return NodeKeyframeStop }
func (ks KeyframeStop) String() string {
	var sb strings.Builder
	sb.WriteString("KeyframeStop{\n")
	sb.WriteString("  Stops: [\n")
	for _, vl := range ks.Stops {
		sb.WriteString(fmt.Sprintf("%s\n", vl.String()))
	}
	sb.WriteString("  ]\n")
	sb.WriteString("  Rules: [\n")
	for _, rule := range ks.Rules {
		sb.WriteString(indentLines(rule.String(), 4) + ",\n")
	}
	sb.WriteString("  ]\n")
	sb.WriteString("}")
	return sb.String()
}

func visitKeyframesAtRule(pv *ParseVisitor, node AtRule) {
	k := node.(*KeyframesAtRule)
	pv.advance() // Consume 'keyframes'

	if !pv.currentTokenIs(tokens.IDENT) {
		pv.addError("Expected keyframes name", pv.currentToken)
		return
	}

	k.Name = pv.currentToken.Literal
	pv.advance()

	if !pv.consume(tokens.LBRACE, "Expected '{' after @keyframes name") {
		return
	}

	for !pv.currentTokenIs(tokens.RBRACE) && !pv.currentTokenIs(tokens.EOF) {
		stop := &KeyframeStop{
			Rules: make([]Node, 0),
		}
		visitKeyframeStop(pv, stop)
		k.Stops = append(k.Stops, *stop)
	}

	pv.consume(tokens.RBRACE, "Expected '}' to close @keyframes block")
}

func visitKeyframeStop(pv *ParseVisitor, node Node) {
	ks := node.(*KeyframeStop)
	for !pv.currentTokenIs(tokens.LBRACE) && !pv.currentTokenIs(tokens.EOF) {
		switch pv.currentToken.Type {
		case tokens.IDENT, tokens.NUMBER:
			ks.Stops = append(ks.Stops, pv.parseValue())
		case tokens.COMMA:
			pv.advance()
		default:
			pv.addError("Unexpected token in keyframe selector", pv.currentToken)
			pv.advance()
			pv.skipToNextSemicolonOrBrace()
			return
		}
	}

	if !pv.consume(tokens.LBRACE, "Expected '{' after keyframe selector") {
		return
	}

	for !pv.currentTokenIs(tokens.RBRACE) && !pv.currentTokenIs(tokens.EOF) {
		if !pv.currentTokenIs(tokens.IDENT) {
			pv.addError("Expected property name", pv.currentToken)
			pv.skipToNextSemicolonOrBrace()
			continue
		}
		declaration := &Declaration{
			Key: pv.currentToken.Literal,
		}
		visitDeclaration(pv, declaration)
		ks.Rules = append(ks.Rules, declaration)

		if pv.currentTokenIs(tokens.SEMICOLON) {
			pv.advance() // Consume ';'
		}
	}

	pv.consume(tokens.RBRACE, "Expected '}' at the end of keyframe block")
}
