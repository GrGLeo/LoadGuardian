package loadguardian

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/GrGLeo/LoadBalancer/src/internal/config"
	servicemanager "github.com/GrGLeo/LoadBalancer/src/internal/servicemanager"
	"github.com/GrGLeo/LoadBalancer/src/pkg/logger"
)

var logChannel = make(chan servicemanager.LogMessage)

func StartProcress(file string) LoadGuardian {
  lg := GetLoadGuardian()
  c, err := config.ParseYAML(file)
  if err != nil {
    fmt.Println(err.Error())
    os.Exit(1)
  }
  lg.Config = c

  lg.Config.CreateNetworks(lg.Client)
  config.PullServices(&lg.Config, true, lg.Client)

  go logger.PrintLogs(logChannel)

  newServices, err := config.CreateAllService(&lg.Config, true, lg.Client)
  lg.RunningServices = newServices
  for _, service := range lg.RunningServices {
    for _, container := range service {
      go func(c servicemanager.Container) {
        if err := container.StartAndFetchLogs(lg.Client, logChannel); err != nil {
          fmt.Println(err.Error())
        }
      }(container)
    }
  }
  return *lg
}

func UpdateProcess(file string) error {
  lg := GetLoadGuardian()
  newConfig, err := config.ParseYAML(file)
  if err != nil {
    return errors.New("Invalid file") 
  }
  cd, err := lg.Config.CompareConfig(newConfig)

  // Pull new services
  err = config.PullServices(&cd, true, lg.Client)
  if err != nil {
    log.Fatal(err.Error())
  }

  // Stop removed service
  for name := range cd.RemovedService {
    containers, ok := lg.RunningServices[name]
    if ok {
      for _, c := range containers {
        fmt.Printf("Removing service: %s\n", name)
        timeout := 0
        c.Stop(lg.Client, &timeout)
      }
    }
  }

  // Create the new services
  newServices, err := config.CreateAllService(&cd, true, lg.Client)
  if err != nil {
    log.Fatal(err.Error())
  }
  // Start the new services
  for name := range cd.AddedService {
    containers, ok := newServices[name]
    if !ok {
      fmt.Println("Failed to match new Services with created one")
      continue
    }
    for _, container := range containers {
      go func(c servicemanager.Container) {
        if err := container.StartAndFetchLogs(lg.Client, logChannel); err != nil {
          fmt.Println(err.Error())
        }
      }(container)
      lg.RunningServices[name] = append(lg.RunningServices[name], container)
    }
  }

  // Rolling update
  err = config.PullServices(&cd, false, lg.Client)
  if err != nil {
    fmt.Println("Failed to pull updated services\n Keeping old version running")
    return nil
  }

  fmt.Println("Updating services")
  for name, service := range cd.UpdatedService {
    matchingRunningService, ok := lg.RunningServices[name]
    if !ok {
      fmt.Println("Failed to match updated Services with past one")
      continue
    }
    pastServiceCount := len(matchingRunningService)

    // We get the len and iterate over the old container
    for i := 0; i < pastServiceCount; i++ {
      // We need to add one more than the current number since container naming start at 1
      n := pastServiceCount + i + 1
      
      container, err := service.Create(lg.Client, n)
      if err != nil {
        fmt.Println("Failed to create container")
        continue
      }
      err = container.Start(lg.Client)
      if err != nil {
        fmt.Println("Failed to start container")
        continue
      }
      // Implement health inspection
      pastContainer := matchingRunningService[i]
      healthy, err := pastContainer.HealthCheck(lg.Client)
      if err != nil {
        fmt.Println("Failed to instpect container")
      }

      if healthy {
        timeout := 0
        pastContainer.Stop(lg.Client, &timeout)
        pastContainer.Remove(lg.Client)
        matchingRunningService[i] = container
      } else {
        // Abort the update, recursively??
      }
    }
  }
  return nil
}
