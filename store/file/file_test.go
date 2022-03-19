package file

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileCancel(t *testing.T) {
	f := aFileNamed(t, "testdata/eg")
	fileHasContents(t, f, []byte{})
	writeWithoutCommit(t, f, []byte{1, 2, 3})
	f = aFileNamed(t, "testdata/eg")
	fileHasContents(t, f, []byte{})
	require.Nil(t, f.Close())
}

func TestFileBrokenJournal(t *testing.T) {
	f := aFileNamed(t, "testdata/eg")
	fileHasContents(t, f, []byte{})
	writeWithBrokenJournal(t, f, []byte{1, 2, 3})
	f = aFileNamed(t, "testdata/eg")
	fileHasContents(t, f, []byte{})
	require.Nil(t, f.Close())
}

func TestFileAppend(t *testing.T) {
	f := aFileNamed(t, "testdata/eg2")
	fileHasContents(t, f, []byte{})
	writeAndCommit(t, f, []byte{1, 2, 3})
	f = aFileNamed(t, "testdata/eg2")
	fileHasContents(t, f, []byte{1, 2, 3})
	clearFile(t, f)
	f = aFileNamed(t, "testdata/eg2")
	fileHasContents(t, f, []byte{})
	require.Nil(t, f.Close())
}

func TestFileUpdate(t *testing.T) {
	f := aFileNamed(t, "testdata/eg3")
	fileHasContents(t, f, []byte{1, 2, 3})
	writeAndCommit(t, f, []byte{1, 3, 4})
	f = aFileNamed(t, "testdata/eg3")
	fileHasContents(t, f, []byte{1, 3, 4})
	writeAndCommit(t, f, []byte{1, 2, 3})
	f = aFileNamed(t, "testdata/eg3")
	fileHasContents(t, f, []byte{1, 2, 3})
	require.Nil(t, f.Close())
}

func aFileNamed(t *testing.T, name string) *File {
	f, err := Open(name)
	require.Nil(t, err)

	return f
}

func fileHasContents(t *testing.T, f *File, expect []byte) {
	t.Helper()

	n, err := f.Size()
	require.Nil(t, err)
	assert.Equal(t, int64(len(expect)), n)

	if len(expect) == 0 {
		return
	}

	buf := make([]byte, 3)
	_, err = f.ReadAt(buf, 0)
	require.Nil(t, err)
	assert.Equal(t, expect, buf)
}

func writeAndCommit(t *testing.T, f *File, buf []byte) {
	t.Helper()

	tx, err := f.Begin()
	require.Nil(t, err)

	_, err = tx.WriteAt(buf, 0)
	require.Nil(t, err)

	err = tx.Commit()
	require.Nil(t, err)

	err = tx.Close()
	require.Nil(t, err)

	err = f.Close()
	require.Nil(t, err)
}

func clearFile(t *testing.T, f *File) {
	t.Helper()

	tx, err := f.Begin()
	require.Nil(t, err)

	err = tx.Truncate(0)
	require.Nil(t, err)

	err = tx.Commit()
	require.Nil(t, err)

	err = tx.Close()
	require.Nil(t, err)

	err = f.Close()
	require.Nil(t, err)
}

func writeWithoutCommit(t *testing.T, f *File, buf []byte) {
	t.Helper()

	tx, err := f.Begin()
	require.Nil(t, err)

	_, err = tx.WriteAt(buf, 0)
	require.Nil(t, err)

	err = tx.Close()
	require.Nil(t, err)

	err = f.Close()
	require.Nil(t, err)
}

func writeWithBrokenJournal(t *testing.T, f *File, b []byte) {
	t.Helper()

	tx, err := f.Begin()
	require.Nil(t, err)

	_, err = tx.WriteAt([]byte{1, 2, 3}, 0)
	require.Nil(t, err)

	err = tx.journal.file.Close()
	require.Nil(t, err)

	err = f.Close()
	require.Nil(t, err)
}
