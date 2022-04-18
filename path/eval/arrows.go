package eval

import (
	"sort"

	"github.com/catlev/pkg/model"
	"github.com/catlev/pkg/store/block"
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
	entityID block.Word
	valueID  block.Word
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
			res.arms = append(res.arms, arm{
				entityID: p.valueID,
				where: []clause{{
					condition: equal,
					value:     c.This().Fields[p.column],
				}},
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
	entityID block.Word
	valueID  block.Word
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
		where := make([]clause, len(xs.model.Types[p.entityID].Rels))
		where[p.column] = clause{
			condition: equal,
			value:     a.where[0].value,
		}
		res.arms = append(res.arms, arm{
			entityID: p.entityID,
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

type valuePath struct {
	valueID block.Word
	value   string
}

func (p *valuePath) Follow(xs Box) Box {
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
		res.arms = append(res.arms, arm{
			entityID: p.valueID,
			where: []clause{{
				condition: equal,
				value:     value,
			}},
		})
	}
	return res
}

func (p *valuePath) Reverse() Arrow {
	return &valueFilter{
		valueID: p.valueID,
		value:   p.value,
	}
}

type valueFilter struct {
	valueID block.Word
	value   string
}

func (p *valueFilter) Follow(xs Box) Box {
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
		return vs[i].where[0].value >= value
	}); idx == len(vs) || vs[idx].where[0].value != value {
		return res
	}
	res.arms = append(res.arms, arm{
		entityID: model.AbsoluteID,
	})
	return res
}

func (p *valueFilter) Reverse() Arrow {
	return &valuePath{
		valueID: p.valueID,
		value:   p.value,
	}
}

type entityPath struct {
	entityID block.Word
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
		res.arms = append(res.arms, arm{
			entityID: p.entityID,
			where:    make([]clause, len(xs.model.Types[p.entityID].Rels)),
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
	entityID block.Word
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
		res.arms = append(res.arms, arm{
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
