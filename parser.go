package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
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

type LogMessage struct {
  containerID string
  Message string
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
    s := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
    s.Suffix = fmt.Sprintf("Pulling Service %s", name)
    s.Start()
    reader, err := cli.ImagePull(context.Background(), service.Image, image.PullOptions{})
    if err != nil {
      s.Stop()
      return err
    }
    ReadProgress(reader, func(status string){
      s.Suffix = fmt.Sprintf(" Pulling Service %s - %s", name, status)
    })
    s.Stop()
  }
  return nil
}

func (s *Service) CreateService(cli *client.Client) (string, error) {
  var cport string
  var hport string
  if len(s.Port) != 0 {
    ports := strings.Split(s.Port[0], ":")
    cport = ports[0]
    hport = ports[1]
  }
  config := &container.Config{
    Image: s.Image,
    ExposedPorts: nat.PortSet{
      nat.Port(cport+"/tcp"): struct{}{},
    },
  }
  hostConfig := &container.HostConfig{
    PortBindings: nat.PortMap{
      nat.Port(hport+"/tcp"): []nat.PortBinding{
        {
          HostIP: "0.0.0.0",
          HostPort: hport,
        },
      },
    },
    NetworkMode: container.NetworkMode(s.Network[0]),
  }

  resp, err := cli.ContainerCreate(context.Background(), config, hostConfig, nil, nil, "hello") 
  if err != nil {
    return "", err
  }

  return resp.ID, nil
}

func (c *Config) ServiceStart(cli *client.Client, id string) error {
  err := cli.ContainerStart(context.Background(), id, container.StartOptions{})
  if err != nil {
    return err
  }
  return nil
}

func (s *Service) FetchLogs(cli *client.Client, id string, logChannel chan<- LogMessage) error {
  return  nil
}


func ReadProgress(r io.ReadCloser, updateStatus func(string)) error {
  defer r.Close()
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
    if status, ok := msg["status"].(string); ok {
      updateStatus(status)
    }
  }
  return nil 
}
