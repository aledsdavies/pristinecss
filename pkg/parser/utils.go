package parser

import (
	"bytes"
	"fmt"
	"strings"
)

func selectorTypeToString(st SelectorType) string {
	switch st {
	case Element:
		return "Element"
	case Class:
		return "Class"
	case ID:
		return "ID"
	case Attribute:
		return "Attribute"
	case Pseudo:
		return "Pseudo"
	case Combinator:
		return "Combinator"
	default:
		return fmt.Sprintf("Unknown(%d)", st)
	}
}

func indentLines(s string, spaces int) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		if i > 0 {
			lines[i] = strings.Repeat(" ", spaces) + line
		}
	}
	return strings.Join(lines, "\n")
}

var units = [][]byte{
	// Absolute length units
	[]byte("cm"), []byte("mm"), []byte("in"), []byte("px"), []byte("pt"), []byte("pc"), []byte("Q"),

	// Relative length units
	[]byte("em"), []byte("ex"), []byte("ch"), []byte("rem"), []byte("lh"), []byte("rlh"), []byte("vb"), []byte("vi"),

	// Viewport-percentage lengths
	[]byte("vw"), []byte("vh"), []byte("vmin"), []byte("vmax"),
	[]byte("svw"), []byte("svh"), []byte("lvw"), []byte("lvh"),
	[]byte("dvw"), []byte("dvh"), []byte("vi"), []byte("vb"),

	// Container query length units
	[]byte("cqw"), []byte("cqh"), []byte("cqi"), []byte("cqb"), []byte("cqmin"), []byte("cqmax"),

	// Percentage (handled separately in most cases, but included for completeness)
	[]byte("%"),

	// Angle units
	[]byte("deg"), []byte("grad"), []byte("rad"), []byte("turn"),

	// Time units
	[]byte("s"), []byte("ms"),

	// Frequency units
	[]byte("Hz"), []byte("kHz"),

	// Resolution units
	[]byte("dpi"), []byte("dpcm"), []byte("dppx"),

	// Flex units
	[]byte("fr"),
}

func isUnit(literal []byte) bool {
	for _, unit := range units {
		if bytes.Equal(literal, unit) {
			return true
		}
	}
	return false
}
