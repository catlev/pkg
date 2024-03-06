package tree

import (
	"sort"

	"github.com/catlev/pkg/domain"
)

// Maps keys to values, based on a block store.
type Tree struct {
	columns int
	key     int
	store   domain.Store
	root    domain.Word
	depth   int
}

type node struct {
	columns int
	key     int
	parent  *node
	pos     int
	id      domain.Word
	width   int
	entries domain.Block
}

func New(columns int, key int, store domain.Store, depth int, root domain.Word) *Tree {
	return &Tree{columns, key, store, root, depth}
}

func (t *Tree) Root() domain.Word {
	return t.root
}

func (t *Tree) Depth() int {
	return t.depth
}

// intuition b - a
// ceteris paribus, a shorter key is less than a longer key
func compareValues(a, b []domain.Word) int {
	for i := 0; i < min(len(a), len(b)); i++ {
		switch {
		case b[i] > a[i]:
			return 1
		case a[i] > b[i]:
			return -1
		}
	}
	switch {
	case len(b) > len(a):
		return 1
	case len(a) > len(b):
		return -1
	default:
		return 0
	}
}

func (t *Tree) findNode(key []domain.Word) (*node, error) {
	if t.depth == 0 {
		return t.readNode(t.columns, t.key, nil, 0, t.root)
	}

	n, err := t.readNode(t.key+1, t.key, nil, 0, t.root)
	if err != nil {
		return nil, err
	}

	for d := t.depth - 1; d > 0; d-- {
		n, err = t.followNode(t.key+1, t.key, n, n.probe(key))
		if err != nil {
			return nil, err
		}
	}

	step := n.probe(key)
	return t.followNode(t.columns, t.key, n, step)
}

func (t *Tree) followNode(columns, key int, n *node, idx int) (*node, error) {
	return t.readNode(columns, key, n, idx, n.getRow(idx)[key])
}

func (t *Tree) readNode(columns, key int, parent *node, pos int, id domain.Word) (*node, error) {
	n := &node{
		columns: columns,
		key:     key,
		parent:  parent,
		pos:     pos,
		id:      id,
	}

	err := t.store.ReadBlock(id, &n.entries)
	if err != nil {
		return nil, err
	}

	n.width = sort.Search(n.maxWidth(), func(i int) bool {
		if i == 0 {
			return false
		}
		candidate := n.getKey(i)
		for _, k := range candidate {
			if k != 0 {
				return false
			}
		}
		return true
	})

	return n, nil
}

func (t *Tree) writeNode(n *node) error {
	id, err := t.store.WriteBlock(n.id, n.entriesAsBlock())
	if err != nil {
		return err
	}
	if id == n.id {
		// the id of the node hasn't changed, so we don't need to update the parent
		return nil
	}
	if n.parent == nil {
		// when the node's parent is nil, it's the root node
		t.root = id
		return nil
	}

	n.parent.getRow(n.pos)[t.key] = id
	return t.writeNode(n.parent)
}

func (t *Tree) writeNodes(ns ...*node) error {
	for _, n := range ns {
		err := t.writeNode(n)
		if err != nil {
			return err
		}
	}
	return nil
}

func (n *node) probe(key []domain.Word) int {
	return sort.Search(n.width-1, func(i int) bool {
		candidate := n.getKey(i + 1)
		return compareValues(candidate, key) < 0
	})
}

func (n *node) insert(idx int, entries ...[]domain.Word) {
	copy(n.entries[(idx+len(entries))*n.columns:], n.entries[idx*n.columns:])
	for i, r := range entries {
		copy(n.entries[(idx+i)*n.columns:], r)
	}
	n.width += len(entries)
}

func (n *node) remove(idx, count int) {
	n.width -= count
	copy(n.entries[idx*n.columns:], n.entries[(idx+count)*n.columns:])
	n.clearRows(n.width, -1)
}

func (n *node) entriesAsBlock() *domain.Block {
	return &n.entries
}

func (n *node) getKey(idx int) []domain.Word {
	if idx == 0 {
		if n.parent == nil {
			return make([]domain.Word, n.key)
		}
		return n.parent.getKey(n.pos)
	}
	return n.getRow(idx)[:n.key]
}

func (n *node) setkey(idx int, from []domain.Word) {
	row := n.getRow(idx)
	copy(row, from)
}

func (n *node) getRow(idx int) []domain.Word {
	return n.entries[idx*n.columns : (idx+1)*n.columns]
}

func (n *node) getRows(from, to int) [][]domain.Word {
	rows := make([][]domain.Word, to-from)
	for i := range rows {
		rows[i] = n.getRow(i + from)
	}
	return rows
}

func (n *node) clearRows(from, to int) {
	stop := to * n.columns
	if to == -1 {
		stop = domain.WordSize
	}
	for i := from * n.columns; i < stop; i++ {
		n.entries[i] = 0
	}
}

func (n *node) minWidth() int {
	return n.maxWidth() / 2
}

func (n *node) maxWidth() int {
	return domain.WordSize / n.columns
}

func (n *node) compareKeyAt(idx int, key []domain.Word) int {
	if n == nil {
		return 1
	}
	return compareValues(n.getKey(idx), key)
}
