package tree

import (
	"errors"
)

// Tree is an implementation of a B-tree over unsigned 64-bit integers. This should form the basis of a disk-based
// database implementation.
//
// Zeros
//
// Due to some details of the representation, it is not currently possible to
// store a value of zero in the tree. Attempts to do so will cause an error to
// be returned from Put.
type Tree struct {
	store BlockStore
	depth int
	start uint64
}

// ErrNotFound signals that the key does not exist in the tree in question.
var ErrNotFound = errors.New("not found")

// ErrStoringZero indicates that an attempt was made to store zero as a value in
// the tree.
var ErrStoringZero = errors.New("cannot store zero value")

// New creates a new tree using the given block store, depth and starting location. The starting location may not be
// zero.
func New(store BlockStore, depth int, start uint64) *Tree {
	return &Tree{store, depth, start}
}

// Get attempts to retrieve the value associated with the provided key.
func (t *Tree) Get(key uint64) (uint64, error) {
	n, err := t.lookupNode(key)
	if err != nil {
		return 0, err
	}
	return n.followLeaf(key)
}

// Put associates the provided key and value.
func (t *Tree) Put(key, value uint64) error {
	if value == 0 {
		return ErrStoringZero
	}
	n, err := t.lookupNode(key)
	if err != nil {
		return err
	}
	return n.put(key, value)
}

func (t *Tree) lookupNode(k uint64) (*node, error) {
	n, err := t.rootNode()
	if err != nil {
		return nil, err
	}
	for i := 0; i < t.depth; i++ {
		n, err = n.followBranch(k)
		if err != nil {
			return nil, err
		}
	}
	return n, nil
}

func (t *Tree) readNode(p uint64) (*node, error) {
	if p == 0 {
		return nil, ErrNotFound
	}
	var n node
	if err := t.store.ReadBlock(p, n.asBytes()); err != nil {
		return nil, err
	}
	return &n, nil
}

func (t *Tree) writeNode(n *node) error {
	return t.store.WriteBlock(n.id, n.asBytes())
}

func (t *Tree) newNode(n *node) (uint64, error) {
	p, err := t.store.NewBlock()
	if err != nil {
		return 0, err
	}
	return p, t.store.WriteBlock(p, n.asBytes())
}
