package domain

import "unsafe"

const ByteSize = 512
const WordSize = 64

type Word uint64

type Block [WordSize]Word

type Store interface {
	ReadBlock(id Word, b *Block) error
	AddBlock(b *Block) (Word, error)
	WriteBlock(id Word, b *Block) (Word, error)
	FreeBlock(id Word) error
}

func (b *Block) Bytes() []byte {
	return unsafe.Slice((*byte)(unsafe.Pointer(b)), ByteSize)
}
