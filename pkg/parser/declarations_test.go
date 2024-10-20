package parser

import (
	"testing"
)

func TestDeclarations(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *Stylesheet
	}{
		{
			name:  "Basic declaration",
			input: "div { color: blue; }",
			expected: &Stylesheet{
				Rules: []Node{
					&Selector{
						Selectors: []SelectorValue{{Type: Element, Value: []byte("div")}},
						Rules: []Node{
							&Declaration{Key: []byte("color"), Value: []Value{&BasicValue{Value: []byte("blue")}}},
						},
					},
				},
			},
		},
		{
			name:  "Declaration with !important",
			input: "div { color: red !important; }",
			expected: &Stylesheet{
				Rules: []Node{
					&Selector{
						Selectors: []SelectorValue{{Type: Element, Value: []byte("div")}},
						Rules: []Node{
							&Declaration{
								Key:       []byte("color"),
								Value:     []Value{&BasicValue{Value: []byte("red")}},
								Important: true,
							},
						},
					},
				},
			},
		},
		{
			name:  "Declaration with multiple values",
			input: "div { margin: 10px 20px 30px 40px; }",
			expected: &Stylesheet{
				Rules: []Node{
					&Selector{
						Selectors: []SelectorValue{{Type: Element, Value: []byte("div")}},
						Rules: []Node{
							&Declaration{
								Key: []byte("margin"),
								Value: []Value{
									&BasicValue{Value: []byte("10px")},
									&BasicValue{Value: []byte("20px")},
									&BasicValue{Value: []byte("30px")},
									&BasicValue{Value: []byte("40px")},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "Declaration with function value",
			input: "div { background-color: rgb(255, 0, 0); }",
			expected: &Stylesheet{
				Rules: []Node{
					&Selector{
						Selectors: []SelectorValue{{Type: Element, Value: []byte("div")}},
						Rules: []Node{
							&Declaration{
								Key: []byte("background-color"),
								Value: []Value{
									&FunctionValue{
										Name: []byte("rgb"),
										Arguments: []Value{
											&BasicValue{Value: []byte("255")},
											&BasicValue{Value: []byte("0")},
											&BasicValue{Value: []byte("0")},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "Declaration with quoted string value",
			input: `div { content: "Hello, world!"; }`,
			expected: &Stylesheet{
				Rules: []Node{
					&Selector{
						Selectors: []SelectorValue{{Type: Element, Value: []byte("div")}},
						Rules: []Node{
							&Declaration{
								Key: []byte("content"),
								Value: []Value{
									&StringValue{Value: []byte("Hello, world!"), SingleQuote: false},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "Declaration with url function",
			input: `div { background-image: url('image.jpg'); }`,
			expected: &Stylesheet{
				Rules: []Node{
					&Selector{
						Selectors: []SelectorValue{{Type: Element, Value: []byte("div")}},
						Rules: []Node{
							&Declaration{
								Key: []byte("background-image"),
								Value: []Value{
									&FunctionValue{
										Name: []byte("url"),
										Arguments: []Value{
											&StringValue{Value: []byte("image.jpg"), SingleQuote: true},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "Declaration with calc function",
			input: "div { width: calc(100% - 20px); }",
			expected: &Stylesheet{
				Rules: []Node{
					&Selector{
						Selectors: []SelectorValue{{Type: Element, Value: []byte("div")}},
						Rules: []Node{
							&Declaration{
								Key: []byte("width"),
								Value: []Value{
									&FunctionValue{
										Name: []byte("calc"),
										Arguments: []Value{
											&BasicValue{Value: []byte("100%")},
											&BasicValue{Value: []byte("-")},
											&BasicValue{Value: []byte("20px")},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "Declaration with multiple functions",
			input: "div { background: linear-gradient(to right, rgb(255,0,0), rgba(0,0,255,0.5)); }",
			expected: &Stylesheet{
				Rules: []Node{
					&Selector{
						Selectors: []SelectorValue{{Type: Element, Value: []byte("div")}},
						Rules: []Node{
							&Declaration{
								Key: []byte("background"),
								Value: []Value{
									&FunctionValue{
										Name: []byte("linear-gradient"),
										Arguments: []Value{
											&BasicValue{Value: []byte("to")},
											&BasicValue{Value: []byte("right")},
											&FunctionValue{
												Name: []byte("rgb"),
												Arguments: []Value{
													&BasicValue{Value: []byte("255")},
													&BasicValue{Value: []byte("0")},
													&BasicValue{Value: []byte("0")},
												},
											},
											&FunctionValue{
												Name: []byte("rgba"),
												Arguments: []Value{
													&BasicValue{Value: []byte("0")},
													&BasicValue{Value: []byte("0")},
													&BasicValue{Value: []byte("255")},
													&BasicValue{Value: []byte("0.5")},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Declaration with color value",
			input: `
			.colorful {
				color: #ff0000;
				background-color: #00ff00;
				border: 1px solid #0000ff;
			}`,
			expected: &Stylesheet{
				Rules: []Node{
					&Selector{
						Selectors: []SelectorValue{{Type: Class, Value: []byte(".colorful")}},
						Rules: []Node{
							&Declaration{Key: []byte("color"), Value: []Value{&BasicValue{Value: []byte("#ff0000")}}},
							&Declaration{Key: []byte("background-color"), Value: []Value{&BasicValue{Value: []byte("#00ff00")}}},
							&Declaration{Key: []byte("border"), Value: []Value{
								&BasicValue{Value: []byte("1px")},
								&BasicValue{Value: []byte("solid")},
								&BasicValue{Value: []byte("#0000ff")},
							}},
						},
					},
				},
			},
		},
	}

	runTests(t, tests)
}
