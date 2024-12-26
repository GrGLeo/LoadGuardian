package config

import servicemanager "github.com/GrGLeo/LoadBalancer/src/internal/container"


type ComparedConfig struct {
  AddedService []servicemanager.Service
  RemovedService []servicemanager.Service
  NewService []servicemanager.Service
  AddedNetworks []Network
  RemovedNetworks []Network
  NewNetworks []Network
}

func (c *Config) CompareConfig(newConfig Config) (ComparedConfig, error) {
  return nil
}
