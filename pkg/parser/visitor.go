package parser

import "fmt"

type NodeType string

type Node interface {
	Type() NodeType
	String() string
}

type VisitFunc func(*ParseVisitor, Node)
type AtVisitFunc func(*ParseVisitor, AtRule)

var nodeRegistry = make(map[NodeType]VisitFunc)

func RegisterNodeType(nodeType NodeType, visitFunc VisitFunc) {
	nodeRegistry[nodeType] = visitFunc
}

func GetNodeHandler(node Node) VisitFunc {
	handler, exists := nodeRegistry[node.Type()]
	if !exists {
		typeName := fmt.Sprintf("%T", node)
		panic(fmt.Sprintf("Parser configuration error: No handler registered for node type: %s. Please register this node type with the parser.", typeName))
	}
	return handler
}

type AtType string

var atRegistry = make(map[AtType]AtVisitFunc)
var keywordRegistry = make(map[string]AtInit)

func RegisterAt(atType AtType, visitFunc AtVisitFunc, initFunc AtInit) {
	atRegistry[atType] = visitFunc
	keywordRegistry[string(atType)] = initFunc
}

func GetAtHandler(atRule AtRule) AtVisitFunc {
	handler, exists := atRegistry[atRule.AtType()]
	if !exists {
		typeName := fmt.Sprintf("%T", atRule)
		panic(fmt.Sprintf("Parser configuration error: No handler registered for @ type: %s. Please register this @ type with the parser.", typeName))
	}
	return handler
}
