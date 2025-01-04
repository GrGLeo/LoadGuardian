package utils

import (
	"math/rand"
	"sort"
)


const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

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


func GenerateName(length int) string {
  result := make([]byte, length)
  for i := 0; i < length; i++ {
    result[i] = charset[rand.Intn(len(charset))]
  }
  return string(result)
}
