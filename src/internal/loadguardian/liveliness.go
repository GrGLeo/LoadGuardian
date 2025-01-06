package loadguardian

import (
	"fmt"
	"sync"
)

func (lg *LoadGuardian) StatCheck() {
  wg := sync.WaitGroup{}
  for _, services := range lg.RunningServices {
    for _, cont := range services {
      wg.Add(1)
      go func() {
        defer wg.Done()
        cont.Info(lg.Client, lg.Logger)
        fmt.Printf("%v\n", cont)
      }()
    }
  }
  wg.Wait()
}
