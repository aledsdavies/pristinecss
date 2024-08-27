package parser

import (
	"fmt"
	"strings"

	"github.com/aledsdavies/pristinecss/pkg/tokens"
)

const (
	NodeValue NodeType = "Value"
)

func init() {
	RegisterNodeType(NodeValue, func(pv *ParseVisitor, node Node) {
		switch v := node.(type) {
		case *BasicValue:
			visitBasicValue(pv, v)
		case *StringValue:
			visitStringValue(pv, v)
		case *FunctionValue:
			visitFunctionValue(pv, v)
		default:
			pv.addError(fmt.Sprintf("Unknown value type: %T", v), pv.currentToken)
		}
	})
}

type ValueType int

const (
	Basic ValueType = iota
	String
	Function
	// Add more value types here as needed
)

type Value interface {
	Node
	ValueType() ValueType
}

var _ Value = (*BasicValue)(nil)
var _ Value = (*StringValue)(nil)
var _ Value = (*FunctionValue)(nil)

type BasicValue struct {
	Value []byte
}

func (bv *BasicValue) ValueType() ValueType { return Basic }
func (bv *BasicValue) Type() NodeType       { return NodeValue }
func (bv *BasicValue) String() string {
	return fmt.Sprintf("BasicValue{Value: %q}", string(bv.Value))
}

type StringValue struct {
	SingleQuote bool
	Value       []byte
}

func (bv *StringValue) ValueType() ValueType { return String }
func (bv *StringValue) Type() NodeType       { return NodeValue }
func (bv *StringValue) String() string {
	return fmt.Sprintf("StringValue{SingleQuote: %v, Value: %q}", bv.SingleQuote, string(bv.Value))
}

var _ Value = (*FunctionValue)(nil)

type FunctionValue struct {
	Name      []byte
	Arguments []Value
}

func (fv *FunctionValue) ValueType() ValueType { return Function }
func (fv *FunctionValue) Type() NodeType       { return NodeValue }
func (fv *FunctionValue) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("FunctionValue{Name: %q, Arguments: [", string(fv.Name)))
	for i, arg := range fv.Arguments {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(arg.String())
	}
	sb.WriteString("]}")
	return sb.String()
}

func visitBasicValue(pv *ParseVisitor, node Node) {
	bv := node.(*BasicValue)
	bv.Value = pv.currentToken.Literal
	pv.advance()
}

func visitStringValue(pv *ParseVisitor, node Node) {
	sv := node.(*StringValue)
	str := pv.currentToken.Literal
	sv.SingleQuote = str[0] == '\''
	sv.Value = str[1 : len(str)-1] // Remove the quotes
	pv.advance()
}

func visitFunctionValue(pv *ParseVisitor, node Node) {
	fv := node.(*FunctionValue)
	fv.Name = pv.currentToken.Literal
	pv.advance() // Move to '('
	pv.advance() // Move past '('

	for !pv.currentTokenIs(tokens.RPAREN) && !pv.currentTokenIs(tokens.EOF) {
		fv.Arguments = append(fv.Arguments, pv.parseValue())
		if pv.currentTokenIs(tokens.COMMA) {
			pv.advance()
		}
	}

	pv.consume(tokens.RPAREN, "Expected ')' to close function")
}

func (pv *ParseVisitor) parseValue() Value {
	var value Value
	switch pv.currentToken.Type {
	case tokens.NUMBER:
		return pv.parseNumberValue()
	case tokens.IDENT:
		if pv.nextTokenIs(tokens.LPAREN) {
			value = &FunctionValue{}
		} else {
			value = &BasicValue{}
		}
	case tokens.URI:
		value = pv.parseURLValue()
		pv.advance()
		return value
	case tokens.STRING:
		value = &StringValue{}
	default:
		value = &BasicValue{}
	}

	handler := GetNodeHandler(value)
	handler(pv, value)

	return value
}

func (pv *ParseVisitor) parseNumberValue() Value {
	number := pv.currentToken.Literal
	pv.advance()
	if pv.currentTokenIs(tokens.PERCENTAGE) || isUnit(pv.currentToken.Literal) {
		number = append(number, pv.currentToken.Literal...)
		pv.advance()
	}
	return &BasicValue{Value: number}
}

func (pv *ParseVisitor) parseURLValue() Value {
	urlContent, singleQuote, quoteless := extractURLContent(pv.currentToken.Literal)

	var arg Value
	if quoteless {
		arg = &BasicValue{Value: urlContent}
	} else {
		arg = &StringValue{SingleQuote: singleQuote, Value: urlContent}
	}

	return &FunctionValue{
		Name:      []byte("url"),
		Arguments: []Value{arg},
	}
}

// Helper function to extract the contents of the url() function
func extractURLContent(uri []byte) ([]byte, bool, bool) {
	// Remove "url(" from the beginning and ")" from the end
	content := uri[4 : len(uri)-1]

	// Check if the content is quoted
	singleQuote := false
	quotless := false
	if len(content) >= 2 {
		if content[0] == '\'' && content[len(content)-1] == '\'' {
			singleQuote = true
			content = content[1 : len(content)-1]
		} else if content[0] == '"' && content[len(content)-1] == '"' {
			content = content[1 : len(content)-1]
		} else {
			quotless = true
		}
	}

	return content, singleQuote, quotless
}
