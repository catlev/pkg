package mem

import (
	"testing"

	"github.com/catlev/pkg/store/block"
)

func TestRead(t *testing.T) {
	store := New()
	store.blocks[2] = 4

	var b block.Block
	store.ReadBlock(0, &b)

	if b[2] != 4 {
		t.Fail()
	}
}

func TestAdd(t *testing.T) {
	store := New()

	var b block.Block
	b[2] = 4
	id, _ := store.AddBlock(&b)
	b[2] = 0
	store.ReadBlock(id, &b)

	if b[2] != 4 {
		t.Fail()
	}
}
