package parser

import (
	"strings"

	"github.com/aledsdavies/pristinecss/pkg/tokens"
)

const(
    FontFace AtType = "font-face"
)

func init() {
	RegisterAt(FontFace, visitFontFaceAtRule, func() AtRule { return &FontFaceAtRule{} })
}

type FontFaceAtRule struct {
	Declarations []Declaration
}

func (r *FontFaceAtRule) Type() NodeType   { return NodeAtRule }
func (r *FontFaceAtRule) AtType() AtType   { return FontFace }
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

func visitFontFaceAtRule(pv *ParseVisitor, node AtRule) {
    ff := node.(*FontFaceAtRule)
	pv.advance() // Consume 'font-face'

	if !pv.consume(tokens.LBRACE, "Expected '{' after @font-face") {
		return
	}

	for !pv.currentTokenIs(tokens.RBRACE) && !pv.currentTokenIs(tokens.EOF) {
		declaration := &Declaration{
			Key: pv.currentToken.Literal,
		}
        visitDeclaration(pv, declaration)
		ff.Declarations = append(ff.Declarations, *declaration)

		if pv.currentTokenIs(tokens.SEMICOLON) {
			pv.advance()
		}
	}

	pv.consume(tokens.RBRACE, "Expected '}' to close @font-face rule")
}
