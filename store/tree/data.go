package tree

import (
	"errors"
	"sort"
	"unsafe"

	"github.com/catlev/pkg/store/block"
)

var ErrNotFound = errors.New("not found")

const NodeMaxWidth = 32
const NodeMinWidth = NodeMaxWidth / 2

// Stores a tree of block references. The referenced blocks contain data that this tree is acting as
// an index for.
type Tree struct {
	store block.Store
	root  block.Word
	depth int
}

type node struct {
	parent  *node
	pos     int
	id      block.Word
	width   int
	entries [32]nodeEntry
}

type nodeEntry struct {
	key, value block.Word
}

func New(s block.Store, dep int, pos block.Word) *Tree {
	return &Tree{s, pos, dep}
}

func (t *Tree) findNode(key block.Word) (*node, error) {
	n, err := t.readNode(nil, 0, t.root)
	if err != nil {
		return nil, err
	}

	for d := t.depth - 1; d >= 0; d-- {
		n, err = t.followNode(n, n.probe(key))
		if err != nil {
			return nil, err
		}
	}

	return n, nil
}

func (t *Tree) followNode(n *node, idx int) (*node, error) {
	return t.readNode(n, idx, n.entries[idx].value)
}

func (t *Tree) readNode(parent *node, pos int, id block.Word) (*node, error) {
	n := &node{
		parent: parent,
		pos:    pos,
		id:     id,
	}

	err := t.store.ReadBlock(id, n.entriesAsBlock())
	if err != nil {
		return nil, err
	}

	n.width = sort.Search(NodeMaxWidth, func(i int) bool { return i != 0 && n.entries[i].key == 0 })

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

	n.parent.entries[n.pos].value = id
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
	return sort.Search(n.width, func(i int) bool { return n.entries[i].key > key }) - 1
}

func (n *node) insert(idx int, key, id block.Word) {
	idx++
	copy(n.entries[idx+1:], n.entries[idx:])
	n.entries[idx] = nodeEntry{key, id}
}

func (n *node) remove(idx int) {
	copy(n.entries[idx:], n.entries[idx+1:])
}

func (n *node) entriesAsBlock() *block.Block {
	return (*block.Block)(unsafe.Pointer(&n.entries))
}
