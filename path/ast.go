package path

// Kind is the kind of the path node.
type Kind byte

const (
	// Invalid - Not a valid AST.
	Invalid Kind = iota

	// String value.
	String

	// Integer value.
	Integer

	// Rel name.
	Rel

	// Op - name of an operation (join, inverse, intersection etc).
	Op
)

// Expr is the result of parsing a path expression.
type Expr struct {
	Kind     Kind
	Value    string
	Children []Expr
}
