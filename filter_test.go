package portal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseFilterString(t *testing.T) {

}

func Test_checkBracketPair(t *testing.T) {
	asserter := assert.New(t)
	asserter.Nil(checkBracketPairs([]byte("speaker")))
	asserter.Nil(checkBracketPairs([]byte("speaker[]")))
	asserter.Nil(checkBracketPairs([]byte("speaker[name,age[user[id]]]")))
	asserter.Equal(ErrUnmatchedBrackets, checkBracketPairs([]byte("speaker[")))
	asserter.Equal(ErrUnmatchedBrackets, checkBracketPairs([]byte("speaker]")))
	asserter.Equal(ErrUnmatchedBrackets, checkBracketPairs([]byte("speaker[user[id]]]")))
}
