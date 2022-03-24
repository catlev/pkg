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
	tree := New(store, 0, start)

	quick.Check(func(key, value block.Word) bool {

		err := tree.Put(key, value)
		if err != nil {
			return false
		}

		v, err := tree.Get(key)
		if err != nil {
			return false
		}

		return v == value

	}, &quick.Config{
		MaxCount: 200 * 1024,
	})

}
