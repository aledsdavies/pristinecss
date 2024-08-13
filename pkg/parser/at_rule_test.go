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
	}

	runTests(t, tests)
}

func TestFontFaceAtRule(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *Stylesheet
	}{
		{
			name: "Basic @font-face rule",
			input: `@font-face {
                font-family: "Open Sans";
                src: url("/fonts/OpenSans-Regular-webfont.woff2") format("woff2");
            }`,
			expected: &Stylesheet{
				Rules: []Node{
					&FontFaceAtRule{
						Declarations: []Declaration{
							{Key: []byte("font-family"), Value: []Value{
								&StringValue{Value: []byte("Open Sans")},
							}},
							{Key: []byte("src"), Value: []Value{
								&FunctionValue{
									Name: []byte("url"),
									Arguments: []Value{
										&StringValue{Value: []byte("/fonts/OpenSans-Regular-webfont.woff2")},
									},
								},
								&FunctionValue{
									Name: []byte("format"),
									Arguments: []Value{
										&StringValue{Value: []byte("woff2")},
									},
								},
							}},
						},
					},
				},
			},
		},
		{
			name: "@font-face with multiple src",
			input: `@font-face {
                font-family: "Bitstream Vera Serif Bold";
                src: url("https://mdn.mozillademos.org/files/2468/VeraSeBd.ttf");
                src: local("Bitstream Vera Serif Bold"),
                     local("BitstreamVeraSerif-Bold"),
                     url("VeraSeBd.ttf") format("truetype");
            }`,
			expected: &Stylesheet{
				Rules: []Node{
					&FontFaceAtRule{
						Declarations: []Declaration{
							{Key: []byte("font-family"), Value: []Value{
								&StringValue{Value: []byte("Bitstream Vera Serif Bold")},
							}},
							{Key: []byte("src"), Value: []Value{
								&FunctionValue{
									Name: []byte("url"),
									Arguments: []Value{
										&StringValue{Value: []byte("https://mdn.mozillademos.org/files/2468/VeraSeBd.ttf")},
									},
								},
							}},
							{Key: []byte("src"), Value: []Value{
								&FunctionValue{
									Name: []byte("local"),
									Arguments: []Value{
										&StringValue{Value: []byte("Bitstream Vera Serif Bold")},
									},
								},
								&FunctionValue{
									Name: []byte("local"),
									Arguments: []Value{
										&StringValue{Value: []byte("BitstreamVeraSerif-Bold")},
									},
								},
								&FunctionValue{
									Name: []byte("url"),
									Arguments: []Value{
										&StringValue{Value: []byte("VeraSeBd.ttf")},
									},
								},
								&FunctionValue{
									Name: []byte("format"),
									Arguments: []Value{
										&StringValue{Value: []byte("truetype")},
									},
								},
							}},
						},
					},
				},
			},
		},
		{
			name: "@font-face with additional properties",
			input: `@font-face {
                font-family: "Roboto";
                src: url("Roboto-Regular.woff2") format("woff2");
                font-weight: 400;
                font-style: normal;
                font-display: swap;
            }`,
			expected: &Stylesheet{
				Rules: []Node{
					&FontFaceAtRule{
						Declarations: []Declaration{
							{Key: []byte("font-family"), Value: []Value{
								&StringValue{Value: []byte("Roboto")},
							}},
							{Key: []byte("src"), Value: []Value{
								&FunctionValue{
									Name: []byte("url"),
									Arguments: []Value{
										&StringValue{Value: []byte("Roboto-Regular.woff2")},
									},
								},
								&FunctionValue{
									Name: []byte("format"),
									Arguments: []Value{
										&StringValue{Value: []byte("woff2")},
									},
								},
							}},
							{Key: []byte("font-weight"), Value: []Value{
								&BasicValue{Value: []byte("400")},
							}},
							{Key: []byte("font-style"), Value: []Value{
								&BasicValue{Value: []byte("normal")},
							}},
							{Key: []byte("font-display"), Value: []Value{
								&BasicValue{Value: []byte("swap")},
							}},
						},
					},
				},
			},
		},
	}
	runTests(t, tests)
}

func TestCharsetAtRule(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *Stylesheet
	}{
		{
			name:  "Simple @charset rule",
			input: `@charset "UTF-8";`,
			expected: &Stylesheet{
				Rules: []Node{
					&CharsetAtRule{
						Charset: &StringValue{Value: []byte("UTF-8")},
					},
				},
			},
		},
		{
			name: "@charset rule at the beginning of a stylesheet",
			input: `@charset "UTF-8";
                    body { font-family: Arial, sans-serif; }`,
			expected: &Stylesheet{
				Rules: []Node{
					&CharsetAtRule{
						Charset: &StringValue{Value: []byte("UTF-8")},
					},
					&Selector{
						Selectors: []SelectorValue{{Type: Element, Value: []byte("body")}},
						Rules: []Node{
							&Declaration{Key: []byte("font-family"), Value: []Value{
								&BasicValue{Value: []byte("Arial")},
								&BasicValue{Value: []byte("sans-serif")},
							}},
						},
					},
				},
			},
		},
	}

	runTests(t, tests)
}
