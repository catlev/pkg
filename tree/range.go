package tree

// Range is an iterator over a range of keys within a particular tree. Because it has knowledge of the underlying data
// structure, this iterator is more efficient than querying a range of keys yourself.
type Range struct {
	to     uint64
	node   *node
	bucket int
	err    error
}

// GetRange returns an iterator-style object that yields all values that have been associated with keys in the half-open
// range [from, to).
func (t *Tree) GetRange(from, to uint64) *Range {
	n, err := t.lookupNode(from)
	if err != nil {
		return &Range{err: err}
	}
	b := n.bucket(from)
	if b == 1 {
		b = 2
	}
	return &Range{to: to, node: n, bucket: b - 2}
}

// Next will attempt to move to the next entry, returning whether this was possible. Once a call to Next has returned
// false, it will continue to do so in perpetuity.
func (r *Range) Next() bool {
	if r.node == nil  {
		return false
	}
	r.bucket++
	if r.bucket == r.node.len() {
		// the current leaf node has been exhausted
		r.node, r.err = r.nextNode(r.node)
		if r.node == nil {
			return false
		}
		r.bucket = 0
	}
	if r.bucket != 0 && r.node.entries[r.bucket].from >= r.to {
		r.node = nil
		return false
	}
	return true
}

// This returns the current value that the Range points to. If Next has not been called, or the last call to Next
// returned false, the behaviour of calling This is undefined.
func (r *Range) This() uint64 {
	return r.node.entries[r.bucket].value
}

func (r *Range) Err() error {
	return r.err
}

// nextNode traverses the tree finding the next leaf node.
func (r *Range) nextNode(n *node) (*node, error) {
	depth := 0
	// go up to an ancestor that has more nodes
	for n.max == r.node.max {
		n = n.parent
		depth++
		if n == nil {
			return nil, nil
		}
	}
	// go back down to the new leaf node
	for i := 0; i < depth; i++ {
		next, err := n.followBranch(r.node.max + 1)
		if err != nil {
			return nil, err
		}
		n = next
	}
	return n, nil
}
