package tree

import (
	"errors"
	"sort"
	"unsafe"

	"github.com/catlev/pkg/store/block"
)

var ErrNotFound = errors.New("not found")

// Maps keys to values, based on a block store.
type Tree struct {
	columns, key int
	store        block.Store
	root         block.Word
	depth        int
}

type node struct {
	columns, key int
	parent       *node
	pos          int
	id           block.Word
	width        int
	entries      block.Block
}

const (
	keyField = iota
	valueField
	branchColumns
)

func New(columns, key int, s block.Store, dep int, root block.Word) *Tree {
	return &Tree{columns, key, s, root, dep}
}

func (t *Tree) Root() block.Word {
	return t.root
}

func (t *Tree) findNode(key block.Word) (*node, error) {
	if t.depth == 0 {
		return t.readNode(t.columns, t.key, nil, 0, t.root)
	}

	n, err := t.readNode(branchColumns, keyField, nil, 0, t.root)
	if err != nil {
		return nil, err
	}

	for d := t.depth - 1; d > 0; d-- {
		n, err = t.followNode(branchColumns, keyField, n, n.probe(key))
		if err != nil {
			return nil, err
		}
	}

	return t.followNode(t.columns, t.key, n, n.probe(key))
}

func (t *Tree) followNode(columns, key int, n *node, idx int) (*node, error) {
	return t.readNode(columns, key, n, idx, n.getRow(idx)[valueField])
}

func (t *Tree) readNode(columns, key int, parent *node, pos int, id block.Word) (*node, error) {
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
		return i != 0 && n.keyFor(i) == 0
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

	n.parent.getRow(n.pos)[valueField] = id
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

func (n *node) probe(key block.Word) int {
	return sort.Search(n.width-1, func(i int) bool {
		return n.keyFor(i+1) > key
	})
}

func (n *node) insert(idx int, entries ...[]block.Word) {
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

func (n *node) entriesAsBlock() *block.Block {
	return (*block.Block)(unsafe.Pointer(&n.entries))
}

func (n *node) keyFor(idx int) block.Word {
	if n == nil {
		return 0
	}
	if idx == 0 {
		return n.parent.keyFor(n.pos)
	}
	return n.getRow(idx)[n.key]
}

func (n *node) getRow(idx int) []block.Word {
	return n.entries[idx*n.columns : (idx+1)*n.columns]
}

func (n *node) getRows(from, to int) [][]block.Word {
	rows := make([][]block.Word, to-from)
	for i := range rows {
		rows[i] = n.entries[(i+from)*n.columns : (i+from+1)*n.columns]
	}
	return rows
}

func (n *node) clearRows(from, to int) {
	stop := to * n.columns
	if to == -1 {
		stop = block.WordSize
	}
	for i := from * n.columns; i < stop; i++ {
		n.entries[i] = 0
	}
}

func (n *node) minWidth() int {
	return n.maxWidth() / 2
}

func (n *node) maxWidth() int {
	return block.WordSize / n.columns
}

func rowsEqual(r, s []block.Word) bool {
	for i := range r {
		if r[i] != s[i] {
			return false
		}
	}
	return true
}
