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
