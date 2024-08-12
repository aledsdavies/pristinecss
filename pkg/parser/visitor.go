package parser

type Visitor interface {
	VisitStylesheet(*Stylesheet)
	VisitSelector(*Selector)
	VisitDeclaration(*Declaration)
	VisitMediaAtRule(m *MediaAtRule)
	VisitKeyframesAtRule(k *KeyframesAtRule)
    VisitKeyframeStop(ks *KeyframeStop)
	VisitComment(*Comment)

}

