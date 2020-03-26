package stream

import (
	"io"
	"strconv"
)

type Writer struct {
	dest  io.Writer
	err   error
	depth int

	Indent, LineEnd string
}

func NewWriter(dest io.Writer) *Writer {
	return &Writer{dest: dest}
}

func (w *Writer) Err() error {
	return w.err
}

func (w *Writer) Record(name string, emitter func(*Writer)) {
	w.writeLine(name, "{")
	w.depth++
	emitter(w)
	w.depth--
	w.writeLine("}")
}

func (w *Writer) StringField(name, value string) {
	w.writeLine(name, ":", strconv.Quote(value))
}

func (w *Writer) IntField(name string, value int) {
	w.StringField(name, strconv.Itoa(value))
}

func (w *Writer) BoolField(name string, value bool) {
	v := "false"
	if value {
		v = "true"
	}
	w.StringField(name, v)
}

func (w *Writer) writeLine(ss ...string) {
	w.writePrefix()
	for _, s := range ss {
		w.write(s)
	}
	w.writeEnd()
}

func (w *Writer) writePrefix() {
	for i := 0; i < w.depth; i++ {
		w.write(w.Indent)
	}
}

func (w *Writer) writeEnd() {
	w.write(w.LineEnd)
}

func (w *Writer) write(s string) {
	if w.err != nil {
		return
	}
	_, w.err = w.dest.Write([]byte(s))
}
