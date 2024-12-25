package servicemanager 

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

type Service struct {
  Image string `yaml:"image,omitempty"`
  Network []string `yaml:"network,omitempty"`
  Volume []string  `yaml:"volume,omitempty"`
  Port []string `yaml:"ports,omitempty"`
  Envs []string `yaml:"envs,omitempty"`
  Dependencies []string `yaml:"dependencies,omitempty"`
}

func (s *Service) Create(cli *client.Client, n int) (Container, error) {
  var cport string
  var hport string
  
  if len(s.Port) != 0 {
    ports := strings.Split(s.Port[0], ":")
    cport = ports[0]
    hport = ports[1]
  }

  fmt.Println(s.Envs)
  config := &container.Config{
    Image: s.Image,
    Env: s.Envs,
  }
  if cport != "" {
    config.ExposedPorts = nat.PortSet{
      nat.Port(cport+"/tcp"): struct{}{},
    }
  }
 
  hostConfig := &container.HostConfig{}
  if len(s.Network) > 0 {
    hostConfig.NetworkMode = container.NetworkMode(s.Network[0])
  }
  if hport != "" {
    hostConfig.PortBindings = nat.PortMap{
      nat.Port(hport+"/tcp"): []nat.PortBinding{
        {
          HostIP: "0.0.0.0",
          HostPort: hport,
        },
      },
    }
  }

  resp, err := cli.ContainerCreate(context.Background(), config, hostConfig, nil, nil, "hello") 
  ContainerID := resp.ID
  if err != nil {
    fmt.Println(err.Error())
    return Container{}, err
  }
  var name string
  names := strings.Split(s.Image, "/")
  if len(names) == 1 {
    name = names[0]
  } else {
    idx := len(names) - 1
    name = names[idx]
  }
  name = name + "-" + strconv.Itoa(n)
  err = cli.ContainerRename(context.Background(), ContainerID, name)
  if err != nil {
    fmt.Println(err.Error())
    return Container{}, err
  }
  return Container{
    ID: ContainerID,
    Name: name,
  }, nil
}
