package parser

type Visitor interface {
	VisitStylesheet(*Stylesheet)
	VisitSelector(*Selector)
	VisitDeclaration(*Declaration)
	VisitComment(*Comment)

    // At Rules
	VisitMediaAtRule(m *MediaAtRule)
	VisitKeyframesAtRule(k *KeyframesAtRule)
	VisitKeyframeStop(ks *KeyframeStop)
	VisitImportAtRule(r *ImportAtRule)
	VisitFontFaceAtRule(r *FontFaceAtRule)
	VisitCharsetAtRule(r *CharsetAtRule)
}
