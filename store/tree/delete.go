package tree

import (
	"github.com/catlev/pkg/store/block"
)

func (t *Tree) Delete(key block.Word) error {
	n, err := t.findNode(key)
	if err != nil {
		return err
	}

	idx := n.probe(key)
	if n.entries[idx].key != key {
		return ErrNotFound
	}

	return t.deleteFromNode(n, idx)
}

func (t *Tree) deleteFromNode(n *node, idx int) error {
	n.remove(idx)

	if n.width <= NodeMinWidth {
		if n.parent != nil {
			return t.balanceTree(n)
		}
		if n.width == 2 {
			t.root = n.entries[(idx+1)%2].key
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

	for _, op := range []struct {
		f func(*node, *node) (bool, error)
		n *node
	}{
		{t.tryBorrowPre, pre},
		{t.tryBorrowSucc, succ},
		{t.tryMergePre, pre},
		{t.tryMergeSucc, succ},
	} {
		ok, err := op.f(n, op.n)
		if err != nil {
			return err
		}
		if ok {
			break
		}
	}

	return nil
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

func (t *Tree) tryBorrowPre(n, pre *node) (bool, error) {
	if pre == nil {
		return false, nil
	}
	if pre.width <= NodeMinWidth {
		return false, nil
	}

	amt := pre.width - NodeMinWidth
	midpoint := pre.entries[NodeMinWidth].key
	copy(n.entries[amt:], n.entries[:])
	copy(n.entries[:amt], pre.entries[NodeMinWidth:])

	for i := NodeMinWidth; i < NodeMaxWidth; i++ {
		pre.entries[i] = nodeEntry{}
	}

	n.parent.entries[n.pos].key = midpoint

	return true, t.writeNodes(n, pre, n.parent)
}

func (t *Tree) tryBorrowSucc(n, succ *node) (bool, error) {
	if succ == nil {
		return false, nil
	}
	if succ.width <= NodeMinWidth {
		return false, nil
	}

	amt := succ.width - NodeMinWidth
	midpoint := succ.entries[amt].key
	copy(n.entries[n.width:], succ.entries[:amt])
	copy(succ.entries[:], succ.entries[amt:])

	for i := NodeMinWidth; i < NodeMaxWidth; i++ {
		succ.entries[i] = nodeEntry{}
	}

	succ.parent.entries[succ.pos].key = midpoint

	return true, t.writeNodes(n, succ, succ.parent)
}

func (t *Tree) tryMergePre(n, pre *node) (bool, error) {
	if pre == nil {
		return false, nil
	}
	if pre.width > NodeMinWidth {
		return false, nil
	}

	copy(pre.entries[pre.width:], n.entries[:])

	err := t.writeNode(pre)
	if err != nil {
		return false, err
	}

	err = t.deleteFromNode(n.parent, n.pos)
	if err != nil {
		return false, err
	}

	return true, t.store.FreeBlock(n.id)
}

func (t *Tree) tryMergeSucc(n, succ *node) (bool, error) {
	if succ == nil {
		return false, nil
	}
	if succ.width > NodeMinWidth {
		return false, nil
	}

	copy(succ.entries[n.width:], succ.entries[:])
	copy(succ.entries[:n.width], n.entries[:])

	err := t.writeNode(succ)
	if err != nil {
		return false, err
	}

	err = t.deleteFromNode(n.parent, n.pos)
	if err != nil {
		return false, err
	}

	return true, t.store.FreeBlock(n.id)
}
