package loadguardian

import (
	"context"
	"sync"
	"time"

	"github.com/docker/docker/api/types/container"
)

func (lg *LoadGuardian) StatCheck() {
  wg := sync.WaitGroup{}
  maxRetries := 3
  delay := 5 * time.Second

  for _, services := range lg.RunningServices {
    for _, cont := range services {
      wg.Add(1)
      go func() {
        defer wg.Done()
        retries := 0
        for {
          cont.Info(lg.Client, lg.Logger)
          if cont.Health == "healthy" || cont.Health == "starting"  {
            break
          }
          // Autoscalling logic could be implemented here
          lg.Logger.Warnw("Container unhealty restarting...", "container", cont.Name)
          time.Sleep(delay)
          retries++
          err := lg.Client.ContainerRestart(context.Background(), cont.ID, container.StopOptions{})
          if err != nil {
            lg.Logger.Warnw("Failed to restart unhealthy container", "retries", retries, "container", cont.Name)
          }
          if retries >= maxRetries {
            lg.Logger.Errorw("Failed to restart unhealthy container max retries reached", "retries", retries, "container", cont.Name)
            break
          }
        }
      }()
    }
  }
  wg.Wait()
}
