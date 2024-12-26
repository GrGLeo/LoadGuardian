package loadguardian

import (
	"errors"
	"fmt"
	"os"

	"github.com/GrGLeo/LoadBalancer/src/internal/config"
	servicemanager "github.com/GrGLeo/LoadBalancer/src/internal/container"
	"github.com/docker/docker/client"
)


type LoadGuardian struct {
  Client *client.Client
  Config config.Config
  RunningServices map[string][]servicemanager.Container
}


func NewLoadGuardian() (*LoadGuardian, error) {
  cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
  if err != nil {
    return &LoadGuardian{}, err
  }

  return &LoadGuardian{
    Client: cli,
  }, nil
}


func (lg *LoadGuardian) StopAll(timeout int) error {
  fmt.Println("Stopping all container")
  for name, containers := range lg.RunningServices {
    fmt.Printf("Stopping services: %s\n", name)
    for _, c := range containers {
      err := c.Stop(lg.Client, &timeout)
      if err != nil {
        return err
      }
    }
  }
  return nil
}


func (lg *LoadGuardian) StopService(serviceName string, timeout int) error {
  fmt.Println("Stopping service:", serviceName) 
  containers, ok := lg.RunningServices[serviceName]
  if !ok {
    return errors.New(fmt.Sprintf("Failed to found service: %s", serviceName))
  }
  for _, c := range containers {
    err := c.Stop(lg.Client, &timeout)
    if err != nil {
      return err
    }
  }
  return nil
}


func (lg *LoadGuardian) CleanUp() {
  fmt.Println("Stopping all services...")
  err := lg.StopAll(0)
  if err != nil {
    fmt.Println("Error while stopping service")
    os.Exit(1)
  }
  fmt.Println("Services stopped. Exiting.")
}
