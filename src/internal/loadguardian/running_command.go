package loadguardian

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/GrGLeo/LoadBalancer/src/internal/config"
	servicemanager "github.com/GrGLeo/LoadBalancer/src/internal/container"
	"github.com/GrGLeo/LoadBalancer/src/pkg/logger"
)


func StartProcress(file string) LoadGuardian {
  lg := GetLoadGuardian()
  c, err := config.ParseYAML(file)
  if err != nil {
    fmt.Println(err.Error())
    os.Exit(1)
  }
  lg.Config = c

  lg.Config.CreateNetworks(lg.Client)
  config.PullServices(&lg.Config, lg.Client)

  logChannel := make(chan servicemanager.LogMessage)
  go logger.PrintLogs(logChannel)

  newServices, err := config.CreateAllService(&lg.Config, lg.Client)
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

  // Handle new & updated services
  fmt.Println(lg.Client)
  err = config.PullServices(&cd, lg.Client)
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

  _, err = config.CreateAllService(&cd, lg.Client)
  if err != nil {
    log.Fatal(err.Error())
  }
  return nil
}
