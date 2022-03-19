package block

import (
	"reflect"
	"unsafe"
)

const ByteSize = 512
const WordSize = 64

type Word uint64

type Block [64]Word

type Reader interface {
	ReadBlock(id Word, block *Block) error
}

type Adder interface {
	AddBlock(block *Block) (Word, error)
}

// Writer provides the ability to write a block to a store. It returns the ID of the block that was
// actually written to. Due to how transactions are managed, this may not be the same as the ID
// passed in.
type Writer interface {
	WriteBlock(id Word, block *Block) (Word, error)
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
	sh := (*reflect.SliceHeader)(unsafe.Pointer(b))
	ws := int(unsafe.Sizeof(Word(0)))

	// b is always in scope, so this is safe to do
	rsh := &reflect.SliceHeader{
		Data: sh.Data,
		Len:  sh.Len * ws,
		Cap:  sh.Cap * ws,
	}

	return *(*[]byte)(unsafe.Pointer(rsh))
}
