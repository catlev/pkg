package tree

import (
	"fmt"

	"github.com/catlev/pkg/store/block"
)

// Put establishes an association between key and value. Errors may be relayed from the block store.
func (t *Tree) Put(key, id block.Word) error {
	n, err := t.findNode(key)
	if err != nil {
		return fmt.Errorf("put %d: %w", id, err)
	}

	idx := n.probe(key)

	if n.entries[idx].value == id {
		// no update required
		return nil
	}

	if n.keyFor(idx) == key {
		// updating existing entry, no need to make room
		n.entries[idx].value = id
		err = t.writeNode(n)
		if err != nil {
			return fmt.Errorf("put %d: %w", id, err)
		}
	}

	_, err = t.addNodeEntry(n, key, id)
	if err != nil {
		return fmt.Errorf("put %d: %w", id, err)
	}
	return nil
}

func (t *Tree) addNodeEntry(n *node, key, id block.Word) (*node, error) {
	var err error

	if n == nil {
		rootNode := &node{
			entries: [32]nodeEntry{{0, t.root}, {key, id}},
		}
		rootNode.id, err = t.store.AddBlock(rootNode.entriesAsBlock())
		if err != nil {
			return nil, err
		}

		t.root = rootNode.id
		t.depth++
		return rootNode, nil
	}

	if n.width == NodeMaxWidth {
		// out of room in this node, so split (and update ancestors)
		n, err = t.splitNode(n, key)
		if err != nil {
			return nil, err
		}
	}

	n.insert(n.probe(key)+1, nodeEntry{key, id})

	err = t.writeNode(n)
	if err != nil {
		return nil, err
	}
	return n, nil
}

func (t *Tree) splitNode(n *node, key block.Word) (*node, error) {
	midpoint := n.keyFor(NodeMinWidth)

	newNode := new(node)
	newNode.insert(0, n.entries[NodeMinWidth:]...)

	for i := NodeMinWidth; i < NodeMaxWidth; i++ {
		n.entries[i] = nodeEntry{}
	}

	id, err := t.store.AddBlock(newNode.entriesAsBlock())
	if err != nil {
		return nil, err
	}

	err = t.writeNode(n)
	if err != nil {
		return nil, err
	}

	n.parent, err = t.addNodeEntry(n.parent, midpoint, id)
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
		parent:  n.parent,
		id:      n.id,
		pos:     0,
		entries: n.entries,
		width:   NodeMinWidth,
	}, nil
}
