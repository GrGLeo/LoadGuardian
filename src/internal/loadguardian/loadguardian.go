package loadguardian

import (
	"errors"
	"fmt"
	"os"

	"github.com/GrGLeo/LoadBalancer/src/internal/config"
	servicemanager "github.com/GrGLeo/LoadBalancer/src/internal/servicemanager"
	"github.com/docker/docker/client"
	"go.uber.org/zap"
)

type LoadGuardian struct {
  Client *client.Client
  Config config.Config
  RunningServices map[string][]servicemanager.Container
  Logger *zap.SugaredLogger
}


func NewLoadGuardian() (*LoadGuardian, error) {
  logger, err := zap.NewProduction()
  if err != nil {
    panic(err)
  }
  defer logger.Sync()
  sugaredLogger := logger.Sugar()
  sugaredLogger.Infoln("Initializing LoadGuardian")
  cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
  if err != nil {
    return &LoadGuardian{}, err
  }

  return &LoadGuardian{
    Client: cli,
    Logger: sugaredLogger,
  }, nil
}


func (lg *LoadGuardian) StopAll(timeout int) error {
  lg.Logger.Infoln("Stopping all container")
  for name, containers := range lg.RunningServices {
    lg.Logger.Infof("Stopping services: %s\n", name)
    for _, c := range containers {
      err := c.Stop(lg.Client, &timeout)
      if err != nil {
        return err
      }
      err = c.Remove(lg.Client)
      if err != nil {
        return err
      }
    }
  }
  return nil
}


func (lg *LoadGuardian) StopService(serviceName string, timeout int) error {
  lg.Logger.Infof("Stopping services: %s\n", serviceName)
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
  lg.Logger.Infoln("Stopping all services")
  err := lg.StopAll(0)
  if err != nil {
    lg.Logger.Infoln("Error while stopping service")
    os.Exit(1)
  }
  lg.Logger.Infoln("Services stopped. Exiting.")
}
