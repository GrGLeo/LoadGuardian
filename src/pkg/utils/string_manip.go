package utils

import "sort"

func CompareStrings(sorted bool, a, b []string) bool {
  if sorted {
    sort.Strings(a)
    sort.Strings(b)
  }
  if len(a) != len(b) {
    return true
  }
  for i := range a {
    if a[i] != b[i] {
      return true
    }
  }
  return false
}
