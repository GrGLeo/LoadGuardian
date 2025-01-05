package config

import (
	"context"
	"errors"
	"strings"
	"sync"

	servicemanager "github.com/GrGLeo/LoadBalancer/src/internal/servicemanager"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"go.uber.org/zap"
)

const greenCheck = "\033[32m✓\033[0m"

type Network struct {
  Driver string `yaml:"driver,omitempty"`
}

type Volume struct {
}

type Config struct {
  Service map[string]servicemanager.Service `yaml:"service"`
  Network map[string]Network `yaml:"networks,omitempty"`
  Volume map[string]Volume `yaml:"volume,omitempty"`
}

type ServiceProvider interface {
  GetService(bool) map[string]servicemanager.Service
}

func (c *Config) GetService(p bool) map[string]servicemanager.Service {
  //p: useless for this but needed for ConfigDiff implementation
  return c.Service
}

// CreateNetworks ensures the necessary Docker networks are created by checking their existence and creating them if they do not already exist.
func (c *Config) CreateNetworks(cli *client.Client, logger *zap.SugaredLogger) error {
  networks, err := cli.NetworkList(context.Background(), network.ListOptions{})
  if err != nil {
    logger.Errorw("Failed to retrieve network list", "error", err.Error())
    return err
  }
  for name, value := range c.Network {
    networkExist := false
    for _,net := range networks {
      if net.Name == name {
        networkExist = true
        break
      }
    }
    if networkExist {
      logger.Infow("Network already exists", "network", name)
    } else {
      opt := network.CreateOptions{
        Driver: value.Driver,
      }
      _, err := cli.NetworkCreate(context.Background(), name, opt)
      logger.Infow("Network created", "network", name)
      if err != nil {
        logger.Errorw("Failed to create network", "error", err.Error())
        return err
      }
    }
  }
  return nil
}

// CheckServices identifies which services need their Docker images pulled by comparing existing images with the required service images.
func CheckServices(sp ServiceProvider, cli *client.Client) (map[string]bool, error) {
  // Retrieve existing image
  resp, err := cli.ImageList(context.Background(), image.ListOptions{})
  if err != nil {
    return nil, err
  }
  // Checking if service already have an existing image
  pullingServices := make(map[string]bool)
  Services := sp.GetService(true)
  for _, serv := range Services {
    pullingServices[serv.Image] = true
  }
  for _, image := range resp {
    imageTag := image.RepoTags
    if len(imageTag) > 0 {
      // splitting tag from name
      name := strings.Split(imageTag[0], ":")[0]
      if _, ok := pullingServices[name]; ok {
        pullingServices[name] = false
      }
    }
  }
  return pullingServices, nil
}


// PullServices pulls Docker images for the given services if they are not already available locally.
func PullServices(sp ServiceProvider, p bool, cli *client.Client, logger *zap.SugaredLogger) error {
  ImageToPull, err:= CheckServices(sp, cli)
  if err != nil {
    logger.Warnln("Failed to inspect images. Pulling image for all services")
  }
  Services := sp.GetService(p)
  wg := sync.WaitGroup{}
  for name, service := range Services {
    // Checking if image need to be pulled
    wg.Add(1)
    go func() error {
      defer wg.Done()
      value, ok := ImageToPull[service.Image]
      if !value && ok {
        logger.Infow("Service already pulled", "service", name)
        return nil
      }
      // Pulling Image
      logger.Infow("Pulling image", "service", name)
      _, err := cli.ImagePull(context.Background(), service.Image, image.PullOptions{})
      if err != nil {
        logger.Errorw("Error while pulling image", "service", name, "error", err.Error())
        return err
      }
      logger.Infow("Successfully pulled image", "service", name)
      return nil
    }()
  }
  wg.Wait()
  return nil
}

// CreateAllService creates and starts the specified number of replicas for each service using the provided Docker client.
func CreateAllService(sp ServiceProvider, p bool, cli *client.Client, logger *zap.SugaredLogger) (map[string][]servicemanager.Container, error) {
  runningCont := make(map[string][]servicemanager.Container)
  Services := sp.GetService(p)

  for name, service := range Services {
    for i := 1; i <= service.Replicas; i++ {
      logger.Infow("creating container", "service", name, "replica", i)
      container, err := service.Create(cli)
      if err != nil {
        logger.Errorw("Failed to create container", "service", name, "replica", i)
        return runningCont, err
      }
      runningCont[name] = append(runningCont[name], container)
    }
  }
  return runningCont, nil
}


// OrderService sorts the services based on their dependencies, detecting cyclic dependencies and ensuring a correct initialization order.
func OrderService(services map[string]servicemanager.Service) (map[string]servicemanager.Service, error) {
  visited := make(map[string]bool)
  stack := []string{}
  temp := make(map[string]bool)

  var visit func(string) error
  visit = func(service string) error {
    if temp[service] {
      return errors.New("cyclic dependency detected on service")
    }
    if !visited[service] {
      temp[service] = true
      for _, dep := range services[service].Dependencies {
        if err := visit(dep); err != nil {
          return err 
        }
      }
      temp[service] = false
      visited[service] = true
      stack = append(stack, service)
    }
    return nil
  }
  for name := range services {
    if err := visit(name); err != nil {
      return nil, err
    }
  }
 
  // Order services based on stack
  newOrder := make(map[string]servicemanager.Service)
  for _, name := range stack {
    newOrder[name] = services[name]
  }
  return newOrder, nil
}
