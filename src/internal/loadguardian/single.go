package loadguardian

import (
	"fmt"
	"os"
	"sync"
)

var (
  once sync.Once
  lg *LoadGuardian
  err error
)

func GetLoadGuardian() *LoadGuardian {
  once.Do(func() {
    lg, err = NewLoadGuardian()
    if err != nil {
      fmt.Println(err.Error())
      os.Exit(1)
    }
  })
  return lg
}
