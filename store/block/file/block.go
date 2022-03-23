package file

import (
	"io"
	"io/fs"

	"github.com/catlev/pkg/store/block"
)

type Store struct {
	f     File
	maxID block.Word
	free  block.Word
}

type File interface {
	io.ReaderAt
	io.WriterAt

	Stat() (fs.FileInfo, error)
}

func New(f File) (*Store, error) {
	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}
	return &Store{
		f:     f,
		maxID: block.Word(fi.Size() - block.ByteSize),
	}, nil
}

func (s *Store) ReadBlock(id block.Word, b *block.Block) error {
	if id > s.maxID {
		return io.ErrUnexpectedEOF
	}

	_, err := s.f.ReadAt(b.Bytes(), int64(id))
	return err
}

func (s *Store) WriteBlock(id block.Word, b *block.Block) (block.Word, error) {
	if id > s.maxID {
		return 0, io.ErrUnexpectedEOF
	}

	_, err := s.f.WriteAt(b.Bytes(), int64(id))
	return id, err
}

func (s *Store) AddBlock(b *block.Block) (block.Word, error) {
	if s.free == 0 {
		s.maxID += block.ByteSize
		_, err := s.f.WriteAt(b.Bytes(), int64(s.maxID))
		return s.maxID, err
	}
	var bb block.Block
	id := s.free
	err := s.ReadBlock(id, &bb)
	if err != nil {
		return 0, err
	}
	s.free = bb[0]
	_, err = s.f.WriteAt(b.Bytes(), int64(s.maxID))
	return id, err
}

func (s *Store) FreeBlock(id block.Word) error {
	var b block.Block
	b[0] = s.free
	_, err := s.WriteBlock(id, &b)
	if err != nil {
		return err
	}
	s.free = id
	return nil
}
