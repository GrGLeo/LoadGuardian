package servicemanager

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/GrGLeo/LoadBalancer/src/pkg/utils"
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
  NextPort *atomic.Uint32 `yaml:"-"`
}

func (s *Service) Create(cli *client.Client, n int) (Container, error) {
  ports := []int{-1,-1}
  var err error
  if len(s.Port) > 0 {
    fmt.Println(s.Image, "Port: ", s.NextPort.Load())
    ports, err = s.GetPort()
    if err != nil {
      fmt.Println("Failed to read port, Ports will not be set")
    }
  }
  
  cport := strconv.Itoa(ports[0])
  hport := strconv.Itoa(ports[1])

  config := &container.Config{
    Image: s.Image,
    Env: s.Envs,
  }
  if cport != "-1" {
    config.ExposedPorts = nat.PortSet{
      nat.Port(cport+"/tcp"): struct{}{},
    }
  }
 
  hostConfig := &container.HostConfig{}
  if len(s.Network) > 0 {
    hostConfig.NetworkMode = container.NetworkMode(s.Network[0])
  }
  if hport != "-1" {
    hostConfig.PortBindings = nat.PortMap{
      nat.Port(hport+"/tcp"): []nat.PortBinding{
        {
          HostIP: "0.0.0.0",
          HostPort: hport,
        },
      },
    }
  }

  var name string
  names := strings.Split(s.Image, "/")
  if len(names) == 1 {
    name = names[0]
  } else {
    idx := len(names) - 1
    name = names[idx]
  }
  resp, err := cli.ContainerCreate(context.Background(), config, hostConfig, nil, nil, name) 
  ContainerID := resp.ID
  if err != nil {
    fmt.Println("Failed to create container: ", name, err.Error())
    return Container{}, err
  }
  name = name + "-" + strconv.Itoa(n)
  err = cli.ContainerRename(context.Background(), ContainerID, name)
  if err != nil {
    fmt.Println(err.Error())
    return Container{}, err
  }
  // Setting the external port
  intPort := -1
  intPort, _ = strconv.Atoi(hport)
  fmt.Println(intPort)

  return Container{
    ID: ContainerID,
    Name: name,
    Port: intPort,
  }, nil
}


func (old *Service) Compare(new *Service) bool {
  // image
  if old.Image != new.Image {
    return true
  }
  // networks
  if diff := utils.CompareStrings(true, old.Network, new.Network); diff {
    return true
  }
  // volume
  if diff := utils.CompareStrings(true, old.Volume, new.Volume); diff {
    return true
  }
  // port
  if diff := utils.CompareStrings(true, old.Port, new.Port); diff {
    return true
  }
  // envs
  if diff := utils.CompareStrings(true, old.Envs, new.Envs); diff {
    return true
  }
  // dependencies
  if diff := utils.CompareStrings(false, old.Dependencies, new.Dependencies); diff {
    return true
  }
  return false
}

