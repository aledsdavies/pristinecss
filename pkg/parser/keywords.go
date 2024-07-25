package parser

import "bytes"

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
