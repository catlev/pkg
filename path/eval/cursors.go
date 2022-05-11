package eval

import "github.com/catlev/pkg/store/block"

type Cursor interface {
	Next() bool
	This() Object
	Err() error
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
	arm  Query
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
	arm Query
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

func (c *boxCursor) Err() error {
	return c.cur.Err()
}

func (c *boxCursor) Next() bool {
	for !c.cur.Next() {
		c.arm++
		if c.arm >= len(c.box.contents) {
			return false
		}
		c.cur = c.box.runQuery(c.box.contents[c.arm])
	}
	return true
}

func (c *boxCursor) This() Object {
	return c.cur.This()
}
