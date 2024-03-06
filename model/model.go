package model

import (
	"errors"
	"fmt"
	"io"

	"github.com/catlev/pkg/domain"
	"github.com/catlev/pkg/path"
	"github.com/catlev/pkg/path/syntax"
	"github.com/catlev/pkg/stream"
)

type EntityModel struct {
	Types Types
}

type Types []Type

type Type struct {
	ID            domain.Word
	Name          string
	Kind          TypeKind
	Attributes    Attributes
	Relationships Relationships
}

type TypeKind int

const (
	Absolute TypeKind = iota
	Value
	Entity
)

type Attributes []Attribute

type Attribute struct {
	Name        string
	Identifying bool
	Type        domain.Word
}

type Relationships []Relationship

type Relationship struct {
	Name string
	Impl path.Expr
}

const (
	AbsoluteID domain.Word = iota
	IntegerID
	StringID
)

var (
	ErrUnknownElement = errors.New("unknown element")
	ErrUnknownType    = errors.New("unknown type")
)

func Read(src io.Reader) (*EntityModel, error) {
	r := stream.NewReader(src)
	m := new(EntityModel)
	for r.Next() {
		if r.Name() != "entity_type" {
			return nil, fmt.Errorf("%s: %w", r.Name(), ErrUnknownElement)
		}
		err := m.Types.read(r.Record())
		if err != nil {
			return nil, err
		}
	}
	return m, r.Err()
}

func (ts *Types) read(r *stream.Reader) error {
	var t Type
	for r.Next() {
		var err error
		switch r.Name() {
		case "name":
			t.Name = r.StringField()
		case "attribute":
			err = t.Attributes.read(r.Record())
		case "relationship":
			err = t.Relationships.read(r.Record())
		default:
			err = fmt.Errorf("%s: %w", r.Name(), ErrUnknownElement)
		}
		if err != nil {
			return err
		}
	}
	*ts = append(*ts, t)
	return nil
}

func (as *Attributes) read(r *stream.Reader) error {
	var a Attribute
	for r.Next() {
		switch r.Name() {
		case "name":
			a.Name = r.StringField()
		case "identifying":
			a.Identifying = r.BoolField()
		case "type":
			id, err := as.parseAttributeType(r.StringField())
			if err != nil {
				return err
			}
			a.Type = id
		default:
			return fmt.Errorf("%s: %w", r.Name(), ErrUnknownElement)
		}
	}
	*as = append(*as, a)
	return nil
}

func (as Attributes) parseAttributeType(name string) (domain.Word, error) {
	switch name {
	case "integer":
		return IntegerID, nil
	default:
		return 0, fmt.Errorf("%s: %w", name, ErrUnknownType)
	}
}

func (rs *Relationships) read(r *stream.Reader) error {
	var a Relationship
	for r.Next() {
		switch r.Name() {
		case "name":
			a.Name = r.StringField()
		case "impl":
			impl, err := syntax.ParseString(r.StringField())
			if err != nil {
				return err
			}
			a.Impl = impl
		default:
			return fmt.Errorf("%s: %w", r.Name(), ErrUnknownElement)
		}
	}
	*rs = append(*rs, a)
	return nil
}
