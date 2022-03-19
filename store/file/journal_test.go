package file

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJournalCheckGood(t *testing.T) {
	jf, err := os.Open("testdata/test-good.journal")
	require.Nil(t, err)

	j := newJournal(jf)

	err = j.Check()
	assert.Nil(t, err)
}

func TestJournalCheckBad(t *testing.T) {
	jf, err := os.Open("testdata/test-bad.journal")
	require.Nil(t, err)

	j := newJournal(jf)

	err = j.Check()
	assert.NotNil(t, err)
}
