package mem

import (
	"io"

	"github.com/catlev/pkg/store/block"
)

type Store struct {
	blocks []block.Word
	free   []block.Word
}

func New() *Store {
	return &Store{make([]block.Word, block.WordSize), nil}
}

func (s *Store) ReadBlock(id block.Word, b *block.Block) error {
	if int(id) > len(s.blocks)-block.WordSize {
		return io.ErrUnexpectedEOF
	}

	copy((*b)[:], s.blocks[id:])
	return nil
}

func (s *Store) AddBlock(b *block.Block) (block.Word, error) {
	var id block.Word

	if len(s.free) != 0 {
		id = s.free[len(s.free)-1]
		s.free = s.free[:len(s.free)-1]
		copy(s.blocks[id:], (*b)[:])
	} else {
		id = block.Word(len(s.blocks))
		s.blocks = append(s.blocks, (*b)[:]...)
	}

	return id, nil
}

func (s *Store) WriteBlock(id block.Word, b *block.Block) (block.Word, error) {
	copy(s.blocks[id:], (*b)[:])
	return id, nil
}

func (s *Store) FreeBlock(id block.Word) error {
	s.free = append(s.free, id)
	return nil
}
