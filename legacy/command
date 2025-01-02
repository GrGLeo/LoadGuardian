package main

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

func Up(file string) {
  // Setting up start up
  pid := os.Getpid()
  err := os.WriteFile("loadguardian.pid", []byte(fmt.Sprintf("%d", pid)), 0644)
  if err != nil {
    fmt.Println("Failed to write pid, gracefull down command will not be available")
  }
  defer os.Remove("loadguardian.pid")

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

func Down() error {
  pidData, err := os.ReadFile("loadguardian.pid")
  if err != nil {
    return errors.New("Failed to find active guardian")
  }
  pid, err := strconv.Atoi(string(pidData))
  if err != nil {
    return errors.New("Invalid pid in file")
  }
  process, err := os.FindProcess(pid)
  err = process.Signal(syscall.SIGTERM)
  if err != nil {
    return errors.New("Failed to end process")
  }
  return nil
}
