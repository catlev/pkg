package eval

import (
	"github.com/catlev/pkg/model"
	"github.com/catlev/pkg/path/syntax"
)

type Host struct {
	model *model.EntityModel
	store Store

	compiler *Compiler
}

func NewHost(m *model.EntityModel, s Store) *Host {
	c := NewCompiler(m, StandardOps)
	return &Host{
		model: m,
		store: s,

		compiler: c,
	}
}

func (h *Host) Absolute() Box {
	return Box{
		model: h.model,
		store: h.store,
		contents: []Query{{
			entityID: model.AbsoluteID,
		}},
	}
}

func (h *Host) Eval(start Box, path string) Box {
	if start.err != nil {
		return Box{err: start.err}
	}

	tree, err := syntax.ParseString(path)
	if err != nil {
		return Box{err: err}
	}

	arrow, err := h.compiler.Compile(tree)
	if err != nil {
		return Box{err: err}
	}

	return arrow.Follow(start)
}
