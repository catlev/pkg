package stream

import (
	"bytes"
	"io"
)

// DOM provides a simple generic interface over streams.
type DOM struct {
	Name     string
	Fields   map[string]string
	Children []DOM
}

// Unmarshal loads from an encoded stream.
func (d *DOM) Unmarshal(src io.Reader) error {
	r := NewReader(src)
	d.parse(r)
	r.ExpectEOF()
	return r.Err()
}

func (d *DOM) Marshal() (io.Reader, error) {
	var buf bytes.Buffer
	w := NewWriter(&buf)
	d.serialize(w)
	return bytes.NewReader(buf.Bytes()), w.Err()
}

func (d *DOM) parse(r *Reader) {
	for r.Next() {
		switch r.Kind() {
		case Field:
			if d.Fields == nil {
				d.Fields = map[string]string{}
			}
			d.Fields[r.Name()] = r.StringField()
		case Record:
			c := DOM{Name: r.Name()}
			c.parse(r.Record())
			d.Children = append(d.Children, c)
		}
	}
}

func (d *DOM) serialize(w *Writer) {
	for n, v := range d.Fields {
		w.StringField(n, v)
	}
	for _, c := range d.Children {
		w.Record(c.Name, c.serialize)
	}
}
