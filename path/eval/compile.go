package eval

import (
	"errors"

	"github.com/catlev/pkg/model"
	"github.com/catlev/pkg/path/syntax"
)

type Compiler struct {
	model *model.EntityModel
}

func NewCompiler(m *model.EntityModel) *Compiler {
	return &Compiler{model: m}
}

var (
	ErrUnsupportedSyntax = errors.New("unsupported syntax")
	ErrUnknownTerm       = errors.New("unknown term")
)

func (c *Compiler) Compile(tree syntax.Tree) (Arrow, error) {
	switch tree.Kind {
	case syntax.Integer:
		return &valuePath{
			valueID: model.IntegerID,
			value:   tree.Value,
		}, nil

	case syntax.Term:
		return c.compileTerm(tree.Value)

	case syntax.Inverse:
		inner, err := c.Compile(tree.Children[0])
		if err != nil {
			return nil, err
		}
		return inner.Reverse(), nil

	case syntax.Union:
		left, err := c.Compile(tree.Children[0])
		if err != nil {
			return nil, err
		}
		right, err := c.Compile(tree.Children[1])
		if err != nil {
			return nil, err
		}

		return &unionPath{
			left:  left,
			right: right,
		}, nil

	case syntax.Intersection:
		left, err := c.Compile(tree.Children[0])
		if err != nil {
			return nil, err
		}
		right, err := c.Compile(tree.Children[1])
		if err != nil {
			return nil, err
		}

		return &intersectionPath{
			left:  left,
			right: right,
		}, nil

	case syntax.Join:
		left, err := c.Compile(tree.Children[0])
		if err != nil {
			return nil, err
		}
		right, err := c.Compile(tree.Children[1])
		if err != nil {
			return nil, err
		}

		return &joinPath{
			left:  left,
			right: right,
		}, nil
	}
	return nil, ErrUnsupportedSyntax
}

func (c *Compiler) compileTerm(name string) (Arrow, error) {
	var res Arrow

	for _, t := range c.model.Types {
		if t.Name == name {
			res = addOption(res, &entityPath{
				entityID: t.ID,
			})
		}
		for i, c := range t.Rels {
			if c.Name != name {
				continue
			}
			res = addOption(res, &attrPath{
				entityID: t.ID,
				valueID:  c.Type,
				column:   i,
			})
		}
	}

	if res == nil {
		return nil, ErrUnknownTerm
	}

	return res, nil
}

func addOption(left, right Arrow) Arrow {
	if left == nil {
		return right
	}

	return &unionPath{
		left:  left,
		right: right,
	}
}
