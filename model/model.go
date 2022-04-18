package model

import "github.com/catlev/pkg/store/block"

type EntityModel struct {
	Types []Type
}

type Type struct {
	ID   block.Word
	Name string
	Kind TypeKind
	Rels []Relationship
}

type TypeKind int

const (
	Absolute TypeKind = iota
	Value
	Entity
)

type Relationship struct {
	Name        string
	Type        block.Word
	Kind        RelKind
	Identifying bool
	Column      int
}

type RelKind int

const (
	AbsoluteID block.Word = iota
	IntegerID
)
