package eval

import (
	"sort"

	"github.com/catlev/pkg/model"
	"github.com/catlev/pkg/store/block"
)

type Store interface {
	FindEntities(entityID block.Word, key []block.Word) Cursor
	ParseValue(valueID block.Word, value string) (block.Word, error)
}

type Object struct {
	EntityID block.Word
	Fields   []block.Word
}

type Cursor interface {
	Next() bool
	This() Object
	Err() error
}

type Box struct {
	store Store
	model *model.EntityModel
	arms  []arm
	err   error
}

type arm struct {
	entityID block.Word
	mask     uint64
	where    []block.Word
}

type errCursor struct {
	err error
}

func (c *errCursor) Err() error {
	return c.err
}

func (*errCursor) Next() bool {
	return false
}

func (*errCursor) This() Object {
	panic("unimplemented")
}

type whereCursor struct {
	base Cursor
	arm  arm
}

func (c *whereCursor) Err() error {
	return c.base.Err()
}

func (c *whereCursor) Next() bool {
	for {
		if !c.base.Next() {
			return false
		}
		o := c.base.This()
		if o.EntityID != c.arm.entityID {
			return false
		}
		if c.rowMatches(o) {
			return true
		}
	}
}

func (c *whereCursor) rowMatches(o Object) bool {
	for i, p := range c.arm.where {
		if !c.arm.constrains(i) {
			continue
		}
		if o.Fields[i] != p {
			return false
		}
	}
	return true
}

func (c *whereCursor) This() Object {
	return c.base.This()
}

type valueCursor struct {
	arm arm
	pos int
}

func (c *valueCursor) Err() error {
	return nil
}

func (c *valueCursor) Next() bool {
	c.pos++
	return c.pos < len(c.arm.where)
}

func (c *valueCursor) This() Object {
	return Object{
		EntityID: c.arm.entityID,
		Fields:   []block.Word{c.arm.where[0]},
	}
}

type boxCursor struct {
	cur Cursor
	arm int
	box Box
}

// Err implements Cursor
func (c *boxCursor) Err() error {
	return c.cur.Err()
}

// Next implements Cursor
func (c *boxCursor) Next() bool {
	for !c.cur.Next() {
		c.arm++
		if c.arm >= len(c.box.arms) {
			return false
		}
		c.cur = c.box.runQuery(c.box.arms[c.arm])
	}
	return true
}

// This implements Cursor
func (c *boxCursor) This() Object {
	return c.cur.This()
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
		store: s.store,
		model: s.model,
		arms:  make([]arm, len(s.arms)+len(t.arms)),
	}

	copy(res.arms, s.arms)
	copy(res.arms[len(s.arms):], t.arms)
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

	if len(s.arms) == 0 {
		return res
	}

	var bs []arm

	for _, a := range s.arms {
		if len(bs) == 0 || bs[0].entityID != a.entityID {
			bs = t.findAll(a.entityID)
		}
		if len(bs) == 0 {
			continue
		}
		for _, b := range bs {
			merged, ok := a.merge(b)
			if ok {
				res.arms = append(res.arms, merged)
			}
		}
	}

	res.simplify()
	return res
}

func (s Box) findAll(entityID block.Word) []arm {
	start := sort.Search(len(s.arms), func(i int) bool {
		return s.arms[i].entityID >= entityID
	})
	end := sort.Search(len(s.arms), func(i int) bool {
		return s.arms[i].entityID > entityID
	})
	return s.arms[start:end]
}

func (s Box) runQuery(a arm) Cursor {
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

func (s Box) buildKey(a arm) []block.Word {
	t := s.model.Types[a.entityID]
	var key []block.Word
	for i, c := range t.Rels {
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
	sort.Slice(s.arms, func(i, j int) bool {
		a := s.arms[i]
		b := s.arms[j]

		return a.before(b)
	})

	dupMap := make([]bool, len(s.arms))
	dupsExist := false
	for i, a := range s.arms {
		if dupMap[i] {
			continue
		}
		for j, b := range s.arms[i+1:] {
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
	var arms []arm
	for i, a := range s.arms {
		if !dupMap[i] {
			arms = append(arms, a)
		}
	}
	s.arms = arms
}

func (a arm) constrains(column int) bool {
	return a.mask&(1<<column) != 0
}

func (a arm) supersetOf(b arm) bool {
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

func (a arm) before(b arm) bool {
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

func (a arm) merge(b arm) (arm, bool) {
	if a.entityID != b.entityID {
		return arm{}, false
	}
	res := make([]block.Word, len(a.where))
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
		return arm{}, false
	}
	return arm{a.entityID, a.mask | b.mask, res}, true
}
