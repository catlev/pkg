package tree

import (
	"fmt"

	"github.com/catlev/pkg/store/block"
)

// Get queries the tree using the given key, yielding the associated value. If no value has been
// associated with the given key, then ErrNotFound is returned as an error. Errors may also
// originate from the block store.
func (t *Tree) Get(key block.Word) (block.Word, error) {
	n, err := t.findNode(key)
	if err != nil {
		return 0, fmt.Errorf("get %d: %w", key, err)
	}

	idx := n.probe(key)
	if n.keyFor(idx) != key {
		return 0, fmt.Errorf("get %d: %w", key, ErrNotFound)
	}
	return n.entries[idx].value, nil
}
