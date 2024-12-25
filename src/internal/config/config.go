package config

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	servicemanager "github.com/GrGLeo/LoadBalancer/src/internal/container"
	"github.com/GrGLeo/LoadBalancer/src/pkg/logger"
	"github.com/briandowns/spinner"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

const greenCheck = "\033[32m✓\033[0m"

type Network struct {
  Driver string `yaml:"driver,omitempty"`
}

type Config struct {
  Service map[string]servicemanager.Service `yaml:"service"`
  Network map[string]Network `yaml:"networks,omitempty"`
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

func (c *Config) CheckServices(cli *client.Client) (map[string]bool, error) {
  // Retrieve existing image
  resp, err := cli.ImageList(context.Background(), image.ListOptions{})
  if err != nil {
    return nil, err
  }
  // Checking if service already have an existing image
  pullingServices := make(map[string]bool)
  for _, serv := range c.Service {
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


func (c *Config) PullServices(cli *client.Client) error {
  ImageToPull, err:= c.CheckServices(cli)
  if err != nil {
    fmt.Println("Failed to inspect images. Pulling image for all services")
  }
  for name, service := range c.Service {
    // Checking if image need to be pulled
    value, ok := ImageToPull[service.Image]
    if ok && !value {
      fmt.Printf("%s Service %s already pulled\n", greenCheck, name)
      continue
    }
    // Pulling Image
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
