package tree

import (
	"fmt"

	"github.com/catlev/pkg/store/block"
)

// Retrieve the block ID that may contain the specified key. If no such ID can be found, the
// returned word is zero and the returned error is ErrNotFound. Errors can also be returned from the
// storage layer.
func (t *Tree) Get(key block.Word) (block.Word, error) {
	n, err := t.findNode(key)
	if err != nil {
		return 0, fmt.Errorf("get %d: %w", key, err)
	}

	idx := n.probe(key)
	if n.entries[idx].key != key {
		return 0, ErrNotFound
	}
	v := n.entries[n.probe(key)].value

	if v == 0 {
		return 0, ErrNotFound
	}
	return v, nil
}
