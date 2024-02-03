package tree

import (
	"github.com/catlev/pkg/store/block"
)

// Put establishes an association between key and value. Errors may be relayed from the block store.
func (t *Tree) Put(row []block.Word) (err error) {
	if len(row) != t.columns {
		return &TreeError{
			Err: ErrBadRow,
			Op:  "Put",
		}
	}

	key := make([]block.Word, len(t.key))
	extractKey(row, t.key, key)

	defer wrapErr(&err, "Put", key)

	n, err := t.findNode(key)
	if err != nil {
		return err
	}

	idx := n.probe(key)

	if compareValues(n.getRow(idx), row) == 0 {
		// no update required
		return nil
	}

	if n.compareKeyAt(idx, key) == 0 {
		// updating existing entry, no need to make room
		copy(n.entries[idx*n.columns:], row)
		return t.writeNode(n)
	}

	_, err = t.addNodeEntry(n, key, row)
	return err
}

func (t *Tree) addNodeEntry(n *node, key []block.Word, r []block.Word) (*node, error) {
	var err error

	if n == nil {
		rootNode := &node{
			columns: len(t.key) + 1,
			key:     t.ixKey,
		}
		rootNode.entries[len(t.key)] = t.root
		copy(rootNode.entries[len(t.key)+1:], r)

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

	n.insert(n.probe(key)+1, r)

	err = t.writeNode(n)
	if err != nil {
		return nil, err
	}
	return n, nil
}

func (t *Tree) splitNode(n *node, key []block.Word) (*node, error) {
	midpoint := n.getKey(n.minWidth())

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

	row := make([]block.Word, len(t.key)+1)
	copy(row, midpoint)
	row[len(t.key)] = id

	n.parent, err = t.addNodeEntry(n.parent, midpoint, row)
	if err != nil {
		return nil, err
	}

	if compareValues(midpoint, key) > 0 {
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
