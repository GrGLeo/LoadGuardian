package main

import (
	"fmt"
	"os"
)

func Up(file string) {
  lg, err := NewLoadGuardian(file)
  
  if err != nil {
    fmt.Println(err.Error())
    os.Exit(1)
  }
  lg.Config.CreateNetworks(lg.Client)
  lg.Config.PullServices(lg.Client)

  logChannel := make(chan LogMessage)
  go PrintLogs(logChannel)

  newServices, err := lg.Config.CreateAllService(lg.Client)
  lg.RunningContainer = newServices
  for _, service := range lg.RunningContainer {
    for _, container := range service {
      go func(c Container) {
        if err := container.StartAndFetchLogs(lg.Client, logChannel); err != nil {
          fmt.Println(err.Error())
        }
      }(container)
    }
  }
  select{}
}
