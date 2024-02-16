package tree

import (
	"strconv"
	"testing"

	"github.com/catlev/pkg/store/block"
	"github.com/catlev/pkg/store/block/mem"
)

func assertDeletionSuccess(t *testing.T, tree *Tree, min, max, without block.Word) {
	t.Helper()
	tree.Delete([]block.Word{without})
	for k := min; k < max; k++ {
		t.Run(strconv.Itoa(int(k)), func(t *testing.T) {
			t.Helper()

			v, err := tree.Get([]block.Word{k})

			if (err != nil) != (k == without) {
				t.Error(err)
				return
			}

			if k != without && v[1] != (k/10)+1 {
				t.Fail()
			}
		})
	}
}

func TestDeleteLeafSimple(t *testing.T) {
	store := mem.New()
	start, _ := store.AddBlock(buildBlock(0))
	tree := New(2, 1, store, 0, start)

	assertDeletionSuccess(t, tree, 0, 32, 10)
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
	tree := New(2, 1, store, 1, start)

	assertDeletionSuccess(t, tree, 0, 48, 35)
}

func TestDeleteBorrowSucc(t *testing.T) {
	store := mem.New()

	b := buildBlock(0)
	for i := 32; i < 64; i++ {
		(*b)[i] = 0
	}

	d1, _ := store.AddBlock(b)

	d2, _ := store.AddBlock(buildBlock(16))
	start, _ := store.AddBlock(&block.Block{0, d1, 16, d2})
	tree := New(2, 1, store, 1, start)

	assertDeletionSuccess(t, tree, 0, 48, 10)
}

func TestDeleteMergePre(t *testing.T) {
	store := mem.New()

	b1 := buildBlock(0)
	for i := 32; i < 64; i++ {
		(*b1)[i] = 0
	}
	d1, _ := store.AddBlock(b1)

	b2 := buildBlock(16)
	for i := 32; i < 64; i++ {
		(*b2)[i] = 0
	}
	d2, _ := store.AddBlock(b2)

	start, _ := store.AddBlock(&block.Block{0, d1, 16, d2})
	tree := New(2, 1, store, 1, start)

	assertDeletionSuccess(t, tree, 0, 32, 20)
}

func TestDeleteMergeSucc(t *testing.T) {
	store := mem.New()

	b1 := buildBlock(0)
	for i := 32; i < 64; i++ {
		(*b1)[i] = 0
	}
	d1, _ := store.AddBlock(b1)

	b2 := buildBlock(16)
	for i := 32; i < 64; i++ {
		(*b2)[i] = 0
	}
	d2, _ := store.AddBlock(b2)

	start, _ := store.AddBlock(&block.Block{0, d1, 16, d2})
	tree := New(2, 1, store, 1, start)

	assertDeletionSuccess(t, tree, 0, 32, 10)
}
