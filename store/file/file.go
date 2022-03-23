package file

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"sync"
)

var ErrWriteAfterEnd = errors.New("writing after end of file")

// A File used as a persistent data store. This attempts to guard against some of the pitfalls in
// reliably writing to a disk.
type File struct {
	file   underlyingFile
	tx     sync.Mutex
	commit sync.RWMutex
}

// Tx represents a transaction of changes applied to a file. A given file may only have one
// transaction open at any given time.
type Tx struct {
	file    *File
	oldSize int64
	newSize int64
	journal *journal
}

type underlyingFile interface {
	io.ReaderAt
	io.WriterAt
	io.Closer

	Name() string
	Stat() (fs.FileInfo, error)
	Truncate(n int64) error
}

// Open a file for reading and possibly writing.
func Open(path string) (*File, error) {
	f, err := os.OpenFile(path, os.O_RDWR, 0755)
	if err != nil {
		return nil, err
	}

	if err := recoverFile(f); err != nil {
		return nil, fmt.Errorf("recovery failed: %w", err)
	}

	return &File{file: f}, nil
}

func recoverFile(dfile *os.File) error {
	jfile, err := os.Open(dfile.Name() + ".journal")
	if errors.Is(err, os.ErrNotExist) {
		// No journal file means that a write operation has not been interrupted, so no recovery is
		// required.
		return nil
	}
	if err != nil {
		return err
	}
	defer jfile.Close()
	defer os.Remove(jfile.Name())

	j := journal{file: jfile}

	if err := j.Check(); err != nil {
		// The journal is not valid. That means that writing to the file has not begun. So to
		// recover the file, all we need to do is delete the journal.
		return nil
	}

	// If we got here, the transaction needs to be rolled back.
	return j.Recover(dfile)
}

// ReadAt reads from the file at the given location.
func (f *File) ReadAt(buf []byte, pos int64) (int, error) {
	f.commit.RLock()
	defer f.commit.RUnlock()

	return f.file.ReadAt(buf, pos)
}

// WriteAt opens a transaction, writes the data, and then commits the transaction.
func (f *File) WriteAt(buf []byte, pos int64) (int, error) {
	tx, err := f.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Close()

	n, err := tx.WriteAt(buf, pos)
	if err != nil {
		return n, err
	}

	err = tx.Commit()
	return n, err
}

// Size returns the current size of the file.
func (f *File) Stat() (fs.FileInfo, error) {
	return f.file.Stat()
}

func (f *File) Close() error {
	return f.file.Close()
}

// Begin writing to the file.
func (f *File) Begin() (*Tx, error) {
	f.tx.Lock()

	var stat fs.FileInfo
	var j *journal
	var tx *Tx

	jf, err := os.OpenFile(f.file.Name()+".journal", os.O_RDWR|os.O_CREATE|os.O_EXCL, 0755)
	if err != nil {
		goto failure
	}

	stat, err = f.file.Stat()
	if err != nil {
		goto failure
	}

	j = newJournal(jf)

	err = j.init(stat.Size())
	if err != nil {
		goto failure
	}

	tx = &Tx{
		file:    f,
		journal: j,
		oldSize: stat.Size(),
		newSize: stat.Size(),
	}

	return tx, nil

failure:
	if jf != nil {
		jf.Close()
	}
	f.tx.Unlock()
	return nil, err
}

// WriteAt stages a change to the file.
func (f *Tx) WriteAt(buf []byte, pos int64) (int, error) {
	if pos > f.newSize {
		return 0, ErrWriteAfterEnd
	}

	// Split the operations into updates and appends to simplify the logic that interprets the
	// journal files.

	n := int64(len(buf))
	var appendBytes, updateBytes []byte
	var split int64

	if pos > f.oldSize {
		appendBytes = buf
	} else if pos+n > f.oldSize {
		split = f.oldSize - pos
		updateBytes = buf[:split]
		appendBytes = buf[split:]
	} else {
		updateBytes = buf
		split = n
	}

	if err := f.stageUpdate(updateBytes, pos); err != nil {
		return 0, err
	}
	if err := f.stageAppend(appendBytes, pos+split); err != nil {
		return 0, err
	}

	return len(buf), nil
}

// Truncate statges a change to the size of the file.
func (f *Tx) Truncate(size int64) error {
	if size > f.newSize {
		return ErrWriteAfterEnd
	}

	f.newSize = size

	if size >= f.oldSize {
		return nil
	}

	return f.stageTruncate(size)
}

// Commit writes all of the staged changes into the file in such a way that if the process is
// interrupted, the file can be restored to a known-good state.
func (f *Tx) Commit() error {
	// Write the checksum to the journal file
	err := f.journal.finalize()
	if err != nil {
		return err
	}

	// Wait for the journal file to have finished writing to disk
	d, err := os.Open(path.Dir(f.file.file.Name()))
	if err != nil {
		return err
	}
	defer d.Close()
	err = d.Sync()
	if err != nil {
		return err
	}
	err = f.journal.file.Sync()
	if err != nil {
		return err
	}

	// At this point we can begin actually making changes to the file. If we are interrupted at any
	// point in this process, whatever changes have been made will be undone during recovery.
	err = f.journal.Apply(f.file.file)
	if err != nil {
		return err
	}

	// Remove the journal to signal that the commit has completed.
	err = os.Remove(f.journal.file.Name())
	if err != nil {
		return err
	}
	err = d.Sync()
	if err != nil {
		return err
	}

	return nil
}

func (f *Tx) Close() error {
	f.file.tx.Unlock()
	return f.journal.Close()
}

func (f *Tx) stageUpdate(buf []byte, pos int64) error {
	if len(buf) == 0 {
		return nil
	}

	from := make([]byte, len(buf))
	_, err := f.file.ReadAt(from, pos)
	if err != nil {
		return err
	}

	return f.journal.Stage(operation{
		at:   pos,
		from: from,
		to:   buf,
	})
}

func (f *Tx) stageAppend(buf []byte, pos int64) error {
	if len(buf) == 0 {
		return nil
	}

	return f.journal.Stage(operation{
		at: pos,
		to: buf,
	})
}

func (f *Tx) stageTruncate(pos int64) error {
	buf := make([]byte, f.oldSize-pos)
	if _, err := f.file.ReadAt(buf, pos); err != nil {
		return err
	}

	return f.journal.Stage(operation{
		at:   pos,
		from: buf,
	})
}
