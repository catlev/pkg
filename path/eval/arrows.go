package eval

import (
	"sort"

	"github.com/catlev/pkg/domain"
	"github.com/catlev/pkg/model"
)

type Arrow interface {
	Follow(xs Box) Box
	Reverse() Arrow
}

type unionPath struct {
	left, right Arrow
}

func (p *unionPath) Follow(xs Box) Box {
	if xs.err != nil {
		return Box{err: xs.err}
	}

	left := p.left.Follow(xs)
	right := p.right.Follow(xs)

	return left.Union(right)
}

func (p *unionPath) Reverse() Arrow {
	return &unionPath{
		left:  p.left.Reverse(),
		right: p.right.Reverse(),
	}
}

type intersectionPath struct {
	left, right Arrow
}

func (p *intersectionPath) Follow(xs Box) Box {
	if xs.err != nil {
		return Box{err: xs.err}
	}

	left := p.left.Follow(xs)
	right := p.right.Follow(xs)

	return left.Intersection(right)
}

func (p *intersectionPath) Reverse() Arrow {
	return &intersectionPath{
		left:  p.left.Reverse(),
		right: p.right.Reverse(),
	}
}

type joinPath struct {
	left, right Arrow
}

func (p *joinPath) Follow(xs Box) Box {
	if xs.err != nil {
		return Box{err: xs.err}
	}

	return p.right.Follow(p.left.Follow(xs))
}

func (p *joinPath) Reverse() Arrow {
	return &joinPath{
		left:  p.right.Reverse(),
		right: p.left.Reverse(),
	}
}

type attrPath struct {
	entityID domain.Word
	valueID  domain.Word
	column   int
}

func (p *attrPath) Follow(xs Box) Box {
	if xs.err != nil {
		return Box{err: xs.err}
	}

	res := Box{
		store: xs.store,
		model: xs.model,
	}
	for _, a := range xs.findAll(p.entityID) {
		c := xs.runQuery(a)
		for c.Next() {
			res.contents = append(res.contents, Query{
				entityID: p.valueID,
				mask:     1,
				where:    []domain.Word{c.This().Fields[p.column]},
			})
		}
	}
	res.simplify()
	return res
}

func (p *attrPath) Reverse() Arrow {
	return &attrFilter{
		entityID: p.entityID,
		valueID:  p.valueID,
		column:   p.column,
	}
}

type attrFilter struct {
	entityID domain.Word
	valueID  domain.Word
	column   int
}

func (p *attrFilter) Follow(xs Box) Box {
	if xs.err != nil {
		return Box{err: xs.err}
	}

	res := Box{
		store: xs.store,
		model: xs.model,
	}
	for _, a := range xs.findAll(p.valueID) {
		where := make([]domain.Word, len(xs.model.Types[p.entityID].Attributes))
		where[p.column] = a.where[0]

		res.contents = append(res.contents, Query{
			entityID: p.entityID,
			mask:     1 << p.column,
			where:    where,
		})
	}
	res.simplify()
	return res
}

func (p *attrFilter) Reverse() Arrow {
	return &attrPath{
		entityID: p.entityID,
		valueID:  p.valueID,
		column:   p.column,
	}
}

type intPath struct {
	valueID domain.Word
	value   domain.Word
}

func (p *intPath) Follow(xs Box) Box {
	if xs.err != nil {
		return Box{err: xs.err}
	}

	res := Box{
		store: xs.store,
		model: xs.model,
	}
	if len(xs.findAll(model.AbsoluteID)) != 0 {
		res.contents = append(res.contents, Query{
			entityID: p.valueID,
			mask:     1,
			where:    []domain.Word{p.value},
		})
	}

	return res
}

func (p *intPath) Reverse() Arrow {
	return &intFilter{p.valueID, p.value}
}

type intFilter struct {
	valueID domain.Word
	value   domain.Word
}

func (p *intFilter) Follow(xs Box) Box {
	if xs.err != nil {
		return Box{err: xs.err}
	}

	res := Box{
		store: xs.store,
		model: xs.model,
	}
	vs := xs.findAll(p.valueID)
	if idx := sort.Search(len(vs), func(i int) bool {
		return vs[i].where[0] >= p.value
	}); idx == len(vs) || vs[idx].where[0] != p.value {
		return res
	}
	res.contents = append(res.contents, Query{
		entityID: model.AbsoluteID,
	})
	return res
}

func (p *intFilter) Reverse() Arrow {
	return &intPath{p.valueID, p.value}
}

type stringPath struct {
	valueID domain.Word
	value   string
}

func (p *stringPath) Follow(xs Box) Box {
	if xs.err != nil {
		return Box{err: xs.err}
	}

	value, err := xs.store.ParseValue(p.valueID, p.value)
	if err != nil {
		return Box{err: err}
	}
	res := Box{
		store: xs.store,
		model: xs.model,
	}
	if len(xs.findAll(model.AbsoluteID)) != 0 {
		res.contents = append(res.contents, Query{
			entityID: p.valueID,
			mask:     1,
			where:    []domain.Word{value},
		})
	}
	return res
}

func (p *stringPath) Reverse() Arrow {
	return &stringFilter{
		valueID: p.valueID,
		value:   p.value,
	}
}

type stringFilter struct {
	valueID domain.Word
	value   string
}

func (p *stringFilter) Follow(xs Box) Box {
	if xs.err != nil {
		return Box{err: xs.err}
	}

	value, err := xs.store.ParseValue(p.valueID, p.value)
	if err != nil {
		return Box{err: err}
	}
	res := Box{
		store: xs.store,
		model: xs.model,
	}
	vs := xs.findAll(p.valueID)
	if idx := sort.Search(len(vs), func(i int) bool {
		return vs[i].where[0] >= value
	}); idx == len(vs) || vs[idx].where[0] != value {
		return res
	}
	res.contents = append(res.contents, Query{
		entityID: model.AbsoluteID,
	})
	return res
}

func (p *stringFilter) Reverse() Arrow {
	return &stringPath{
		valueID: p.valueID,
		value:   p.value,
	}
}

type entityPath struct {
	entityID domain.Word
}

func (p *entityPath) Follow(xs Box) Box {
	if xs.err != nil {
		return Box{err: xs.err}
	}

	res := Box{
		store: xs.store,
		model: xs.model,
	}
	if len(xs.findAll(model.AbsoluteID)) != 0 {
		res.contents = append(res.contents, Query{
			entityID: p.entityID,
			where:    make([]domain.Word, len(xs.model.Types[p.entityID].Attributes)),
		})
	}
	return res
}

func (p *entityPath) Reverse() Arrow {
	return &entityFilter{
		entityID: p.entityID,
	}
}

type entityFilter struct {
	entityID domain.Word
}

func (p *entityFilter) Follow(xs Box) Box {
	if xs.err != nil {
		return Box{err: xs.err}
	}

	res := Box{
		store: xs.store,
		model: xs.model,
	}
	if len(xs.findAll(p.entityID)) != 0 {
		res.contents = append(res.contents, Query{
			entityID: model.AbsoluteID,
		})
	}
	return res
}

func (p *entityFilter) Reverse() Arrow {
	return &entityPath{
		entityID: p.entityID,
	}
}
