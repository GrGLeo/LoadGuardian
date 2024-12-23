package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/briandowns/spinner"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"gopkg.in/yaml.v3"
)

type Service struct {
  Image string `yaml:"image,omitempty"`
  Network []string `yaml:"network,omitempty"`
  Volume []string  `yaml:"volume,omitempty"`
  Port []string `yaml:"ports,omitempty"`
}

type Network struct {
  Driver string `yaml:"driver,omitempty"`
}


type Config struct {
  Service map[string]Service `yaml:"service"`
  Network map[string]Network `yaml:"networks,omitempty"`
}

type LoadGuardian struct {
  Client *client.Client
  Config Config
}

func NewLoadGuardian(file string) (LoadGuardian, error) {
  cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
  if err != nil {
    return LoadGuardian{}, err
  }
  c, err := ParseYAML(file)
  if err != nil {
    return LoadGuardian{}, err
  }

  return LoadGuardian{
    Client: cli,
    Config: c,
  }, nil
}
    

func ParseYAML(file string) (Config, error) {
  f, err := os.ReadFile(file)
  if err != nil {
    return Config{}, err
  }
  c := Config{}
  yaml.Unmarshal(f, &c)
  return c, nil
}

func (c *Config) CreateNetworks(client *client.Client) error {
  for name, value := range c.Network {
    opt := network.CreateOptions{
      Driver: value.Driver,
    }
    s := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
    s.Suffix = fmt.Sprintf("Pulling Service %s", name)
    s.Start()
    _, err := client.NetworkCreate(context.Background(), name, opt)
    s.Stop()
    if err != nil {
      return err
    }
  }
  return nil
}


func (c *Config) PullServices() error {
  cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
  if err != nil {
      return err
  }
  for name, service := range c.Service {
    fmt.Println(fmt.Sprintf("Pulling Service %s", name))
    // s := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
    // s.Suffix = fmt.Sprintf("Pulling Service %s", name)
    // s.Start()
    reader, err := cli.ImagePull(context.Background(), service.Image, image.PullOptions{})
    ReadProgress(reader)
    if err != nil {
      return err
    }
    
    // s.Stop()
  }
  return nil
}

func (s *Service) CreateService() error {
  cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
  if err != nil {
      return err
  }
  config := &container.Config{
  }
  hostConfig := &container.HostConfig{
  }

  cli.ContainerCreate(context.Background(), config, hostConfig, nil, nil, "hello") 

  return nil
}


func ReadProgress(r io.ReadCloser) error {
  decoder := json.NewDecoder(r)
  for {
    var msg map[string]interface{}
    if err := decoder.Decode(&msg); err == io.EOF {
      break
    } else if err != nil {
      log.Fatal(err)
    }
    if id, ok := msg["id"]; ok {
      fmt.Printf("Image Id: %s\n", id)
    }
    if status, ok := msg["status"]; ok {
      fmt.Printf("Status: %s\n", status)
    }
    fmt.Println("---")
  }
  return nil 
}
