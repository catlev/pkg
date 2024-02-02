package block

import (
	"unsafe"
)

const ByteSize = 512
const WordSize = 64

type Word uint64

type Block [WordSize]Word

type Reader interface {
	ReadBlock(id Word, b *Block) error
}

type Adder interface {
	AddBlock(b *Block) (Word, error)
}

// Writer provides the ability to write a block to a store. It returns the ID of the block that was
// actually written to. Due to how transactions are managed, this may not be the same as the ID
// passed in.
type Writer interface {
	WriteBlock(id Word, b *Block) (Word, error)
}

type Freer interface {
	FreeBlock(id Word) error
}

type Store interface {
	Reader
	Adder
	Writer
	Freer
}

func (b *Block) Bytes() []byte {
	return unsafe.Slice((*byte)(unsafe.Pointer(b)), ByteSize)
}
