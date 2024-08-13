package parser

import (
	"fmt"
	"strings"

	"github.com/aledsdavies/pristinecss/pkg/tokens"
)

type AtType int

const (
	AtRuleMedia AtType = iota
	AtRuleKeyframe
	AtRuleImport
	AtRuleFontFace
	AtRuleCharset
	// Add more at-rule types here as needed
)

type AtRule interface {
	Node
	AtType() AtType
}

func (pv *ParseVisitor) parseAtRule() Node {
	pv.advance() // Consume '@'
	atRuleName := string(pv.currentToken.Literal)

	switch atRuleName {
	case "media":
		mediaAtRule := &MediaAtRule{
			Name: pv.currentToken.Literal,
		}
		mediaAtRule.Accept(pv)
		return mediaAtRule
	case "keyframes", "-webkit-keyframes":
		keyframesAtRule := &KeyframesAtRule{
			WebKitPrefix: atRuleName == "-webkit-keyframes",
			Name:         pv.currentToken.Literal,
		}
		keyframesAtRule.Accept(pv)
		return keyframesAtRule
	case "import":
		imp := &ImportAtRule{}
		imp.Accept(pv)
		return imp
	case "font-face":
		fontFace := &FontFaceAtRule{Declarations: []Declaration{}}
		fontFace.Accept(pv)
		return fontFace
	case "charset":
		charset := &CharsetAtRule{}
		charset.Accept(pv)
		return charset
	// Add cases for other at-rules as needed
	default:
		pv.addError("Unsupported at-rule", pv.currentToken)
		pv.skipToNextRule()
		return nil
	}
}

// TODO add supports conditions
type ImportAtRule struct {
    URL     Value
    Layer   Value
    Media   MediaQuery
}

func (r *ImportAtRule) Type() NodeType   { return NodeAtRule }
func (r *ImportAtRule) AtType() AtType   { return AtRuleImport }
func (r *ImportAtRule) Accept(v Visitor) { v.VisitImportAtRule(r) }
func (r *ImportAtRule) String() string {
    var sb strings.Builder
    sb.WriteString("ImportAtRule{\n")
    sb.WriteString(fmt.Sprintf("  URL: %s,\n", r.URL.String()))
    if r.Layer != nil {
        sb.WriteString(fmt.Sprintf("  Layer: %s,\n", r.Layer.String()))
    }
    if len(r.Media.Queries) > 0 {
        sb.WriteString(fmt.Sprintf("  Media: %s,\n", r.Media.String()))
    }
    sb.WriteString("}")
    return sb.String()
}

func (pv *ParseVisitor) VisitImportAtRule(r *ImportAtRule) {
    pv.advance() // Consume 'import'
    if pv.currentTokenIs(tokens.URI) || pv.currentTokenIs(tokens.STRING) {
        r.URL = pv.parseValue()
    } else {
        pv.addError("Expected string or URI after @import", pv.currentToken)
        return
    }

    for !pv.currentTokenIs(tokens.SEMICOLON) && !pv.currentTokenIs(tokens.EOF) {
        currTok := string(pv.currentToken.Literal)
        switch {
        case pv.currentTokenIs(tokens.IDENT) && currTok == "layer":
            if pv.currentTokenIs(tokens.LPAREN) {
                r.Layer = pv.parseValue() // This should parse the function-like syntax
            } else {
                pv.advance()
                r.Layer = &BasicValue{Value: []byte("layer")}
            }
        case pv.currentTokenIs(tokens.IDENT) || pv.currentTokenIs(tokens.LPAREN):
            r.Media = *pv.parseMediaQuery()
        default:
            pv.addError("Unexpected token in @import rule", pv.currentToken)
            return
        }
    }

    pv.consume(tokens.SEMICOLON, "Expected ';' after @import rule")
}

type FontFaceAtRule struct {
	Declarations []Declaration
}

func (r *FontFaceAtRule) Type() NodeType   { return NodeAtRule }
func (r *FontFaceAtRule) AtType() AtType   { return AtRuleFontFace }
func (r *FontFaceAtRule) Accept(v Visitor) { v.VisitFontFaceAtRule(r) }
func (r *FontFaceAtRule) String() string {
	var sb strings.Builder
	sb.WriteString("FontFaceAtRule{\n")
	sb.WriteString("  Declarations: [\n")
	for _, decl := range r.Declarations {
		sb.WriteString(indentLines(decl.String(), 4))
		sb.WriteString(",\n")
	}
	sb.WriteString("  ]\n")
	sb.WriteString("}")
	return sb.String()
}

func (pv *ParseVisitor) VisitFontFaceAtRule(r *FontFaceAtRule) {
	pv.advance() // Consume 'font-face'

	if !pv.consume(tokens.LBRACE, "Expected '{' after @font-face") {
		return
	}

	for !pv.currentTokenIs(tokens.RBRACE) && !pv.currentTokenIs(tokens.EOF) {
		declaration := &Declaration{
			Key:   pv.currentToken.Literal,
		}
		declaration.Accept(pv)
		r.Declarations = append(r.Declarations, *declaration)

		if pv.currentTokenIs(tokens.SEMICOLON) {
			pv.advance()
		}
	}

	pv.consume(tokens.RBRACE, "Expected '}' to close @font-face rule")
}

type CharsetAtRule struct {
	Charset Value
}

func (r *CharsetAtRule) Type() NodeType   { return NodeAtRule }
func (r *CharsetAtRule) AtType() AtType   { return AtRuleCharset }
func (r *CharsetAtRule) Accept(v Visitor) { v.VisitCharsetAtRule(r) }
func (r *CharsetAtRule) String() string {
	return fmt.Sprintf("CharsetAtRule{Charset: %q}", r.Charset)
}

func (pv *ParseVisitor) VisitCharsetAtRule(r *CharsetAtRule) {
	pv.advance() // Consume 'charset'

	if pv.currentTokenIs(tokens.STRING) {
		r.Charset = pv.parseValue()
	} else {
		pv.addError("Expected string after @charset", pv.currentToken)
		return
	}

	pv.consume(tokens.SEMICOLON, "Expected ';' after @charset rule")
}
