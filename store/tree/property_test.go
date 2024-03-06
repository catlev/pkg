//go:build property

package tree

import (
	"testing"
	"testing/quick"

	"github.com/catlev/pkg/domain"
	"github.com/catlev/pkg/store/block/mem"
)

func TestTreeProperties(t *testing.T) {
	store := mem.New()
	start, _ := store.AddBlock(&domain.Block{})
	tree := New(3, 1, store, 0, start)

	if err := quick.Check(func(key, value1, value2 domain.Word) bool {

		err := tree.Put([]domain.Word{key, value1, value2})
		if err != nil {
			return false
		}

		row, err := tree.Get([]domain.Word{key})
		if err != nil {
			return false
		}

		return row[1] == value1 && row[2] == value2

	}, &quick.Config{
		// need to do it a lot of times to get some depth in the tree
		MaxCount: 200 * 1024,
	}); err != nil {
		t.Error(err)
	}

}

func TestWideTreeProperties(t *testing.T) {
	store := mem.New()
	start, _ := store.AddBlock(&domain.Block{})
	tree := New(4, 2, store, 0, start)

	if err := quick.Check(func(key1, key2, value1, value2 domain.Word) bool {

		err := tree.Put([]domain.Word{key1, key2, value1, value2})
		if err != nil {
			return false
		}

		row, err := tree.Get([]domain.Word{key1, key2})
		if err != nil {
			return false
		}

		return row[2] == value1 && row[3] == value2

	}, &quick.Config{
		// need to do it a lot of times to get some depth in the tree
		MaxCount: 200 * 1024,
	}); err != nil {
		t.Error(err)
	}

}
