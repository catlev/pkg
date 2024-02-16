package tree

import (
	"testing"

	"github.com/catlev/pkg/store/block"
	"github.com/catlev/pkg/store/block/mem"
)

func assertTreeProperty(t *testing.T, i int, k block.Word) {
	t.Helper()
	if int(k) != (i/10)+1 {
		t.Errorf("failed on %d: %d != %d", i, k, (i/10)+1)
	}
}

func buildBlock(off int) *block.Block {
	var b block.Block
	for i := 0; i < 32; i++ {
		n := i + off
		b[i*2] = block.Word(n)
		b[i*2+1] = block.Word((n / 10) + 1)
	}
	return &b
}

func buildBlock2(off int) *block.Block {
	var b block.Block
	for i := 0; i < 16; i++ {
		n := i + off
		b[i*4] = block.Word(n)
		b[i*4+1] = block.Word(n * 2)
		b[i*4+2] = block.Word(n)
		b[i*4+3] = block.Word((n / 10) + 1)
	}
	return &b
}

func TestCompare(t *testing.T) {
	for _, test := range []struct {
		name        string
		left, right []block.Word
		expected    int
	}{
		{"Empty", nil, nil, 0},
		{"OneSame", []block.Word{10}, []block.Word{10}, 0},
		{"TwoSame", []block.Word{10, 15}, []block.Word{10, 15}, 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			actual := compareValues(test.left, test.right)
			if actual != test.expected {
				t.Errorf("got %d, expected %d", actual, test.expected)
			}
		})
	}
}

func TestProbe(t *testing.T) {
	store := mem.New()
	start, _ := store.AddBlock(buildBlock(0))
	tree := New(2, 1, store, 0, start)
	node, _ := tree.findNode([]block.Word{0})

	t.Log(node)

	for i := 0; i < 25; i++ {
		k := node.probe([]block.Word{block.Word(i)})
		if i != k {
			t.Errorf("got %d, expected %d", k, i)
		}
	}
}

func TestWideProbe(t *testing.T) {
	store := mem.New()
	start, _ := store.AddBlock(buildBlock2(0))
	tree := New(4, 2, store, 0, start)
	node, _ := tree.findNode([]block.Word{0, 0})

	t.Log(node)

	for i := 0; i < 16; i++ {
		k := node.probe([]block.Word{block.Word(i), block.Word(i * 2)})
		if i != k {
			t.Errorf("got %d, expected %d", k, i)
		}
	}
}

func TestGetRange(t *testing.T) {
	store := mem.New()
	d1, _ := store.AddBlock(buildBlock(0))
	d2, _ := store.AddBlock(buildBlock(32))
	start, _ := store.AddBlock(&block.Block{0, d1, 32, d2})
	tree := New(2, 1, store, 1, start)

	for i := 0; i < 64; i++ {
		r := tree.GetRange([]block.Word{block.Word(i)})
		j := i

		for r.Next() {
			assertTreeProperty(t, j, r.This()[1])
			j++
		}

		if j != 64 {
			t.Errorf("start %d end %d", i, j)
		}
	}
}

func TestWideKey(t *testing.T) {
	store := mem.New()
	d1, _ := store.AddBlock(buildBlock2(0))
	d2, _ := store.AddBlock(buildBlock2(16))
	start, _ := store.AddBlock(&block.Block{0, 0, d1, 16, 32, d2})
	tree := New(4, 2, store, 1, start)

	for i := 0; i < 32; i++ {
		r := tree.GetRange([]block.Word{block.Word(i), block.Word(i * 2)})
		j := i

		for r.Next() {
			assertTreeProperty(t, j, r.This()[3])
			j++
		}

		if j != 32 {
			t.Errorf("start %d end %d", i, j)
		}
	}
}

func TestWideKeyPartial(t *testing.T) {
	store := mem.New()
	d1, _ := store.AddBlock(buildBlock2(0))
	d2, _ := store.AddBlock(buildBlock2(16))
	start, _ := store.AddBlock(&block.Block{0, 0, d1, 16, 32, d2})
	tree := New(4, 2, store, 1, start)

	for i := 0; i < 32; i++ {
		r := tree.GetRange([]block.Word{block.Word(i), block.Word(i * 2)})
		j := i

		for r.Next() {
			assertTreeProperty(t, j, r.This()[3])
			j++
		}

		if j != 32 {
			t.Errorf("start %d end %d", i, j)
		}
	}
}
