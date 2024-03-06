package tree

import "github.com/catlev/pkg/domain"

// Delete removes the association between the given key and its value. If no association exists,
// ErrNotFound is returned. Errors may also originate from the block store.
func (t *Tree) Delete(key []domain.Word) (err error) {
	defer wrapErr(&err, "Delete", key)

	if len(key) != t.key {
		return ErrKeyWidth
	}

	n, err := t.findNode(key)
	if err != nil {
		return err
	}

	idx := n.probe(key)
	if compareValues(n.getKey(idx), key) != 0 {
		return err
	}

	return t.deleteFromNode(n, idx)
}

func (t *Tree) deleteFromNode(n *node, idx int) error {
	n.remove(idx, 1)

	if n.width <= n.minWidth() {
		if n.parent != nil {
			return t.balanceTree(n)
		}
		if n.width == 2 {
			// superfluous root node
			t.root = n.getRow((idx + 1) % 2)[t.key]
			t.depth--
			return t.store.FreeBlock(n.id)
		}
	}

	return t.writeNode(n)
}

func (t *Tree) balanceTree(n *node) error {
	pre, err := t.getPre(n)
	if err != nil {
		return err
	}
	succ, err := t.getSucc(n)
	if err != nil {
		return err
	}

	if pre != nil && pre.width > pre.minWidth() {
		return t.borrowPre(n, pre)
	}
	if succ != nil && succ.width > succ.minWidth() {
		return t.borrowSucc(n, succ)
	}
	if pre != nil && pre.width <= pre.minWidth() {
		return t.mergePre(n, pre)
	}
	if succ != nil && succ.width <= succ.minWidth() {
		return t.mergeSucc(n, succ)
	}

	panic("unreachable")
}

func (t *Tree) getPre(n *node) (*node, error) {
	if n.parent == nil {
		return nil, nil
	}
	if n.pos == 0 {
		pp, err := t.getPre(n.parent)
		if err != nil {
			return nil, err
		}
		if pp == nil {
			return nil, nil
		}
		return t.followNode(n.columns, n.key, pp, pp.width-1)
	}
	return t.followNode(n.columns, n.key, n.parent, n.pos-1)
}

func (t *Tree) getSucc(n *node) (*node, error) {
	if n.parent == nil {
		return nil, nil
	}
	if n.pos == n.parent.width-1 {
		pp, err := t.getSucc(n.parent)
		if err != nil {
			return nil, err
		}
		if pp == nil {
			return nil, nil
		}
		return t.followNode(n.columns, n.key, pp, 0)
	}
	return t.followNode(n.columns, n.key, n.parent, n.pos+1)
}

func (t *Tree) borrowPre(n, pre *node) error {
	midpoint := make([]domain.Word, t.key)
	copy(midpoint, pre.getKey(n.minWidth()))

	n.setkey(0, n.getKey(0))
	n.insert(0, pre.getRows(n.minWidth(), pre.width)...)
	n.setkey(0, make([]domain.Word, t.key))

	pre.clearRows(n.minWidth(), -1)

	n.parent.setkey(n.pos, midpoint)

	return t.writeNodes(n, pre, n.parent)
}

func (t *Tree) borrowSucc(n, succ *node) error {
	amt := succ.width - succ.minWidth()
	midpoint := make([]domain.Word, t.key)
	copy(midpoint, succ.getKey(amt))

	succ.setkey(0, succ.getKey(0))
	n.insert(n.width, succ.getRows(0, amt)...)
	succ.remove(0, amt)
	succ.setkey(0, make([]domain.Word, t.key))

	succ.parent.setkey(succ.pos, midpoint)

	return t.writeNodes(n, succ, succ.parent)
}

func (t *Tree) mergePre(n, pre *node) error {
	n.setkey(0, n.getKey(0))
	copy(pre.entries[pre.width*pre.columns:], n.entries[:])

	err := t.writeNode(pre)
	if err != nil {
		return err
	}

	err = t.deleteFromNode(n.parent, n.pos)
	if err != nil {
		return err
	}

	return t.store.FreeBlock(n.id)
}

func (t *Tree) mergeSucc(n, succ *node) error {
	n.setkey(0, n.getKey(0))
	succ.setkey(0, succ.getKey(0))
	copy(succ.entries[n.width*n.columns:], succ.entries[:])
	copy(succ.entries[:n.width*n.columns], n.entries[:])
	n.setkey(0, make([]domain.Word, t.key))

	err := t.writeNode(succ)
	if err != nil {
		return err
	}

	err = t.deleteFromNode(n.parent, n.pos)
	if err != nil {
		return err
	}

	return t.store.FreeBlock(n.id)
}
