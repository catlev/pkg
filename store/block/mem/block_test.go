package mem

import (
	"testing"

	"github.com/catlev/pkg/domain"
)

func TestReadEOF(t *testing.T) {
	store := New()
	store.blocks[2] = 4

	var b domain.Block
	err := store.ReadBlock(64, &b)

	if err == nil {
		t.Fail()
	}
}

func TestRead(t *testing.T) {
	store := New()
	store.blocks[2] = 4

	var b domain.Block
	store.ReadBlock(0, &b)

	if b[2] != 4 {
		t.Fail()
	}
}

func TestAdd(t *testing.T) {
	store := New()

	var b domain.Block
	b[2] = 4
	id, _ := store.AddBlock(&b)
	b[2] = 0
	store.ReadBlock(id, &b)

	if b[2] != 4 {
		t.Fail()
	}
}

func TestWrite(t *testing.T) {
	store := New()

	var b domain.Block
	b[2] = 4
	store.WriteBlock(0, &b)
	b[2] = 0
	store.ReadBlock(0, &b)

	if b[2] != 4 {
		t.Fail()
	}
}

func TestFree(t *testing.T) {
	store := New()

	store.FreeBlock(0)
	id, err := store.AddBlock(new(domain.Block))

	if err != nil {
		t.Error(err)
	}
	if id != 0 {
		t.Fail()
	}
}
