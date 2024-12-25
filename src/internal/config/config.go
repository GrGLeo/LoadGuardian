package config

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	servicemanager "github.com/GrGLeo/LoadBalancer/src/internal/container"
	"github.com/GrGLeo/LoadBalancer/src/pkg/logger"
	"github.com/briandowns/spinner"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"gopkg.in/yaml.v3"
)

const greenCheck = "\033[32m✓\033[0m"

type Network struct {
  Driver string `yaml:"driver,omitempty"`
}

type Config struct {
  Service map[string]servicemanager.Service `yaml:"service"`
  Network map[string]Network `yaml:"networks,omitempty"`
}

func ParseYAML(file string) (Config, error) {
  f, err := os.ReadFile(file)
  if err != nil {
    return Config{}, err
  }
  c := Config{}
  yaml.Unmarshal(f, &c)
  // verify all services have an associated image
  for name, value := range c.Service {
    if value.Image == "" {
      log.Fatalf("Service %s unknown Image name", name)
    }
  }

  // Order services based on dependiencies
  c.Service, err = OrderService(c.Service)
  if err != nil {
    fmt.Println(err.Error())
    os.Exit(1)
  }
  return c, nil
}

func (c *Config) CreateNetworks(cli *client.Client) error {
  networks, err := cli.NetworkList(context.Background(), network.ListOptions{})
  if err != nil {
    fmt.Println("Failed to retrieve networks list: ",err.Error())
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
      fmt.Printf("%s Network %s already exist\n", greenCheck, name)
    } else {
      opt := network.CreateOptions{
        Driver: value.Driver,
      }
      _, err := cli.NetworkCreate(context.Background(), name, opt)
      fmt.Printf("%s Network %s created\n", greenCheck, name)
      if err != nil {
        fmt.Println("Failed to create network ",name, err.Error())
        return err
      }
    }
  }
  return nil
}

func (c *Config) PullServices(cli *client.Client) error {
  for name, service := range c.Service {
    s := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
    s.Suffix = fmt.Sprintf("Pulling Service %s", name)
    s.Start()
    reader, err := cli.ImagePull(context.Background(), service.Image, image.PullOptions{})
    if err != nil {
      s.Stop()
      return err
    }
    logger.ReadProgress(reader, func(status string){
      s.Suffix = fmt.Sprintf(" Pulling Service %s - %s", name, status)
    })
    s.Stop()
  }
  return nil
}

func (c *Config) CreateAllService(cli *client.Client) (map[string][]servicemanager.Container, error) {
  runningCont := make(map[string][]servicemanager.Container)
  for name, service := range c.Service {
    container, err := service.Create(cli, 1)
    if err != nil {
      return runningCont, err
    }
    runningCont[name] = append(runningCont[name], container)
  }
  return runningCont, nil
}

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
  for name, _ := range services {
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
