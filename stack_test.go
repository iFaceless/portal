package portal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStack(t *testing.T) {
	asserter := assert.New(t)

	stack := NewStack()
	asserter.Equal(0, stack.Size())

	stack.Push(1)
	stack.Push(2)
	stack.Push(3)
	asserter.Equal(3, stack.Size())

	x, err := stack.Top()
	asserter.Equal(3, x)
	asserter.Nil(err)

	x, err = stack.Pop()
	asserter.Equal(3, x)
	asserter.Nil(err)

	x, err = stack.Pop()
	asserter.Equal(2, x)
	asserter.Nil(err)

	x, err = stack.Pop()
	asserter.Equal(1, x)
	asserter.Nil(err)

	_, err = stack.Pop()
	asserter.Equal(ErrStackIsEmpty, err)

	_, err = stack.Top()
	asserter.Equal(ErrStackIsEmpty, err)
}
