package utils_test

import (
	"strconv"
	"testing"

	"github.com/GrGLeo/LoadBalancer/src/pkg/utils"
)


func TestCompareStrings(t *testing.T) {
	tests := []struct {
    sorting bool
		a, b   []string
		result bool
	}{
    {true, []string{"8080"}, []string{"8080"}, false},
		{true, []string{"apple", "banana", "cherry"}, []string{"banana", "cherry", "apple"}, false},
		{true, []string{"apple", "banana", "cherry"}, []string{"apple", "banana", "date"}, true},
		{true, []string{"apple", "banana", "cherry"}, []string{"apple", "banana"}, true},
		{true, []string{"apple"}, []string{"apple"}, false},
		{true, []string{"apple", "banana"}, []string{"banana", "apple"}, false},
		{false, []string{"apple", "banana"}, []string{"banana", "apple"}, true},
	}

	// Running tests
	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got := utils.CompareStrings(test.sorting, test.a, test.b)
			if got != test.result {
				t.Errorf("Test %d failed: CompareStrings(%v, %v) = %v, want %v", i+1, test.a, test.b, got, test.result)
			}
		})
	}
}


func TestStack(t *testing.T) {
	// Initialize a new stack
	stack := &utils.Stack{}

	t.Run("Push and Size", func(t *testing.T) {
		stack.Push(10)
		stack.Push(20)
		stack.Push(30)
		if got := stack.Size(); got != 3 {
			t.Errorf("expected size 3, got %d", got)
		}
	})

	t.Run("Peek", func(t *testing.T) {
		item, err := stack.Peek()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if item != 30 {
			t.Errorf("expected top item 30, got %v", item)
		}
	})

	t.Run("Pop", func(t *testing.T) {
		item, err := stack.Pop()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if item != 30 {
			t.Errorf("expected popped item 30, got %v", item)
		}
		if got := stack.Size(); got != 2 {
			t.Errorf("expected size 2 after pop, got %d", got)
		}
	})

	t.Run("IsEmpty", func(t *testing.T) {
		if got := stack.IsEmpty(); got {
			t.Errorf("expected stack not to be empty")
		}
		stack.Pop()
		stack.Pop()
		if got := stack.IsEmpty(); !got {
			t.Errorf("expected stack to be empty")
		}
	})

	t.Run("EmptyStackError", func(t *testing.T) {
		_, err := stack.Pop()
		if err != utils.EmptyStackError {
			t.Errorf("expected EmptyStackError, got %v", err)
		}

		_, err = stack.Peek()
		if err != utils.EmptyStackError {
			t.Errorf("expected EmptyStackError, got %v", err)
		}
	})
}

