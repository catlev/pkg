package eval

import (
	"errors"

	"github.com/catlev/pkg/model"
	"github.com/catlev/pkg/path/syntax"
)

type Compiler struct {
	model *model.EntityModel
	ops   map[string]OpSpec
}

type OpSpec func(args []Arrow) (Arrow, error)

var StandardOps = map[string]OpSpec{
	"inverse": func(args []Arrow) (Arrow, error) {
		return args[0].Reverse(), nil
	},
	"union": func(args []Arrow) (Arrow, error) {
		return &unionPath{left: args[0], right: args[1]}, nil
	},
	"intersection": func(args []Arrow) (Arrow, error) {
		return &intersectionPath{left: args[0], right: args[1]}, nil
	},
	"join": func(args []Arrow) (Arrow, error) {
		return &joinPath{left: args[0], right: args[1]}, nil
	},
}

func NewCompiler(m *model.EntityModel, ops map[string]OpSpec) *Compiler {
	return &Compiler{model: m, ops: ops}
}

var (
	ErrUnsupportedSyntax = errors.New("unsupported syntax")
	ErrUnknownTerm       = errors.New("unknown term")
	ErrUnknownOp         = errors.New("unknown op")
)

func (c *Compiler) Compile(tree syntax.Tree) (Arrow, error) {
	switch tree.Kind {
	case syntax.Integer:
		return &valuePath{
			valueID: model.IntegerID,
			value:   tree.Value,
		}, nil

	case syntax.Rel:
		return c.compileTerm(tree.Value)

	case syntax.Op:
		op := c.ops[tree.Value]
		if op == nil {
			return nil, ErrUnknownOp
		}
		args := make([]Arrow, len(tree.Children))
		for i, p := range tree.Children {
			a, err := c.Compile(p)
			if err != nil {
				return nil, err
			}
			args[i] = a
		}
		return op(args)
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
		for i, c := range t.Attributes {
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
