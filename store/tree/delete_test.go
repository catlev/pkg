package tree

import (
	"errors"
	"testing"

	"github.com/catlev/pkg/store/block"
	"github.com/catlev/pkg/store/block/mem"
)

func TestDeleteLeafSimple(t *testing.T) {
	store := mem.New()
	start, _ := store.AddBlock(buildBlock(0))
	tree := New(store, 0, start)

	tree.Delete(10)

	if _, err := tree.Get(10); !errors.Is(err, ErrNotFound) {
		t.Fail()
	}
}

func TestDeleteBorrowPre(t *testing.T) {
	store := mem.New()
	d1, _ := store.AddBlock(buildBlock(0))

	b := buildBlock(32)
	for i := 32; i < 64; i++ {
		(*b)[i] = 0
	}

	d2, _ := store.AddBlock(b)
	start, _ := store.AddBlock(&block.Block{0, d1, 32, d2})
	tree := New(store, 1, start)

	tree.Delete(35)

	if _, err := tree.Get(35); !errors.Is(err, ErrNotFound) {
		t.Fail()
	}
}

func TestDeleteBorrowSucc(t *testing.T) {
	store := mem.New()

	b := buildBlock(0)
	for i := 32; i < 64; i++ {
		(*b)[i] = 0
	}

	d1, _ := store.AddBlock(b)

	d2, _ := store.AddBlock(buildBlock(32))
	start, _ := store.AddBlock(&block.Block{0, d1, 32, d2})
	tree := New(store, 1, start)

	tree.Delete(10)

	if _, err := tree.Get(10); !errors.Is(err, ErrNotFound) {
		t.Fail()
	}
}

func TestDeleteMergePre(t *testing.T) {
	store := mem.New()

	b1 := buildBlock(0)
	for i := 32; i < 64; i++ {
		(*b1)[i] = 0
	}
	d1, _ := store.AddBlock(b1)

	b2 := buildBlock(32)
	for i := 32; i < 64; i++ {
		(*b2)[i] = 0
	}
	d2, _ := store.AddBlock(b2)

	start, _ := store.AddBlock(&block.Block{0, d1, 32, d2})
	tree := New(store, 1, start)

	tree.Delete(35)

	if _, err := tree.Get(35); !errors.Is(err, ErrNotFound) {
		t.Log(store)
		t.Fail()
	}
}

func TestDeleteMergeSucc(t *testing.T) {
	store := mem.New()

	b1 := buildBlock(0)
	for i := 32; i < 64; i++ {
		(*b1)[i] = 0
	}
	d1, _ := store.AddBlock(b1)

	b2 := buildBlock(32)
	for i := 32; i < 64; i++ {
		(*b2)[i] = 0
	}
	d2, _ := store.AddBlock(b2)

	start, _ := store.AddBlock(&block.Block{0, d1, 32, d2})
	tree := New(store, 1, start)

	tree.Delete(10)

	if _, err := tree.Get(10); !errors.Is(err, ErrNotFound) {
		t.Log(store)
		t.Fail()
	}
}
