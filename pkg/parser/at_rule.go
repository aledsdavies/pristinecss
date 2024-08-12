package parser

type AtType int

const (
	MEDIA AtType = iota
	KEYFRAMES
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
            Name: pv.currentToken.Literal,
        }
        keyframesAtRule.Accept(pv)
        return keyframesAtRule
    // Add cases for other at-rules as needed
    default:
        pv.addError("Unsupported at-rule", pv.currentToken)
        pv.skipToNextRule()
        return nil
    }
}

