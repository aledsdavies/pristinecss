package parser

type StyleSheet struct {
	Rules []Node
}

type CSSOption func(*Selector) *Selector

type Node interface {
	ToCSS(options ...CSSOption) string
}

// SelectorType defines the type of CSS selector.
type SelectorType string

// Constants for different types of selectors.
const (
	TypeElement   SelectorType = "element"
	TypeClass     SelectorType = "class"
	TypeID        SelectorType = "id"
	TypeAttribute SelectorType = "attribute"
)


var _ Node = (*Selector)(nil)

// Selector represents a CSS selector and its associated styles.
type Selector struct {
	Type         SelectorType
	Name         string
	Declarations map[string]string
}

func (s *Selector) ToCSS(options ...CSSOption) string {
    return ""
}
