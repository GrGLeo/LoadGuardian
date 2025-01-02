package utils

import "errors"

var (
  EmptyStackError = errors.New("Empty stack")
)

type Stack[T any] struct {
  items []T
}

func (s *Stack[T]) Push(item T) {
  s.items = append(s.items, item)
}

func (s *Stack[T]) Size() int {
  return len(s.items)
}

func (s *Stack[T]) Peek() (T, error) {
  n := len(s.items)
  if n > 0 {
    return s.items[n-1], nil
  } else {
    var zeroValue T
    return zeroValue, EmptyStackError
  }
}

func (s *Stack[T]) Pop() (T, error) {
  n := len(s.items)
  if n > 0 {
    last_item := s.items[n-1]
    s.items = s.items[:n-1]
    return last_item, nil
  } else {

    var zeroValue T
    return zeroValue, EmptyStackError
  }
}

func (s *Stack[T]) IsEmpty() bool {
  return len(s.items) == 0
}
