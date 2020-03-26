package types

import (
	"github.com/catlev/pkg/path/syntax"
)

type Path struct {
	Expr         syntax.Tree
	Alternatives []Alternative
}

type Alternative struct {
	Source, Target Type
}

type Kind int

const (
	Absolute Kind = iota
	Attribute
	Entity
)

type Type struct {
	Kind Kind
	Name string
}

type Model interface {
	Lookup(name string) (Path, error)
}

var (
	absoluteType  = Type{Kind: Absolute, Name: "^"}
	attributeType = Type{Kind: Attribute, Name: "$"}
)
