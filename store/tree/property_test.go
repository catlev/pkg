package tree

import (
	"testing"
	"testing/quick"

	"github.com/catlev/pkg/store/block"
	"github.com/catlev/pkg/store/block/mem"
)

func TestTreeProperties(t *testing.T) {
	store := mem.New()
	start, _ := store.AddBlock(&block.Block{})
	tree := New(3, 0, store, 0, start)

	quick.Check(func(key, value1, value2 block.Word) bool {

		err := tree.Put([]block.Word{key, value1, value2})
		if err != nil {
			return false
		}

		v, err := tree.Get(key)
		if err != nil {
			return false
		}

		return v[1] == value1 && v[2] == value2

	}, &quick.Config{
		// need to do it a lot of times to get some depth in the tree
		MaxCount: 200 * 1024,
	})

}
