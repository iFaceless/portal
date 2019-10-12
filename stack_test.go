package portal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStack(t *testing.T) {
	asserter := assert.New(t)

	stack := newStack()
	asserter.Equal(0, stack.size())

	stack.push(1)
	stack.push(2)
	stack.push(3)
	asserter.Equal(3, stack.size())

	x, err := stack.top()
	asserter.Equal(3, x)
	asserter.Nil(err)

	x, err = stack.pop()
	asserter.Equal(3, x)
	asserter.Nil(err)

	x, err = stack.pop()
	asserter.Equal(2, x)
	asserter.Nil(err)

	x, err = stack.pop()
	asserter.Equal(1, x)
	asserter.Nil(err)

	_, err = stack.pop()
	asserter.Equal(errStackIsEmpty, err)

	_, err = stack.top()
	asserter.Equal(errStackIsEmpty, err)
}
