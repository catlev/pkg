package tree

import (
	"errors"

	"github.com/catlev/pkg/store/block"
)

var ErrBadRow = errors.New("bad row")

// Put establishes an association between key and value. Errors may be relayed from the block store.
func (t *Tree) Put(key, value block.Word) error {

	n, err := t.findNode(key)
	if err != nil {
		return err
	}

	idx := n.probe(key)

	row := []block.Word{key, value}

	if rowsEqual(n.getRow(idx), row) {
		// no update required
		return nil
	}

	if n.keyFor(idx) == key {
		// updating existing entry, no need to make room
		copy(n.entries[idx*n.columns:], row)
		return t.writeNode(n)
	}

	_, err = t.addNodeEntry(n, key, row)
	return err
}

func (t *Tree) addNodeEntry(n *node, key block.Word, row []block.Word) (*node, error) {
	var err error

	if n == nil {
		rootNode := &node{
			columns: branchColumns,
			key:     keyField,
			entries: block.Block{0, t.root, key, row[valueField]},
		}
		rootNode.id, err = t.store.AddBlock(rootNode.entriesAsBlock())
		if err != nil {
			return nil, err
		}

		t.root = rootNode.id
		t.depth++
		return rootNode, nil
	}

	if n.width == n.maxWidth() {
		// out of room in this node, so split (and update ancestors)
		n, err = t.splitNode(n, key)
		if err != nil {
			return nil, err
		}
	}

	n.insert(n.probe(key)+1, row)

	err = t.writeNode(n)
	if err != nil {
		return nil, err
	}
	return n, nil
}

func (t *Tree) splitNode(n *node, key block.Word) (*node, error) {
	midpoint := n.keyFor(n.minWidth())

	newNode := &node{
		columns: n.columns,
		key:     n.key,
	}
	newNode.insert(0, n.getRows(n.minWidth(), n.maxWidth())...)

	n.clearRows(n.minWidth(), -1)

	id, err := t.store.AddBlock(newNode.entriesAsBlock())
	if err != nil {
		return nil, err
	}

	err = t.writeNode(n)
	if err != nil {
		return nil, err
	}

	n.parent, err = t.addNodeEntry(n.parent, midpoint, []block.Word{midpoint, id})
	if err != nil {
		return nil, err
	}

	if key > midpoint {
		newNode.parent = n.parent
		newNode.pos = 1
		newNode.id = id
		return newNode, nil
	}

	return &node{
		columns: n.columns,
		key:     n.key,
		parent:  n.parent,
		id:      n.id,
		pos:     0,
		entries: n.entries,
		width:   n.minWidth(),
	}, nil
}
