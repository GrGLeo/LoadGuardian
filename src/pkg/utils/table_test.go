package utils_test

import (
	"testing"

	"github.com/GrGLeo/LoadBalancer/src/pkg/utils"
)

func TestLengthPrep(t *testing.T) {
  length := 10
  str := "log_ten"
  expected := "log_ten   "
  utils.PadString(&str, length)
  if len(str) != len(expected) {
    t.Errorf("expected: %q, got: %q", expected, str)
  }
}

func TestGenerateRow(t *testing.T) {
  raw := []string{"a", "b", "c"}
  rawlen := []int{3, 3, 3}
  expected := "|a   |b   |c   |\n"

  t.Run("valid input", func(t *testing.T) {
    got, err := utils.GenerateRow(raw, rawlen)
    if err != nil {
      t.Errorf("Got error: %q", err.Error())
    }
    if got != expected {
      t.Errorf("expected: %q, got %q", expected, got)
    }
  })

  t.Run("error scenario", func(t *testing.T) {
    invalidRawlen := []int{3, 3} // Less elements than raw input
    got, err := utils.GenerateRow(raw, invalidRawlen)
    
    if err == nil {
      t.Errorf("Expected error but got nil")
    }
    if got != "" {
      t.Errorf("Expected empty result but got %q", got)
    }
  })
}

func TestFloatToValue(t *testing.T) {
  fl := float64(13.45)
  val := "MB"
  expected := "13.45 MB"
  got := utils.ConvertFloatToValue(fl, val)
  if got != expected {
    t.Errorf("expected %q, got %q", expected, got)
  }
}

func TestGetBaseLength(t *testing.T) {
  header := []string{"Service", "Container", "Health", "CPU", "Memory"}
  expected := []int{7, 9, 6, 3, 6}
  got := utils.GetBaseLength(header)
  if len(expected) != len(got) {
    t.Errorf("Both length should be equal got: %d, expected %d", len(got), len(expected))
  }
  for i := range got {
    if got[i] != expected[i] {
      t.Errorf("expected %d, got %d", expected[i], got[i])
    }
  }
}

func TestUpdateBaseLength(t *testing.T) {
  base := []int{7, 9, 6, 3, 6}
  row := []int{5, 7, 7, 5, 4}
  expected := []int{7, 9, 7, 5, 6}

  t.Run("valid input", func(t *testing.T) {
    utils.UpdateBaseLength(&base, &row)
    for i := range base {
      if base[i] != expected[i] {
        t.Errorf("expected %d, got %d", expected[i], base[i])
      }
    }
  })

  t.Run("invalid input", func(t *testing.T) {
    row := []int{5, 7, 7, 5}
    err := utils.UpdateBaseLength(&base, &row)
    if err == nil {
      t.Errorf("Expected error, got nil")
    }
  })
}
