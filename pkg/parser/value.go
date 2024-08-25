package parser

import (
	"fmt"
	"strings"

	"github.com/aledsdavies/pristinecss/pkg/tokens"
)

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

type BasicValue struct {
	Value []byte
}

func (bv *BasicValue) ValueType() ValueType { return Basic }
func (bv *BasicValue) Type() NodeType       { return NodeValue }
func (bv *BasicValue) Accept(v Visitor)     {}
func (bv *BasicValue) String() string {
	return fmt.Sprintf("BasicValue{Value: %q}", string(bv.Value))
}

type StringValue struct {
	SingleQuote bool
	Value       []byte
}

func (bv *StringValue) ValueType() ValueType { return String }
func (bv *StringValue) Type() NodeType       { return NodeValue }
func (bv *StringValue) Accept(v Visitor)     {}
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
func (fv *FunctionValue) Accept(v Visitor)     {}
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

func (pv *ParseVisitor) parseValue() Value {
	var value Value
	switch pv.currentToken.Type {
	case tokens.NUMBER:
		value = pv.parseNumberValue()
	case tokens.IDENT:
		if pv.nextTokenIs(tokens.LPAREN) {
			value = pv.parseFunctionValue()
		} else {
			value = &BasicValue{Value: pv.currentToken.Literal}
			pv.advance()
		}
	case tokens.URI:
		urlContent, singleQuote, quoteless := extractURLContent(pv.currentToken.Literal)

		args := []Value{}
		if quoteless {
			args = append(args, &BasicValue{Value: urlContent})
		} else {
			args = append(args, &StringValue{SingleQuote: singleQuote, Value: urlContent})
		}

		value = &FunctionValue{
			Name:      []byte("url"),
			Arguments: args,
		}
		pv.advance()
	case tokens.STRING:
		value = pv.parseStringValue()
	default:
		value = &BasicValue{Value: pv.currentToken.Literal}
		pv.advance()
	}
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

func (pv *ParseVisitor) parseFunctionValue() Value {
	functionName := pv.currentToken.Literal
	pv.advance() // Move to '('
	pv.advance() // Move past '('
	fv := &FunctionValue{Name: functionName}

	for !pv.currentTokenIs(tokens.RPAREN) && !pv.currentTokenIs(tokens.EOF) {
		fv.Arguments = append(fv.Arguments, pv.parseValue())
		if pv.currentTokenIs(tokens.COMMA) {
			pv.advance()
		}
	}

	if pv.currentTokenIs(tokens.RPAREN) {
		pv.advance() // Move past ')'
	} else {
		pv.addError("Expected ')' to close function", pv.currentToken)
	}

	return fv
}

func (pv *ParseVisitor) parseStringValue() Value {
	str := pv.currentToken.Literal
	singleQuote := str[0] == '\''
	// Remove the quotes
	str = str[1 : len(str)-1]

	value := &StringValue{
		SingleQuote: singleQuote,
		Value:       str,
	}
	pv.advance()
	return value
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
