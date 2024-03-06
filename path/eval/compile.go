package eval

import (
	"errors"

	"github.com/catlev/pkg/domain"
	"github.com/catlev/pkg/model"
	"github.com/catlev/pkg/path"
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

func (c *Compiler) Compile(tree path.Expr) (Arrow, error) {
	return path.Visit[Arrow](&compileVisitor{*c}, tree)
}

type compileVisitor struct {
	Compiler
}

func success[T any](x T) (T, error) {
	return x, nil
}

func (v *compileVisitor) String(x string) (Arrow, error) {
	return success(&stringPath{
		valueID: model.StringID,
		value:   x,
	})
}

func (v *compileVisitor) Integer(x int) (Arrow, error) {
	return success(&intPath{
		valueID: model.IntegerID,
		value:   domain.Word(x),
	})
}

func (v *compileVisitor) Rel(name string) (Arrow, error) {
	var res Arrow

	for _, t := range v.model.Types {
		if t.Name == name {
			res = addOption(res, &entityPath{
				entityID: t.ID,
			})
		}
		for i, c := range t.Attributes {
			if c.Name == name {
				res = addOption(res, &attrPath{
					entityID: t.ID,
					valueID:  c.Type,
					column:   i,
				})
			}
		}
	}

	if res == nil {
		return nil, ErrUnknownTerm
	}

	return res, nil

}

func (v *compileVisitor) Op(name string, args []Arrow) (Arrow, error) {
	op := v.ops[name]
	if op == nil {
		return nil, ErrUnknownOp
	}
	return op(args)
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
