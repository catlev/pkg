package path

import (
	"errors"
	"strconv"
)

var ErrInvalidPath = errors.New("invalid path")

type Visitor[T any] interface {
	String(x string) (T, error)
	Integer(x int) (T, error)
	Rel(name string) (T, error)
	Op(name string, args []T) (T, error)
}

func Visit[T any](v Visitor[T], x Expr) (T, error) {
	switch x.Kind {
	case Integer:
		x, err := strconv.Atoi(x.Value)
		if err != nil {
			return fail[T](err)
		}
		return v.Integer(x)

	case String:
		x, err := strconv.Unquote(x.Value)
		if err != nil {
			return fail[T](err)
		}
		return v.String(x)

	case Rel:
		return v.Rel(x.Value)

	case Op:
		args := make([]T, len(x.Children))
		for i, a := range x.Children {
			x, err := Visit(v, a)
			if err != nil {
				return fail[T](err)
			}
			args[i] = x
		}
		return v.Op(x.Value, args)
	}

	return fail[T](ErrInvalidPath)
}

func fail[T any](err error) (T, error) {
	var z T
	return z, err
}
