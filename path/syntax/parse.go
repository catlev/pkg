package syntax

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"text/scanner"
	"unsafe"

	"github.com/catlev/pkg/path"
)

// ErrFailedToMatch indicates that there is a syntax error in the path.
var ErrFailedToMatch = errors.New("failed to parse path")

// ParseString parses a string into a path.Expr.
func ParseString(s string) (path.Expr, error) {
	return Parse(strings.NewReader(s))
}

// Parse parses a Reader into a path.Expr.
func Parse(r io.Reader) (path.Expr, error) {
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return path.Expr{}, err
	}
	s := new(scanner.Scanner).Init(bytes.NewReader(buf))
	p := parser{buf: buf, s: *s}
	p.s.Error = func(s *scanner.Scanner, err string) {
		p.err = fmt.Errorf("%w: %s", ErrFailedToMatch, err)
	}
	res := p.parse(0)
	return res, p.err
}

// Information required to parse a path expression, including scanning the
// source and producing the results.
type parser struct {
	buf    []byte
	s      scanner.Scanner
	tok    rune
	cached bool
	err    error
}

// Set the parser's error if not already set.
func (p *parser) setErr(err error) {
	if p.err == nil {
		p.err = err
	}
}

// Parse an expression at a particular precedence level.
func (p *parser) parse(prec int) path.Expr {
	left := p.prefix()
	for {
		if p.err != nil || p.tok == scanner.EOF {
			break
		}
		right := p.infix(prec, left)
		if right.Kind == path.Invalid {
			break
		}
		left = right
	}
	return left
}

// Parse an expression in prefix position.
func (p *parser) prefix() path.Expr {
	switch p.next() {
	case scanner.String:
		start, end := p.pos()
		return p.leaf(path.String, start, end)
	case scanner.Int:
		start, end := p.pos()
		return p.leaf(path.Integer, start, end)
	case scanner.Ident, '*':
		start, end := p.pos()
		return p.leaf(path.Rel, start, end)
	case '~':
		return p.branch(path.Op, "inverse", p.parse(100))
	case '(':
		path := p.parse(0)
		if p.next() != ')' {
			p.setErr(fmt.Errorf("expecting ')'"))
		}
		return path
	}
	p.setErr(ErrFailedToMatch)
	return path.Expr{}
}

// Infix parsing productions.
var infixes = [256]struct {
	op           string
	outer, inner int
}{
	'/': {"join", 80, 80},
	'&': {"intersection", 70, 70},
	'|': {"union", 60, 60},
}

// Parse an expression in infix position.
func (p *parser) infix(prec int, left path.Expr) path.Expr {
	var op string
	var outer, inner int
	c := p.peek()
	if c == '(' {
		p.next()
		return p.namedOperation(left.Value)
	}
	if c >= 0 && c < 256 {
		inf := infixes[c]
		op = inf.op
		outer = inf.outer
		inner = inf.inner
	}
	if p.err != nil || op == "" || prec >= outer {
		return path.Expr{}
	}
	p.next()
	return p.branch(path.Op, op, left, p.parse(inner))
}

func (p *parser) namedOperation(name string) path.Expr {
	res := path.Expr{Kind: path.Op, Value: name}
	for {
		c := p.peek()
		if c == ')' {
			p.next()
			return res
		}
		if c == ',' && len(res.Children) > 0 {
			p.next()
		}
		res.Children = append(res.Children, p.prefix())
	}
}

// Look at the next token without advancing position.
func (p *parser) peek() rune {
	p.tok = p.next()
	p.cached = true
	return p.tok
}

// Look at the next token.
func (p *parser) next() rune {
	if p.err != nil {
		return 0
	}
	if p.cached {
		p.cached = false
		return p.tok
	}
	p.tok = p.s.Scan()
	return p.tok
}

// Get the (start,end) of the current token.
func (p *parser) pos() (int, int) {
	off := p.s.Pos().Offset
	return off - len(p.s.TokenText()), off
}

func (p *parser) leaf(kind path.Kind, start, end int) path.Expr {
	return path.Expr{Kind: kind, Value: p.value(start, end)}
}

func (p *parser) branch(kind path.Kind, name string, children ...path.Expr) path.Expr {
	return path.Expr{Kind: kind, Value: name, Children: children}
}

func (p *parser) value(start, end int) string {
	return (*((*string)(unsafe.Pointer(&p.buf))))[start:end]
}
