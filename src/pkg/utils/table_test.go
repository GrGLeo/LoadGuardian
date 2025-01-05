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

func TestGenerateTable(t *testing.T) {
  var rows = [][]string{}
  rows = append(rows, []string{"Service", "Container", "Health", "CPU", "Memory"})
  rows = append(rows, []string{"LogFile", "healthchecks-yZqMw", "unhealthy", "0.00 %", "0.06 MB"})
  rows = append(rows, []string{"LogTen", "log_ten-WJwjC", "healthy", "0.00 %", "0.02 MB"})
  rows = append(rows, []string{"", "log_ten-mnvaZ", "healthy", "0.00 %", "0.02 MB"})
  rows = append(rows, []string{"", "log_ten-LvQZr", "healthy", "0.00 %", "0.02 MB"})
  baseLength := []int{7, 18, 9, 6, 7}
  table := utils.GenerateTable(rows, baseLength)
  expected := "----------------------------------------------------------\n|Service |Container          |Health    |CPU    |Memory  |\n----------------------------------------------------------\n|LogFile |healthchecks-yZqMw |unhealthy |0.00 % |0.06 MB |\n----------------------------------------------------------\n|LogTen  |log_ten-WJwjC      |healthy   |0.00 % |0.02 MB |\n----------------------------------------------------------\n|        |log_ten-mnvaZ      |healthy   |0.00 % |0.02 MB |\n----------------------------------------------------------\n|        |log_ten-LvQZr      |healthy   |0.00 % |0.02 MB |\n----------------------------------------------------------\n"
  if table != expected {
    t.Errorf("\nexpected:\n %s\n got\n %s", expected, table)
  }
}
