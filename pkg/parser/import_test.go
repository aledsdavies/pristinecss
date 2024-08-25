package parser

import "testing"

func TestImportAtRule(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *Stylesheet
	}{
		{
			name:  "Simple @import with URL",
			input: `@import url("styles.css");`,
			expected: &Stylesheet{
				Rules: []Node{
					&ImportAtRule{
						URL: &FunctionValue{
							Name: []byte("url"),
							Arguments: []Value{
								&StringValue{SingleQuote: false, Value: []byte("styles.css")},
							},
						},
					},
				},
			},
		},
		{
			name:  "Simple @import with string",
			input: `@import "styles.css";`,
			expected: &Stylesheet{
				Rules: []Node{
					&ImportAtRule{
						URL: &StringValue{SingleQuote: false, Value: []byte("styles.css")},
					},
				},
			},
		},
		{
			name:  "@import with media query",
			input: `@import url("print-styles.css") print;`,
			expected: &Stylesheet{
				Rules: []Node{
					&ImportAtRule{
						URL: &FunctionValue{
							Name: []byte("url"),
							Arguments: []Value{
								&StringValue{SingleQuote: false, Value: []byte("print-styles.css")},
							},
						},
						Media: MediaQuery{
							Queries: []MediaQueryExpression{
								{
									MediaType: []byte("print"),
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "@import with complex media query",
			input: `@import url("mobile-styles.css") screen and (max-width: 600px);`,
			expected: &Stylesheet{
				Rules: []Node{
					&ImportAtRule{
						URL: &FunctionValue{
							Name: []byte("url"),
							Arguments: []Value{
								&StringValue{SingleQuote: false, Value: []byte("mobile-styles.css")},
							},
						},
						Media: MediaQuery{
							Queries: []MediaQueryExpression{
								{
									MediaType: []byte("screen"),
									Features: []MediaFeature{
										{Name: []byte("max-width"), Value: []byte("600px")},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "@import with single-quoted URL",
			input: `@import url('single-quoted.css');`,
			expected: &Stylesheet{
				Rules: []Node{
					&ImportAtRule{
						URL: &FunctionValue{
							Name: []byte("url"),
							Arguments: []Value{
								&StringValue{SingleQuote: true, Value: []byte("single-quoted.css")},
							},
						},
					},
				},
			},
		},
		{
			name:  "@import with unquoted URL",
			input: `@import url(unquoted.css);`,
			expected: &Stylesheet{
				Rules: []Node{
					&ImportAtRule{
						URL: &FunctionValue{
							Name: []byte("url"),
							Arguments: []Value{
								&BasicValue{Value: []byte("unquoted.css")},
							},
						},
					},
				},
			},
		},
		{
			name:  "@import with multiple media queries",
			input: `@import "responsive.css" screen and (color), projection and (color);`,
			expected: &Stylesheet{
				Rules: []Node{
					&ImportAtRule{
						URL: &StringValue{SingleQuote: false, Value: []byte("responsive.css")},
						Media: MediaQuery{
							Queries: []MediaQueryExpression{
								{
									MediaType: []byte("screen"),
									Features: []MediaFeature{
										{Name: []byte("color")},
									},
								},
								{
									MediaType: []byte("projection"),
									Features: []MediaFeature{
										{Name: []byte("color")},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "@import with layer",
			input: `@import url("theme.css") layer(theme);`,
			expected: &Stylesheet{
				Rules: []Node{
					&ImportAtRule{
						URL: &FunctionValue{
							Name: []byte("url"),
							Arguments: []Value{
								&StringValue{SingleQuote: false, Value: []byte("theme.css")},
							},
						},
						Layer: &FunctionValue{
							Name: []byte("layer"),
							Arguments: []Value{
								&BasicValue{Value: []byte("theme")},
							},
						},
					},
				},
			},
		},
		{
			name:  "@import with layer and media query",
			input: `@import url('components.css') layer(framework) screen and (min-width: 800px);`,
			expected: &Stylesheet{
				Rules: []Node{
					&ImportAtRule{
						URL: &FunctionValue{
							Name: []byte("url"),
							Arguments: []Value{
								&StringValue{SingleQuote: true, Value: []byte("components.css")},
							},
						},
						Layer: &FunctionValue{
							Name: []byte("layer"),
							Arguments: []Value{
								&BasicValue{Value: []byte("framework")},
							},
						},
						Media: MediaQuery{
							Queries: []MediaQueryExpression{
								{
									MediaType: []byte("screen"),
									Features: []MediaFeature{
										{Name: []byte("min-width"), Value: []byte("800px")},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "@import with layer without parentheses",
			input: `@import url("base.css") layer;`,
			expected: &Stylesheet{
				Rules: []Node{
					&ImportAtRule{
						URL: &FunctionValue{
							Name: []byte("url"),
							Arguments: []Value{
								&StringValue{SingleQuote: false, Value: []byte("base.css")},
							},
						},
						Layer: &BasicValue{Value: []byte("layer")},
					},
				},
			},
		},
		{
			name:  "@import with named layer without parentheses",
			input: `@import url("theme.css") layer(theme);`,
			expected: &Stylesheet{
				Rules: []Node{
					&ImportAtRule{
						URL: &FunctionValue{
							Name: []byte("url"),
							Arguments: []Value{
								&StringValue{SingleQuote: false, Value: []byte("theme.css")},
							},
						},
						Layer: &FunctionValue{
							Name: []byte("layer"),
							Arguments: []Value{
								&BasicValue{Value: []byte("theme")},
							},
						},
					},
				},
			},
		},
		{
			name:  "@import with simple supports syntax",
			input: `@import url("modern.css") supports(display: flex);`,
			expected: &Stylesheet{
				Rules: []Node{
					&ImportAtRule{
						URL: &FunctionValue{
							Name: []byte("url"),
							Arguments: []Value{
								&StringValue{SingleQuote: false, Value: []byte("modern.css")},
							},
						},
						Supports: &SupportsDecleration{
							Key:   []byte("display"),
							Value: []Value{&BasicValue{Value: []byte("flex")}},
						},
					},
				},
			},
		},
		{
			name:  "@import with complex supports syntax",
			input: `@import url("advanced.css") supports((display: grid) and (color: rebeccapurple));`,
			expected: &Stylesheet{
				Rules: []Node{
					&ImportAtRule{
						URL: &FunctionValue{
							Name: []byte("url"),
							Arguments: []Value{
								&StringValue{SingleQuote: false, Value: []byte("advanced.css")},
							},
						},
						Supports: &SupportsGroup{
							Conditions: []SupportsCondition{
								&SupportsDecleration{
									Key:   []byte("display"),
									Value: []Value{&BasicValue{Value: []byte("grid")}},
								},
								&SupportsOperator{Operator: "and"},
								&SupportsDecleration{
									Key:   []byte("color"),
									Value: []Value{&BasicValue{Value: []byte("rebeccapurple")}},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "@import with supports function",
			input: `@import url("feature.css") supports(selector(:has(> img)));`,
			expected: &Stylesheet{
				Rules: []Node{
					&ImportAtRule{
						URL: &FunctionValue{
							Name: []byte("url"),
							Arguments: []Value{
								&StringValue{SingleQuote: false, Value: []byte("feature.css")},
							},
						},
						Supports: &SupportsFunction{
							Name: []byte("selector"),
							Args: []byte(":has(>img)"),
						},
					},
				},
			},
		},
		{
			name:  "@import with supports not",
			input: `@import url("fallback.css") supports(not (display: flex));`,
			expected: &Stylesheet{
				Rules: []Node{
					&ImportAtRule{
						URL: &FunctionValue{
							Name: []byte("url"),
							Arguments: []Value{
								&StringValue{SingleQuote: false, Value: []byte("fallback.css")},
							},
						},
						Supports: &SupportsNot{
							Condition: &SupportsDecleration{
								Key:   []byte("display"),
								Value: []Value{&BasicValue{Value: []byte("flex")}},
							},
						},
					},
				},
			},
		},
		{
			name:  "@import with complex supports, layer, and media query",
			input: `@import url("complex.css") layer(utilities) supports((display: flex) and (not (color: green))) screen and (min-width: 1024px);`,
			expected: &Stylesheet{
				Rules: []Node{
					&ImportAtRule{
						URL: &FunctionValue{
							Name: []byte("url"),
							Arguments: []Value{
								&StringValue{SingleQuote: false, Value: []byte("complex.css")},
							},
						},
						Layer: &FunctionValue{
							Name: []byte("layer"),
							Arguments: []Value{
								&BasicValue{Value: []byte("utilities")},
							},
						},
						Supports: &SupportsGroup{
							Conditions: []SupportsCondition{

								&SupportsDecleration{
									Key:   []byte("display"),
									Value: []Value{&BasicValue{Value: []byte("flex")}},
								},
								&SupportsOperator{Operator: "and"},
								&SupportsNot{
									Condition: &SupportsDecleration{
										Key:   []byte("color"),
										Value: []Value{&BasicValue{Value: []byte("green")}},
									},
								},
							},
						},
						Media: MediaQuery{
							Queries: []MediaQueryExpression{
								{
									MediaType: []byte("screen"),
									Features: []MediaFeature{
										{Name: []byte("min-width"), Value: []byte("1024px")},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "@import with multiple support groups",
			input: `@import url("complex.css") supports((display: flex) and ((color: rebeccapurple) or (transform: rotate(45deg))));`,
			expected: &Stylesheet{
				Rules: []Node{
					&ImportAtRule{
						URL: &FunctionValue{
							Name: []byte("url"),
							Arguments: []Value{
								&StringValue{SingleQuote: false, Value: []byte("complex.css")},
							},
						},
						Supports: &SupportsGroup{
							Conditions: []SupportsCondition{
								&SupportsDecleration{
									Key:   []byte("display"),
									Value: []Value{&BasicValue{Value: []byte("flex")}},
								},
								&SupportsOperator{Operator: "and"},
								&SupportsGroup{
									Conditions: []SupportsCondition{
										&SupportsDecleration{
											Key:   []byte("color"),
											Value: []Value{&BasicValue{Value: []byte("rebeccapurple")}},
										},
										&SupportsOperator{Operator: "or"},
										&SupportsDecleration{
											Key: []byte("transform"),
											Value: []Value{
												&FunctionValue{
													Name: []byte("rotate"),
													Arguments: []Value{
														&BasicValue{Value: []byte("45deg")},
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
		},
	}

	runTests(t, tests)
}
