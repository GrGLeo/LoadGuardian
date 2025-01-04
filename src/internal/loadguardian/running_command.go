package loadguardian

import (
	"errors"
	"fmt"

	"github.com/GrGLeo/LoadBalancer/src/internal/config"
	servicemanager "github.com/GrGLeo/LoadBalancer/src/internal/servicemanager"
	"github.com/GrGLeo/LoadBalancer/src/pkg/logger"
	"github.com/GrGLeo/LoadBalancer/src/pkg/utils"
	"go.uber.org/zap"
)

var zaplog = zap.L().Sugar()
var logChannel = make(chan servicemanager.LogMessage)

func StartProcress(file string){
  lg := GetLoadGuardian()
  c, err := config.ParseYAML(file)
  if err != nil {
    fmt.Println(err)
    zaplog.Fatalln(err.Error())
  }
  lg.Config = c

  lg.Config.CreateNetworks(lg.Client)
  config.PullServices(&lg.Config, true, lg.Client)

  go logger.PrintLogs(logChannel)

  newServices, err := config.CreateAllService(&lg.Config, true, lg.Client)
  if err != nil {
    zaplog.Errorf("Error while creating service: %q", err.Error())
  }
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
}

func UpdateProcess(file string) error {
  lg := GetLoadGuardian()
  lg.Logger.Infoln("Parsing configuration")
  newConfig, err := config.ParseYAML(file)
  if err != nil {
  lg.Logger.Errorln("Error while parsing new configuration")
    return errors.New("Invalid file") 
  }
  cd, err := lg.Config.CompareConfig(newConfig)

  // Pull new services
  err = config.PullServices(&cd, true, lg.Client)
  if err != nil {
    lg.Logger.Fatalln("Failed pull services")
  }

  // Stop removed service
  for name := range cd.RemovedService {
    containers, ok := lg.RunningServices[name]
    if ok {
      for _, c := range containers {
        lg.Logger.Infof("Removing service: %s\n", name)
        timeout := 0
        c.Stop(lg.Client, &timeout)
      }
    }
  }

  // Create the new services
  newServices, err := config.CreateAllService(&cd, true, lg.Client)
  if err != nil {
    lg.Logger.Fatalln("Failed to create services")
  }
  // Start the new services
  for name := range cd.AddedService {
    containers, ok := newServices[name]
    if !ok {
      lg.Logger.Warnln("Failed to match new Services with created one")
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
    lg.Logger.Errorln("Failed to pull updated services. Keeping old version running")
    return nil
  }

  lg.Logger.Infoln("Updating services")
  var rollbackPairs = utils.Stack[utils.Stack[servicemanager.ContainerRollbackConfig]]{}
  for name, service := range cd.UpdatedService {
    currentIteration :=  utils.Stack[servicemanager.ContainerRollbackConfig]{}
    matchingRunningService, ok := lg.RunningServices[name]
    if !ok {
      lg.Logger.Warnln("Failed to match updated Services with past one")
      continue
    }
    pastServiceCount := len(matchingRunningService)

    // We get the len and iterate over the old container
    for i := 0; i < pastServiceCount; i++ {
      container, err := service.Create(lg.Client)
      if err != nil {
        lg.Logger.Errorln("Failed to create container")
        continue
      }
      go func(c servicemanager.Container) {
        if err := container.StartAndFetchLogs(lg.Client, logChannel); err != nil {
          fmt.Println(err.Error())
        }
      }(container)

      // Implement health inspection
      pastContainer := matchingRunningService[i]
      healthy := servicemanager.CheckContainerHealth(lg.Client, container, lg.Logger)

      if healthy {
        timeout := 0
        pastContainer.Stop(lg.Client, &timeout)
        pastContainer.Remove(lg.Client)
        matchingRunningService[i] = container
        // Store the pair in case or rollback
        currentIteration.Push(servicemanager.ContainerRollbackConfig{
          PastService: service,
          New: container,
        })
      } else {
        // If not healthy, we revert back the full update.
        // idealy the all config should be revert cause new or removed
        // container might be needed by the old image
        // Stop and remove the current iteration
        timeout := 0
        container.Stop(lg.Client, &timeout)
        container.Remove(lg.Client)
        for !rollbackPairs.IsEmpty() {
          lg.Logger.Infoln("Rolling Back...")
          // Rollback services by services in reverse order
          pastIteration, _ := rollbackPairs.Pop()
          for !pastIteration.IsEmpty() {
            // Rollback container by container in reverse order
            pastIterationContainer, _ := pastIteration.Pop()
            lg.Logger.Infoln("Rolling back Container: ", pastIterationContainer.PastService.Image)
            cont, _ := pastIterationContainer.PastService.Create(lg.Client)
            cont.Start(lg.Client)
            pastIterationContainer.New.Stop(lg.Client, &timeout)
            pastIterationContainer.New.Remove(lg.Client)
            serviceName := pastIterationContainer.PastService.Image
            lg.RunningServices[serviceName] = append(lg.RunningServices[serviceName], cont)
          }
        }
        return nil
      }
    }
  rollbackPairs.Push(currentIteration)
  }
  return nil
}
