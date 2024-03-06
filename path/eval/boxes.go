package eval

import (
	"sort"

	"github.com/catlev/pkg/domain"
	"github.com/catlev/pkg/model"
)

type Store interface {
	FindEntities(entityID domain.Word, key []domain.Word) Cursor
	ParseValue(valueID domain.Word, value string) (domain.Word, error)
}

type Object struct {
	EntityID domain.Word
	Fields   []domain.Word
}

type Box struct {
	store    Store
	model    *model.EntityModel
	contents []Query
	err      error
}

type Query struct {
	entityID domain.Word
	mask     uint64
	where    []domain.Word
}

func (s Box) Enumerate() Cursor {
	if s.err != nil {
		return &errCursor{s.err}
	}
	return &boxCursor{
		cur: &errCursor{nil},
		arm: -1,
		box: s,
	}
}

func (s Box) Union(t Box) Box {
	if s.err != nil {
		return Box{err: s.err}
	}

	res := Box{
		store:    s.store,
		model:    s.model,
		contents: make([]Query, len(s.contents)+len(t.contents)),
	}

	copy(res.contents, s.contents)
	copy(res.contents[len(s.contents):], t.contents)
	res.simplify()

	return res
}

func (s Box) Intersection(t Box) Box {
	if s.err != nil {
		return Box{err: s.err}
	}

	res := Box{
		store: s.store,
		model: s.model,
	}

	if len(s.contents) == 0 {
		return res
	}

	var bs []Query

	for _, a := range s.contents {
		if len(bs) == 0 || bs[0].entityID != a.entityID {
			bs = t.findAll(a.entityID)
		}
		if len(bs) == 0 {
			continue
		}
		for _, b := range bs {
			merged, ok := a.merge(b)
			if ok {
				res.contents = append(res.contents, merged)
			}
		}
	}

	res.simplify()
	return res
}

func (s Box) findAll(entityID domain.Word) []Query {
	start := sort.Search(len(s.contents), func(i int) bool {
		return s.contents[i].entityID >= entityID
	})
	end := sort.Search(len(s.contents), func(i int) bool {
		return s.contents[i].entityID > entityID
	})
	return s.contents[start:end]
}

func (s Box) runQuery(a Query) Cursor {
	if s.model.Types[a.entityID].Kind == model.Value {
		return &valueCursor{
			arm: a,
			pos: -1,
		}
	}
	base := s.store.FindEntities(a.entityID, s.buildKey(a))
	return &whereCursor{
		base: base,
		arm:  a,
	}
}

func (s Box) buildKey(a Query) []domain.Word {
	t := s.model.Types[a.entityID]
	var key []domain.Word
	for i, c := range t.Attributes {
		if !c.Identifying {
			break
		}
		if !a.constrains(i) {
			break
		}
		key = append(key, a.where[i])
	}
	return key
}

func (s *Box) simplify() {
	sort.Slice(s.contents, func(i, j int) bool {
		a := s.contents[i]
		b := s.contents[j]

		return a.before(b)
	})

	dupMap := make([]bool, len(s.contents))
	dupsExist := false
	for i, a := range s.contents {
		if dupMap[i] {
			continue
		}
		for j, b := range s.contents[i+1:] {
			j += i + 1
			if a.entityID != b.entityID {
				break
			}
			if dupMap[j] {
				continue
			}
			if a.supersetOf(b) {
				dupMap[j] = true
				dupsExist = true
			}
		}
	}

	if !dupsExist {
		return
	}
	var arms []Query
	for i, a := range s.contents {
		if !dupMap[i] {
			arms = append(arms, a)
		}
	}
	s.contents = arms
}

func (a Query) constrains(column int) bool {
	return a.mask&(1<<column) != 0
}

func (a Query) supersetOf(b Query) bool {
	if a.entityID != b.entityID {
		return false
	}
	for i, p := range a.where {
		q := b.where[i]

		if !a.constrains(i) {
			continue
		}
		if p == q {
			continue
		}

		return false
	}
	return true
}

func (a Query) before(b Query) bool {
	if a.entityID != b.entityID {
		return a.entityID < b.entityID
	}

	if a.mask != b.mask {
		return a.mask < b.mask
	}

	for i, a := range a.where {
		b := b.where[i]
		if a != b {
			return a < b
		}
	}

	return false
}

func (a Query) merge(b Query) (Query, bool) {
	if a.entityID != b.entityID {
		return Query{}, false
	}
	res := make([]domain.Word, len(a.where))
	for i, p := range a.where {
		q := b.where[i]
		if !a.constrains(i) {
			res[i] = q
			continue
		}
		if !b.constrains(i) {
			res[i] = p
			continue
		}
		if p == q {
			res[i] = p
			continue
		}
		return Query{}, false
	}
	return Query{a.entityID, a.mask | b.mask, res}, true
}
