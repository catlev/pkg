package tree

import (
	"math/rand"
	"testing"
)

func TestTreeOperations(t *testing.T) {
	rand.Seed(0)
	tree := New(NewBlankMemoryStore(), 0, blockSize)

	for i := 1; i < 4096; i++ {
		err := tree.Put(rand.Uint64(), uint64(i))

		if err != nil {
			t.Fatalf("got unexpected error %v", err)
		}
	}

	err := tree.Put(13, 1234)
	if err != nil {
		t.Fatalf("got unexpected error %v", err)
	}

	v, err := tree.Get(13)
	if err != nil {
		t.Fatalf("got unexpected error %v", err)
	}
	if v != 1234 {
		t.Errorf("got %v, expecting 1234", v)
	}
}

func TestRange(t *testing.T) {
	tree := New(NewBlankMemoryStore(), 0, blockSize)

	for i := 20; i < 4096; i++ {
		err := tree.Put(uint64(i), uint64(i))

		if err != nil {
			t.Fatalf("got unexpected error %v", err)
		}
	}

	for _, test := range []struct {
		name     string
		from, to uint64
		min, max uint64
	}{
		{
			"EntirelyWithin",
			100, 200,
			100, 200,
		},
		{
			"MissingMinimum",
			0, 100,
			20, 100,
		},
		{
			"MissingMaximum",
			2048, 5000,
			2048, 4096,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			r := tree.GetRange(test.from, test.to)
			expect := test.min
			for r.Next() {
				if r.This() != expect {
					t.Log(tree.lookupNode(113))
					t.Fatalf("got %v, expecting %v", r.This(), expect)
				}
				expect++
			}
			if expect != test.max {
				t.Fatalf("got %v, expecting %v", r.This(), test.max)
			}
			if r.Err() != nil {
				t.Fatalf("got unexpected error %v", r.Err())
			}
		})
	}
}
