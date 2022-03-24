package tree

import (
	"testing"

	"github.com/catlev/pkg/store/block"
	"github.com/catlev/pkg/store/block/mem"
	"github.com/stretchr/testify/assert"
)

func TestPutUpdate(t *testing.T) {
	store := mem.New()
	d1, _ := store.AddBlock(&block.Block{0, 1, 10, 2, 20, 3})
	d2, _ := store.AddBlock(&block.Block{0, 4, 40, 5, 50, 6})
	start, _ := store.AddBlock(&block.Block{0, d1, 30, d2})
	tree := New(2, 0, store, 1, start)

	tree.Put([]block.Word{20, 7})
	row, _ := tree.Get(20)

	assert.Equal(t, block.Word(7), row[1])
}

func TestPutAddWithRoom(t *testing.T) {
	store := mem.New()
	d1, _ := store.AddBlock(&block.Block{0, 1, 10, 2, 20, 3})
	d2, _ := store.AddBlock(&block.Block{0, 4, 40, 5, 50, 6})
	start, _ := store.AddBlock(&block.Block{0, d1, 30, d2})
	tree := New(2, 0, store, 1, start)

	tree.Put([]block.Word{25, 7})
	row, _ := tree.Get(25)

	assert.Equal(t, block.Word(7), row[1])
}

func TestPutAddRoot(t *testing.T) {
	store := mem.New()
	var b block.Block
	for i := range b {
		b[i] = block.Word(i)
	}
	start, _ := store.AddBlock(&b)
	tree := New(2, 0, store, 0, start)

	tree.Put([]block.Word{25, 100})
	row, _ := tree.Get(25)

	assert.Equal(t, block.Word(100), row[1])
}

func TestPutAddRootLarge(t *testing.T) {
	store := mem.New()
	var b block.Block
	for i := range b {
		b[i] = block.Word(i)
	}
	start, _ := store.AddBlock(&b)
	tree := New(2, 0, store, 0, start)

	tree.Put([]block.Word{45, 100})
	row, _ := tree.Get(45)

	assert.Equal(t, block.Word(100), row[1])
}

func TestPutDeep(t *testing.T) {
	store := mem.New()
	start, _ := store.AddBlock(&block.Block{})
	tree := New(2, 0, store, 0, start)

	for i := 0; i < 100; i++ {
		tree.Put([]block.Word{block.Word(i), block.Word(i * 2)})
	}

	row, _ := tree.Get(50)

	assert.Equal(t, block.Word(100), row[1])
}

type appendingMemStore struct {
	mem.Store
}

func (s *appendingMemStore) WriteBlock(id block.Word, b *block.Block) (block.Word, error) {
	return s.AddBlock(b)
}

func TestPutNewBlock(t *testing.T) {
	store := &appendingMemStore{*mem.New()}
	d1, _ := store.AddBlock(&block.Block{0, 1, 10, 2, 20, 3})
	d2, _ := store.AddBlock(&block.Block{0, 4, 40, 5, 50, 6})
	start, _ := store.AddBlock(&block.Block{0, d1, 30, d2})
	tree := New(2, 0, store, 1, start)

	tree.Put([]block.Word{20, 7})
	row, _ := tree.Get(20)

	assert.Equal(t, block.Word(7), row[1])
	assert.NotEqual(t, start, tree.Root())
}
