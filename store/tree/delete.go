package tree

import (
	"fmt"

	"github.com/catlev/pkg/store/block"
)

// Delete removes the association between the given key and its value. If no association exists,
// ErrNotFound is returned. Errors may also originate from the block store.
func (t *Tree) Delete(key block.Word) error {
	n, err := t.findNode(key)
	if err != nil {
		return fmt.Errorf("delete %d: %w", key, err)
	}

	idx := n.probe(key)
	if n.keyFor(idx) != key {
		return fmt.Errorf("delete %d: %w", key, ErrNotFound)
	}

	return t.deleteFromNode(n, idx)
}

func (t *Tree) deleteFromNode(n *node, idx int) error {
	n.remove(idx, 1)

	if n.width <= NodeMinWidth {
		if n.parent != nil {
			return t.balanceTree(n)
		}
		if n.width == 2 {
			// superfluous root node
			t.root = n.entries[(idx+1)%2].value
			t.depth--
			return t.store.FreeBlock(n.id)
		}
		return t.writeNode(n)
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

	if pre != nil && pre.width > NodeMinWidth {
		return t.borrowPre(n, pre)
	}
	if succ != nil && succ.width > NodeMinWidth {
		return t.borrowSucc(n, succ)
	}
	if pre != nil && pre.width <= NodeMinWidth {
		return t.mergePre(n, pre)
	}
	if succ != nil && succ.width <= NodeMinWidth {
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
		return t.followNode(pp, pp.width-1)
	}
	return t.followNode(n.parent, n.pos-1)
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
		return t.followNode(pp, 0)
	}
	return t.followNode(n.parent, n.pos+1)
}

func (t *Tree) borrowPre(n, pre *node) error {
	midpoint := pre.keyFor(NodeMinWidth)

	n.entries[0].key = n.keyFor(0)
	n.insert(0, pre.entries[NodeMinWidth:pre.width]...)
	n.entries[0].key = 0

	for i := NodeMinWidth; i < NodeMaxWidth; i++ {
		pre.entries[i] = nodeEntry{}
	}

	n.parent.entries[n.pos].key = midpoint

	return t.writeNodes(n, pre, n.parent)
}

func (t *Tree) borrowSucc(n, succ *node) error {
	amt := succ.width - NodeMinWidth
	midpoint := succ.keyFor(amt)

	succ.entries[0].key = succ.keyFor(0)
	n.insert(n.width, succ.entries[:amt]...)
	succ.remove(0, amt)
	succ.entries[0].key = 0

	succ.parent.entries[succ.pos].key = midpoint

	return t.writeNodes(n, succ, succ.parent)
}

func (t *Tree) mergePre(n, pre *node) error {
	n.entries[0].key = n.keyFor(0)
	copy(pre.entries[pre.width:], n.entries[:])

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
	n.entries[0].key = n.keyFor(0)
	succ.entries[0].key = succ.keyFor(0)
	copy(succ.entries[n.width:], succ.entries[:])
	copy(succ.entries[:n.width], n.entries[:])
	n.entries[0].key = 0

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
