package mem

import (
	"io"

	"github.com/catlev/pkg/domain"
)

type Store struct {
	blocks []domain.Word
	free   []domain.Word
}

func New() *Store {
	return &Store{make([]domain.Word, domain.WordSize), nil}
}

func (s *Store) ReadBlock(id domain.Word, b *domain.Block) error {
	if int(id) > len(s.blocks)-domain.WordSize {
		return io.ErrUnexpectedEOF
	}

	copy((*b)[:], s.blocks[id:])
	return nil
}

func (s *Store) AddBlock(b *domain.Block) (domain.Word, error) {
	var id domain.Word

	if len(s.free) != 0 {
		id = s.free[len(s.free)-1]
		s.free = s.free[:len(s.free)-1]
		copy(s.blocks[id:], (*b)[:])
	} else {
		id = domain.Word(len(s.blocks))
		s.blocks = append(s.blocks, (*b)[:]...)
	}

	return id, nil
}

func (s *Store) WriteBlock(id domain.Word, b *domain.Block) (domain.Word, error) {
	copy(s.blocks[id:], (*b)[:])
	return id, nil
}

func (s *Store) FreeBlock(id domain.Word) error {
	s.free = append(s.free, id)
	return nil
}
