package syntax

import (
	"errors"
)

// ErrFailedToMatch indicates that there is a syntax error in the path.
var ErrFailedToMatch = errors.New("failed to parse path")

// Kind is the kind of the path node.
type Kind byte

const (
	// Invalid - Not a valid AST
	Invalid Kind = iota

	// String value
	String

	// Integer value
	Integer

	// Rel name.
	Rel

	// Op - name of an operation (join, and, or, reverse).
	Op
)

// A Tree is the result of parsing a path expression.
type Tree struct {
	Kind     Kind
	Value    string
	Children []Tree
}
