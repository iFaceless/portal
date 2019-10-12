package portal

import "github.com/pkg/errors"

var errStackIsEmpty = errors.New("stack is empty")

type stack struct {
	elements []interface{}
}

func newStack() *stack {
	return &stack{}
}

func (stack *stack) size() int {
	return len(stack.elements)
}

func (stack *stack) push(x interface{}) {
	if stack.elements == nil {
		stack.elements = make([]interface{}, 0)
	}

	stack.elements = append(stack.elements, x)
}

func (stack *stack) top() (interface{}, error) {
	if stack.size() == 0 {
		return nil, errStackIsEmpty
	}

	return stack.elements[stack.size()-1], nil
}

func (stack *stack) pop() (interface{}, error) {
	if stack.size() == 0 {
		return nil, errStackIsEmpty
	}

	x := stack.elements[stack.size()-1]
	stack.elements = stack.elements[0 : stack.size()-1]

	return x, nil
}
