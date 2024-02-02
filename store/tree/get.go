package tree

import (
	"github.com/catlev/pkg/store/block"
)

type Range struct {
	tree *Tree
	node *node
	key  []block.Word
	pos  int
	err  error
}

// Get queries the tree using the given key, yielding the associated value. If no value has been
// associated with the given key, then ErrNotFound is returned as an error. Errors may also
// originate from the block store.
func (t *Tree) Get(key []block.Word) (v []block.Word, err error) {
	wrapErr(&err, "Get", key)

	if len(key) != len(t.key) {
		return nil, ErrKeyWidth
	}

	n, err := t.findNode(key)
	if err != nil {
		return nil, err
	}

	idx := n.probe(key)
	candidate := make([]block.Word, len(t.key))
	n.keyFor(idx, candidate)

	if compareValues(candidate, key) != 0 {
		return nil, ErrNotFound
	}

	return n.getRow(idx), nil
}

// GetRange queries the tree using the given key and returns an iterator over entries of the tree,
// starting with the largest key that is less than or equal to the given key.
func (t *Tree) GetRange(key []block.Word) *Range {
	n, err := t.findNode(key)
	pos := 0
	if n != nil {
		pos = n.probe(key)
	}

	return &Range{
		tree: t,
		node: n,
		key:  key,
		pos:  pos - 1,
		err:  err,
	}
}

func (r *Range) Next() bool {
	if r.err != nil {
		return false
	}

	r.pos++

	if r.pos < r.node.width {
		return true
	}

	n, ok := r.loadNextNode(r.node)
	r.node = n
	r.pos = 0

	return ok
}

func (r *Range) This() []block.Word {
	return r.node.getRow(r.pos)
}

func (r *Range) Err() error {
	if r.err == nil {
		return nil
	}
	return &TreeError{
		Op:  "GetRange",
		Key: r.key,
		Err: r.err,
	}
}

func (r *Range) loadNextNode(n *node) (*node, bool) {
	if n.parent == nil {
		return nil, false
	}

	next := n.pos + 1

	if next >= n.parent.width {
		p, ok := r.loadNextNode(n.parent)
		if !ok {
			return nil, false
		}
		n.parent = p
	}

	p, err := r.tree.followNode(n.columns, n.key, n.parent, next)
	if err != nil {
		r.err = err
		return nil, false
	}

	return p, true
}
