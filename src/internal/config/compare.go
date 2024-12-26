package config

import (
	servicemanager "github.com/GrGLeo/LoadBalancer/src/internal/container"
)

// TODO: handle the network changes
type ConfigDiff struct {
  AddedService  map[string]servicemanager.Service
  RemovedService map[string]servicemanager.Service 
  UpdatedService map[string]servicemanager.Service
  AddedNetworks []Network
  RemovedNetworks []Network
  UpdatedNetworks []Network
}

func (c *Config) CompareConfig(newConfig Config) (ConfigDiff, error) {
  compConfig := ConfigDiff{
		AddedService:   make(map[string]servicemanager.Service),
		RemovedService: make(map[string]servicemanager.Service),
		UpdatedService: make(map[string]servicemanager.Service),
	}
  for name, service := range newConfig.Service {
    oldService, ok := c.Service[name]
    if !ok {
      compConfig.AddedService[name] = service
    } else {
      if service.Compare(&oldService) {
        compConfig.UpdatedService[name] = service
      }
    }
  }

  for name, oldService := range c.Service {
    _, ok := newConfig.Service[name]
    if !ok {
      compConfig.RemovedService[name] = oldService
    }
  }
  return compConfig, nil
}


func (cf *ConfigDiff) GetService() map[string]servicemanager.Service {
  services := make(map[string]servicemanager.Service)
  for name, service := range cf.AddedService {
    services[name] = service
  }
  for name, service := range cf.UpdatedService {
    services[name] = service
  }
  return services
}

   

