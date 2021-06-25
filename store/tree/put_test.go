package tree

import (
	"math"
	"testing"

	"github.com/catlev/pkg/store/block"
	"github.com/catlev/pkg/store/block/mem"
)

func TestPutUpdate(t *testing.T) {
	store := mem.New()
	d1, _ := store.AddBlock(&block.Block{0, 1, 10, 2, 20, 3})
	d2, _ := store.AddBlock(&block.Block{0, 4, 40, 5, 50, 6})
	start, _ := store.AddBlock(&block.Block{0, d1, 30, d2})
	tree := New(store, 1, start)

	tree.Put(20, 7)
	id, _ := tree.Get(20)

	if id != 7 {
		t.Fail()
	}
}

func TestPutAddWithRoom(t *testing.T) {
	store := mem.New()
	d1, _ := store.AddBlock(&block.Block{0, 1, 10, 2, 20, 3})
	d2, _ := store.AddBlock(&block.Block{0, 4, 40, 5, 50, 6})
	start, _ := store.AddBlock(&block.Block{0, d1, 30, d2})
	tree := New(store, 1, start)

	tree.Put(25, 7)
	id, _ := tree.Get(25)

	if id != 7 {
		t.Fail()
	}
}

func TestPutAddRoot(t *testing.T) {
	store := mem.New()
	var b block.Block
	for i := range b {
		b[i] = block.Word(i)
	}
	start, _ := store.AddBlock(&b)
	tree := New(store, 0, start)

	tree.Put(25, 100)
	id, _ := tree.Get(25)

	if id != 100 {
		t.Fail()
	}
}

func TestPutAddRootLarge(t *testing.T) {
	store := mem.New()
	var b block.Block
	for i := range b {
		b[i] = block.Word(i)
	}
	start, _ := store.AddBlock(&b)
	tree := New(store, 0, start)

	tree.Put(45, 100)
	id, _ := tree.Get(45)

	if id != 100 {
		t.Fail()
	}
}

func TestPutDeep(t *testing.T) {
	store := mem.New()
	start, _ := store.AddBlock(&block.Block{})
	tree := New(store, 0, start)

	for i := 0; i < 100; i++ {
		tree.Put(block.Word(i), block.Word(i*2))
	}

	id, _ := tree.Get(50)

	if id != 100 {
		t.Log(tree.readNode(nil, 0, 2560, 0, math.MaxUint64))
		t.Log(id)
		t.Fail()
	}
}
