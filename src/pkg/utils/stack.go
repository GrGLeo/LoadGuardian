package utils

import "errors"

var (
  EmptyStackError = errors.New("Empty stack")
)

type Stack struct {
  items []interface{}
}

func (s *Stack) Push(item interface{}) {
  s.items = append(s.items, item)
}

func (s *Stack) Size() int {
  return len(s.items)
}

func (s *Stack) Peek() (interface{}, error) {
  n := len(s.items)
  if n > 0 {
    return s.items[n-1], nil
  } else {
    return nil, EmptyStackError
  }
}

func (s *Stack) Pop() (interface{}, error) {
  n := len(s.items)
  if n > 0 {
    last_item := s.items[n-1]
    s.items = s.items[:n-1]
    return last_item, nil
  } else {
    return nil, EmptyStackError
  }
}

func (s *Stack) IsEmpty() bool {
  return len(s.items) == 0
}
