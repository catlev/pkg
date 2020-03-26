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

	// Value - literal value.
	Value

	// Term - name of a model element.
	Term

	// Inverse - a path with its direction reversed.
	Inverse

	// Join - two paths connected serially.
	Join

	// Intersection - set operation on paths.
	Intersection

	// Union - set operation on paths.
	Union
)

// A Tree is the result of parsing a path expression.
type Tree struct {
	Kind     Kind
	Value    string
	Children []Tree
}
