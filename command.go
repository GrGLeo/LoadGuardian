package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
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
  // Handling keyboard shutdown
  signalChannel := make(chan os.Signal, 1)
  signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
  <-signalChannel

  // Clean up
  fmt.Println("Stopping all services...")
  err = lg.StopAll(0)
  if err != nil {
    fmt.Println("Error while stopping service")
    os.Exit(1)
  }
  fmt.Println("Servies stopped. Exiting.")
}

