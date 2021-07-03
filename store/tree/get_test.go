package tree

import (
	"testing"

	"github.com/catlev/pkg/store/block"
	"github.com/catlev/pkg/store/block/mem"
)

func assertTreeProperty(t *testing.T, i int, k block.Word) {
	t.Helper()
	if int(k) != (i/10)+1 {
		t.Errorf("failed on %d: %d != %d", i, k, (i/10)+1)
	}
}

func buildBlock(off int) *block.Block {
	var b block.Block
	for i := 0; i < NodeMaxWidth; i++ {
		n := i + off
		b[i*2] = block.Word(n)
		b[i*2+1] = block.Word((n / 10) + 1)
	}
	b[0] = 0
	return &b
}

func TestProbe(t *testing.T) {
	store := mem.New()
	start, _ := store.AddBlock(buildBlock(0))
	tree := New(store, 0, start)
	node, _ := tree.readNode(nil, 0, start)

	for i := 0; i < 25; i++ {
		k := node.probe(block.Word(i))
		if i != k {
			t.Fail()
		}
	}
}

func TestGetShallow(t *testing.T) {
	store := mem.New()
	start, _ := store.AddBlock(buildBlock(0))
	tree := New(store, 0, start)

	for i := 0; i < 25; i++ {
		k, _ := tree.Get(block.Word(i))
		assertTreeProperty(t, i, k)
	}
}

func TestGetDeep(t *testing.T) {
	store := mem.New()
	d1, _ := store.AddBlock(buildBlock(0))
	d2, _ := store.AddBlock(buildBlock(32))
	start, _ := store.AddBlock(&block.Block{0, d1, 32, d2})
	tree := New(store, 1, start)

	for i := 0; i < 60; i++ {
		k, _ := tree.Get(block.Word(i))
		assertTreeProperty(t, i, k)
	}
}
