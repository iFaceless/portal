package portal

import "github.com/pkg/errors"

var ErrStackIsEmpty = errors.New("stack is empty")

type Stack struct {
	elements []interface{}
}

func NewStack() *Stack {
	return &Stack{}
}

func (stack *Stack) Size() int {
	return len(stack.elements)
}

func (stack *Stack) Push(x interface{}) {
	if stack.elements == nil {
		stack.elements = make([]interface{}, 0)
	}

	stack.elements = append(stack.elements, x)
}

func (stack *Stack) Top() (interface{}, error) {
	if stack.Size() == 0 {
		return nil, ErrStackIsEmpty
	}

	return stack.elements[stack.Size()-1], nil
}

func (stack *Stack) Pop() (interface{}, error) {
	if stack.Size() == 0 {
		return nil, ErrStackIsEmpty
	}

	x := stack.elements[stack.Size()-1]
	stack.elements = stack.elements[0 : stack.Size()-1]

	return x, nil
}
