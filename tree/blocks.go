package tree

import (
	"io"
)

const blockSize = 512

// BlockStore represents a data store for a B-tree.
type BlockStore interface {
	ReadBlock(at uint64, buf []byte) error
	WriteBlock(at uint64, buf []byte) error
	NewBlock() (uint64, error)
}

// MemoryStore implements a block store in-memory.
type MemoryStore struct {
	blocks []byte
}

// NewBlankMemoryStore returns an uninitialised in-memory block store. This could be useful for testing and for making
// in-memory caches.
func NewBlankMemoryStore() *MemoryStore {
	return &MemoryStore{make([]byte, blockSize*2)}
}

func (m *MemoryStore) ReadBlock(at uint64, buf []byte) error {
	if at + blockSize > uint64(len(m.blocks)) {
		return io.ErrUnexpectedEOF
	}
	copy(buf, m.blocks[at:at+blockSize])
	return nil
}

func (m *MemoryStore) WriteBlock(at uint64, buf []byte) error {
	if at + blockSize > uint64(len(m.blocks)) {
		return io.ErrUnexpectedEOF
	}
	copy(m.blocks[at:at+blockSize], buf)
	return nil
}

func (m *MemoryStore) NewBlock() (uint64, error) {
	id := uint64(len(m.blocks))
	m.blocks = append(m.blocks, make([]byte, blockSize)...)
	return id, nil
}



