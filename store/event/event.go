package event

import (
	"errors"

	"github.com/catlev/pkg/store/block"
	"github.com/catlev/pkg/store/tree"
)

var ErrVFM = errors.New("vfm error")

type Event struct {
	stream    EventSource
	base      block.Store
	vfm       tree.Tree
	threshold block.Word
	free      block.Word
}

type EventSource interface {
	GetEvent(id block.Word) (block.Store, error)
}

func New(s EventSource, b block.Store) *Event {
	return (&Event{
		stream: s,
		base:   b,
	}).init()
}

func (e *Event) init() *Event {
	return e
}

func (e *Event) ReadBlock(id block.Word, block *block.Block) error {
	if id >= e.threshold {
		return e.base.ReadBlock(id-e.threshold, block)
	}
	r := e.vfm.GetRange(id)
	if !r.Next() {
		return ErrVFM
	}
	prev, err := e.stream.GetEvent(r.This()[1])
	if err != nil {
		return err
	}
	return prev.ReadBlock(id-r.This()[0], block)
}

func (e *Event) AddBlock(b *block.Block) (block.Word, error) {
	if e.free == 0 {
		return e.base.AddBlock(b)
	}
	var f block.Block
	err := e.base.ReadBlock(e.free, &f)
	if err != nil {
		return 0, err
	}
	next := e.free
	e.free = f[0]
	return e.base.WriteBlock(next, b)
}

func (e *Event) WriteBlock(id block.Word, block *block.Block) (block.Word, error) {
	if id >= e.threshold {
		return e.base.WriteBlock(id-e.threshold, block)
	}
	return e.addThreshold(e.base.AddBlock(block))
}

func (e *Event) FreeBlock(id block.Word) error {
	if id < e.threshold {
		return nil
	}
	if e.free != 0 {
		_, err := e.base.WriteBlock(id-e.threshold, &block.Block{0: e.free})
		if err != nil {
			return err
		}
	}
	e.free = id - e.threshold
	return nil
}

func (e *Event) addThreshold(id block.Word, err error) (block.Word, error) {
	return id + e.threshold, err
}
