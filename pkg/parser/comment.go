package parser

import "fmt"

const (
	NodeComment NodeType = "Comment"
)

func init() {
	RegisterNodeType(NodeComment, visitComment)
}

var _ Node = (*Comment)(nil)


type Comment struct {
    Text []byte
}

func (c *Comment) Type() NodeType { return NodeComment }
func (c *Comment) String() string {
    return fmt.Sprintf("Comment{Text: %q}", string(c.Text))
}

func visitComment(pv *ParseVisitor, node Node) {
    // The comment's text has already been set when the node was created,
    // so we just need to advance past the comment token.
    pv.advance()
}
