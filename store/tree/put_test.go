package tree

import (
	"testing"

	"github.com/catlev/pkg/domain"
	"github.com/catlev/pkg/store/block/mem"
	"github.com/stretchr/testify/assert"
)

func getRow(t *testing.T, tree *Tree, key []domain.Word) []domain.Word {
	t.Helper()
	r, err := tree.Get(key)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestPutUpdate(t *testing.T) {
	store := mem.New()
	d1, _ := store.AddBlock(&domain.Block{0, 1, 10, 2, 20, 3})
	d2, _ := store.AddBlock(&domain.Block{0, 4, 40, 5, 50, 6})
	start, _ := store.AddBlock(&domain.Block{0, d1, 30, d2})
	tree := New(2, 1, store, 1, start)

	tree.Put([]domain.Word{20, 7})
	row := getRow(t, tree, []domain.Word{20})

	assert.Equal(t, domain.Word(7), row[1])
}

func TestPutAddWithRoom(t *testing.T) {
	store := mem.New()
	d1, _ := store.AddBlock(&domain.Block{0, 1, 10, 2, 20, 3})
	d2, _ := store.AddBlock(&domain.Block{0, 4, 40, 5, 50, 6})
	start, _ := store.AddBlock(&domain.Block{0, d1, 30, d2})
	tree := New(2, 1, store, 1, start)

	tree.Put([]domain.Word{25, 7})
	row := getRow(t, tree, []domain.Word{25})

	assert.Equal(t, domain.Word(7), row[1])
}

func TestPutAddRoot(t *testing.T) {
	store := mem.New()
	var b domain.Block
	for i := range b {
		b[i] = domain.Word(i)
	}
	start, _ := store.AddBlock(&b)
	tree := New(2, 1, store, 0, start)

	tree.Put([]domain.Word{25, 100})
	row := getRow(t, tree, []domain.Word{25})

	assert.Equal(t, domain.Word(100), row[1])
}

func TestPutAddRootLarge(t *testing.T) {
	store := mem.New()
	var b domain.Block
	for i := range b {
		b[i] = domain.Word(i)
	}
	start, _ := store.AddBlock(&b)
	tree := New(2, 1, store, 0, start)

	tree.Put([]domain.Word{45, 100})
	row := getRow(t, tree, []domain.Word{45})

	assert.Equal(t, domain.Word(100), row[1])
}

func TestPutDeep(t *testing.T) {
	store := mem.New()
	start, _ := store.AddBlock(&domain.Block{})
	tree := New(2, 1, store, 0, start)

	for i := 0; i < 100; i++ {
		tree.Put([]domain.Word{domain.Word(i), domain.Word(i * 2)})
	}

	row := getRow(t, tree, []domain.Word{50})

	assert.Equal(t, domain.Word(100), row[1])
}

type appendingMemStore struct {
	mem.Store
}

func (s *appendingMemStore) WriteBlock(id domain.Word, b *domain.Block) (domain.Word, error) {
	return s.AddBlock(b)
}

func TestPutNewBlock(t *testing.T) {
	store := &appendingMemStore{*mem.New()}
	d1, _ := store.AddBlock(&domain.Block{0, 1, 10, 2, 20, 3})
	d2, _ := store.AddBlock(&domain.Block{0, 4, 40, 5, 50, 6})
	start, _ := store.AddBlock(&domain.Block{0, d1, 30, d2})
	tree := New(2, 1, store, 1, start)

	tree.Put([]domain.Word{20, 7})
	row := getRow(t, tree, []domain.Word{20})

	assert.Equal(t, domain.Word(7), row[1])
	assert.NotEqual(t, start, tree.Root())
}
