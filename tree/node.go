package tree

import (
	"math"
	"reflect"
	"sort"
	"unsafe"
)

const nodeSize = 32

type node struct {
	// metadata inferred from navigating the tree
	tree     *Tree
	parent   *node
	id       uint64
	min, max uint64

	// actual stored data
	entries [nodeSize]entry
}

type entry struct {
	from, value uint64
}

type nodePut struct {
	node   *node
	bucket int
	key    uint64
	err    error
}

func (t *Tree) rootNode() (*node, error) {
	n, err := t.readNode(t.start)
	if err != nil {
		return nil, err
	}

	n.tree = t
	n.id = t.start
	n.min = 0
	n.max = math.MaxUint64

	return n, nil
}

func (n *node) asBytes() []byte {
	sh := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(&n.entries)),
		Len:  512,
		Cap:  512,
	}
	return *(*[]byte)(unsafe.Pointer(&sh))
}

func (n *node) bucket(k uint64) int {
	return sort.Search(n.len(), func(idx int) bool {
		return idx != 0 && n.entries[idx].from > k
	}) - 1
}

func (n *node) len() int {
	return sort.Search(nodeSize, func(idx int) bool {
		return idx != 0 && n.entries[idx].from == 0
	})
}

func (n *node) entryRange(bucket int) (min, max uint64) {
	if bucket == 0 {
		min = n.min
	} else {
		min = n.entries[bucket].from
	}
	if bucket >= n.len()-1 {
		max = n.max
	} else {
		max = n.entries[bucket+1].from
	}
	return min, max
}

func (n *node) followBranch(k uint64) (*node, error) {
	b := n.bucket(k)
	entry := n.entries[b]

	next, err := n.tree.readNode(entry.value)
	if err != nil {
		return nil, err
	}

	next.tree = n.tree
	next.id = entry.value
	next.parent = n
	next.min, next.max = n.entryRange(b)

	return next, nil
}

func (n *node) followLeaf(k uint64) (uint64, error) {
	b := n.bucket(k)
	e := n.entries[b]

	min, _ := n.entryRange(b)
	if k != min {
		return 0, ErrNotFound
	}
	if e.value == 0 {
		return 0, ErrNotFound
	}
	return e.value, nil
}

func (n *node) full() bool {
	return n.entries[nodeSize-1].from != 0
}

func (n *node) put(k, v uint64) error {
	p := nodePut{node: n, key: k}

	p.probeNode()
	p.insert()
	p.saveValue(v)

	return p.err
}

func (p *nodePut) split() {
	p.node, p.err = p.node.split(p.key)
	if p.err == nil {
		p.bucket = p.node.bucket(p.key)
	}
}

func (p *nodePut) probeNode() {
	if p.err != nil {
		return
	}

	p.bucket = p.node.bucket(p.key)
	if p.bucket == nodeSize {
		p.split()
	}
}

func (p *nodePut) insert() {
	if p.err != nil {
		return
	}

	min, _ := p.node.entryRange(p.bucket)
	if p.key == min {
		return
	}

	if p.node.full() {
		p.split()
		if p.err != nil {
			return
		}
	}

	p.bucket++
	prev := p.node.insert(p.bucket)
	prev.from = p.key
}

func (p *nodePut) saveValue(value uint64) {
	if p.err != nil {
		return
	}
	p.node.entries[p.bucket].value = value
	p.err = p.node.tree.writeNode(p.node)
}

func (n *node) insert(at int) *entry {
	copy(n.entries[at+1:], n.entries[at:])
	return &n.entries[at]
}

func (n *node) split(k uint64) (*node, error) {
	left := n
	right := &node{
		tree:   n.tree,
		parent: n.parent,
		min:    n.entries[nodeSize/2].from,
		max:    n.max,
	}
	left.max = right.min

	copy(right.entries[:], left.entries[nodeSize/2:])
	right.entries[0].from = 0

	for i := nodeSize / 2; i < nodeSize; i++ {
		left.entries[i] = entry{}
	}

	err := n.tree.writeNode(left)
	if err != nil {
		return nil, err
	}

	right.id, err = n.tree.newNode(right)
	if err != nil {
		return nil, err
	}

	if n.parent == nil {
		parent := &node{
			tree:    n.tree,
			min:     0,
			max:     math.MaxUint64,
			entries: [nodeSize]entry{{value: left.id}},
		}
		parent.id, err = n.tree.newNode(parent)
		if err != nil {
			return nil, err
		}
		left.parent = parent
		right.parent = parent
		n.tree.start = parent.id
		n.tree.depth++
	}

	err = n.parent.put(right.min, right.id)
	if err != nil {
		return nil, err
	}

	if left.max > k {
		return left, nil
	}
	return right, nil
}
