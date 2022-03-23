package tree

import (
	"fmt"

	"github.com/catlev/pkg/store/block"
)

type TreeError struct {
	Op  string
	Key block.Word
	Err error
}

func (e *TreeError) Error() string {
	return fmt.Sprintf("%s %d: %s", e.Op, e.Key, e.Err.Error())
}

func (e *TreeError) Unwrap() error {
	return e.Err
}

func wrapErr(err *error, op string, key block.Word) {
	if *err == nil {
		return
	}
	*err = &TreeError{
		Op:  op,
		Key: key,
		Err: *err,
	}
}
