package file

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"hash"
	"io"
	"os"

	"golang.org/x/crypto/sha3"
)

var ErrHashCheck = errors.New("hash check failed")

var hashSize = sha3.New256().Size()

type journal struct {
	file   *os.File
	hash   hash.Hash
	writer io.Writer
}

type operation struct {
	at       int64
	from, to []byte
}

func newJournal(file *os.File) *journal {
	hash := sha3.New256()
	writer := io.MultiWriter(file, hash)

	return &journal{
		file:   file,
		hash:   hash,
		writer: writer,
	}
}

func (j *journal) Close() error {
	if err := j.finalize(); err != nil {
		return nil
	}
	return j.file.Close()
}

func (j *journal) init(fsize int64) error {
	padding := make([]byte, hashSize)
	if _, err := j.file.Write(padding); err != nil {
		return err
	}
	return writeInt(j.writer, fsize)
}

func (j *journal) finalize() error {
	if err := j.writeHeader(); err != nil {
		return err
	}
	return nil
}

func (j *journal) Stage(op operation) error {
	if err := writeInt(j.writer, op.at); err != nil {
		return err
	}
	if err := writeBytes(j.writer, op.from); err != nil {
		return err
	}
	if err := writeBytes(j.writer, op.to); err != nil {
		return err
	}
	return nil
}

func (j *journal) Check() error {
	err := j.seekAfterHeader()
	if err != nil {
		return err
	}

	h := sha3.New256()

	_, err = io.Copy(h, j.file)
	if err != nil {
		return err
	}

	_, err = j.file.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	hash := make([]byte, hashSize)
	_, err = io.ReadFull(j.file, hash)
	if err != nil {
		return err
	}

	if !bytes.Equal(hash, h.Sum(nil)) {
		return ErrHashCheck
	}

	return nil
}

func (j *journal) Apply(f underlyingFile) error {
	err := j.seekAfterHeader()
	if err != nil {
		return err
	}
	r := bufio.NewReader(j.file)
	if _, err := binary.ReadVarint(r); err != nil {
		return err
	}

	return exec(r, f, operation.Apply)
}

func (j *journal) Recover(f underlyingFile) error {
	err := j.seekAfterHeader()
	if err != nil {
		return err
	}
	r := bufio.NewReader(j.file)

	n, err := binary.ReadVarint(r)
	if err != nil {
		return err
	}

	err = restoreSize(f, n)
	if err != nil {
		return err
	}

	return exec(r, f, operation.Recover)
}

func (j *journal) writeHeader() error {
	header := j.hash.Sum(nil)
	_, err := j.file.WriteAt(header, 0)
	return err
}

func (j *journal) seekAfterHeader() error {
	_, err := j.file.Seek(int64(hashSize), io.SeekStart)
	return err
}

func exec(r *bufio.Reader, f underlyingFile, fn func(operation, underlyingFile) error) error {
	var op operation
	for {
		err := op.ReadFrom(r)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if err := fn(op, f); err != nil {
			return err
		}
	}

	return nil
}

func restoreSize(f underlyingFile, n int64) error {
	stat, err := f.Stat()
	if err != nil {
		return err
	}
	if n < stat.Size() {
		return f.Truncate(n)
	}
	_, err = f.WriteAt(make([]byte, n-stat.Size()), stat.Size())
	return err
}

func (o *operation) ReadFrom(r *bufio.Reader) error {
	at, err := binary.ReadVarint(r)
	if err != nil {
		return err
	}

	from, err := readBytes(r)
	if err != nil {
		return err
	}

	to, err := readBytes(r)
	if err != nil {
		return err
	}

	o.at = at
	o.from = from
	o.to = to

	return nil
}

func (o operation) Apply(f underlyingFile) error {
	if len(o.to) == 0 {
		return f.Truncate(o.at)
	}
	_, err := f.WriteAt(o.to, o.at)
	return err
}

func (o operation) Recover(f underlyingFile) error {
	if len(o.from) == 0 {
		return nil
	}
	_, err := f.WriteAt(o.from, o.at)
	return err
}

func writeInt(w io.Writer, x int64) error {
	var buf [binary.MaxVarintLen64]byte
	n := binary.PutVarint(buf[:], x)

	if _, err := w.Write(buf[:n]); err != nil {
		return err
	}
	return nil
}

func writeBytes(w io.Writer, buf []byte) error {
	if err := writeInt(w, int64(len(buf))); err != nil {
		return err
	}
	if _, err := w.Write(buf); err != nil {
		return err
	}
	return nil
}

func readBytes(r interface {
	io.ByteReader
	io.Reader
}) ([]byte, error) {

	n, err := binary.ReadVarint(r)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, n)
	_, err = io.ReadFull(r, buf)
	if err != nil {
		return nil, err
	}

	return buf, nil
}
