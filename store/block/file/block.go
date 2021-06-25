package file

import (
	"encoding/binary"
	"github.com/catlev/pkg/store/block"
	"io"
	"os"
)

type Store struct {
	f    *os.File
	maxID block.Word
}

func New(f *os.File) (*Store, error) {
	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}
	return &Store{f, block.Word(fi.Size() - block.ByteSize)}, nil
}

func (s *Store) ReadBlock(id block.Word, b *block.Block) error {
	if id > s.maxID {
		return io.ErrUnexpectedEOF
	}

	if _, err := s.f.Seek(int64(id), io.SeekStart); err != nil {
		return err
	}

	return binary.Read(s.f, binary.LittleEndian, b)
}

func (s *Store) WriteBlock(b *block.Block) (block.Word, error) {
	id, err := s.f.Seek(0, io.SeekEnd)
	if err != nil {
		return 0, err
	}

	err = binary.Write(s.f, binary.LittleEndian, *b)
	if err != nil {
		return 0, err
	}

	return block.Word(id), nil
}
