package loadguardian

import (
	"fmt"

	"github.com/GrGLeo/LoadBalancer/src/internal/config"
	servicemanager "github.com/GrGLeo/LoadBalancer/src/internal/container"
	"github.com/docker/docker/client"
)

type LoadGuardian struct {
  Client *client.Client
  Config config.Config
  RunningContainer map[string][]servicemanager.Container
}

func NewLoadGuardian(file string) (LoadGuardian, error) {
  cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
  if err != nil {
    return LoadGuardian{}, err
  }
  c, err := config.ParseYAML(file)
  if err != nil {
    return LoadGuardian{}, err
  }

  return LoadGuardian{
    Client: cli,
    Config: c,
  }, nil
}

func (lg *LoadGuardian) StopAll(timeout int) error {
  fmt.Println("Stopping all container")
  for name, containers := range lg.RunningContainer {
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
