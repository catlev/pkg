package mem

import (
	"io"

	"github.com/catlev/pkg/store/block"
)

type Store struct {
	blocks []block.Block
}

func New() *Store {
	return &Store{make([]block.Block, 1)}
}

func (s *Store) ReadBlock(id block.Word, b *block.Block) error {
	idx := int(id / block.ByteSize)

	if idx >= len(s.blocks) {
		return io.ErrUnexpectedEOF
	}

	copy((*b)[:], s.blocks[idx][:])
	return nil
}

func (s *Store) AddBlock(b *block.Block) (block.Word, error) {
	idx := len(s.blocks)
	s.blocks = append(s.blocks, *b)
	return block.Word(idx * block.ByteSize), nil
}

func (s *Store) WriteBlock(id block.Word, b *block.Block) (block.Word, error) {
	idx := id / block.ByteSize
	s.blocks[idx] = *b
	return id, nil
}

func (s *Store) FreeBlock(id block.Word) error {
	return nil
}
