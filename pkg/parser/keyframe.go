package parser

import (
	"fmt"
	"strings"

	"github.com/aledsdavies/pristinecss/pkg/tokens"
)

var _ Node = (*KeyframesAtRule)(nil)

type KeyframesAtRule struct {
	WebKitPrefix bool
	Name         []byte
	Stops        []KeyframeStop
}

func (k *KeyframesAtRule) Type() NodeType { return NodeAtRule }

func (k *KeyframesAtRule) AtType() AtType { return KEYFRAMES }

func (k *KeyframesAtRule) Accept(v Visitor) { v.VisitKeyframesAtRule(k) }

func (k *KeyframesAtRule) String() string {
	var sb strings.Builder
	sb.WriteString("KeyframesAtRule{\n")
	if k.WebKitPrefix {
		sb.WriteString(fmt.Sprintf("  Has Webkit Prefix: %t,\n", k.WebKitPrefix))
	}
	sb.WriteString(fmt.Sprintf("  Name: %q,\n", k.Name))
	if len(k.Stops) > 0 {
		sb.WriteString("  Stops: [\n")
		for _, stop := range k.Stops {
			sb.WriteString(indentLines(stop.String(), 4))
			sb.WriteString(",\n")
		}
		sb.WriteString("  ]\n")
	}
	sb.WriteString("}")
	return sb.String()
}

type KeyframeStop struct {
	Selectors [][]byte // Could be percentages or "from"/"to"
	Rules     []Node   // Declarations for this keyframe stop
}

func (ks *KeyframeStop) Type() NodeType { return NodeKeyframeStop }
func (ks *KeyframeStop) Accept(v Visitor) { v.VisitKeyframeStop(ks) }
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

func (pv *ParseVisitor) VisitKeyframesAtRule(k *KeyframesAtRule) {
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
            Selectors: make([][]byte, 0),
            Rules:     make([]Node, 0),
        }
        stop.Accept(pv)
        k.Stops = append(k.Stops, *stop)
    }

    pv.consume(tokens.RBRACE, "Expected '}' to close @keyframes block")
}

func (pv *ParseVisitor) VisitKeyframeStop(ks *KeyframeStop) {
    for !pv.currentTokenIs(tokens.LBRACE) && !pv.currentTokenIs(tokens.EOF) {
        switch pv.currentToken.Type {
        case tokens.IDENT:
            ks.Selectors = append(ks.Selectors, pv.currentToken.Literal)
            pv.advance()
        case tokens.NUMBER:
            number := pv.currentToken.Literal
            pv.advance()
            if pv.currentTokenIs(tokens.PERCENTAGE) || isUnit(pv.currentToken.Literal) {
                number = append(number, pv.currentToken.Literal...)
                pv.advance()
            }
            ks.Selectors = append(ks.Selectors, number)
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
            Key:   pv.currentToken.Literal,
            Value: make([][]byte, 0),
        }
        declaration.Accept(pv)
        ks.Rules = append(ks.Rules, declaration)

        if pv.currentTokenIs(tokens.SEMICOLON) {
            pv.advance() // Consume ';'
        }
    }

    pv.consume(tokens.RBRACE, "Expected '}' at the end of keyframe block")
}
