package tree

import (
	"errors"
	"fmt"

	"github.com/catlev/pkg/store/block"
)

var (
	ErrNotFound = errors.New("not found")
	ErrBadRow   = errors.New("bad row")
	ErrKeyWidth = errors.New("wrong number of values in key")
)

type TreeError struct {
	Op  string
	Key []block.Word
	Err error
}

func (e *TreeError) Error() string {
	return fmt.Sprintf("%s %d: %s", e.Op, e.Key, e.Err.Error())
}

func (e *TreeError) Unwrap() error {
	return e.Err
}

func wrapErr(err *error, op string, key []block.Word) {
	if *err == nil {
		return
	}
	*err = &TreeError{
		Op:  op,
		Key: key,
		Err: *err,
	}
}
