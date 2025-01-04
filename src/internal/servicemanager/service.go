package servicemanager

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/GrGLeo/LoadBalancer/src/pkg/utils"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	dockerspec "github.com/moby/docker-image-spec/specs-go/v1"
)

type Service struct {
  Image string `yaml:"image,omitempty"`
  Network []string `yaml:"network,omitempty"`
  Volume []string  `yaml:"volume,omitempty"`
  Port []string `yaml:"ports,omitempty"`
  Envs []string `yaml:"envs,omitempty"`
  Replicas int `yaml:"replicas,omitempty"`
  HealthCheck HealthcheckConfig `yaml:"healthcheck,omitempty"`
  Dependencies []string `yaml:"dependencies,omitempty"`
  NextPort *atomic.Uint32 `yaml:"-"`
}

type HealthcheckConfig struct {
  Test []string `yaml:"test,omitempty"`
	// Zero means to inherit. Durations are expressed as integer nanoseconds.
	Interval      time.Duration `yaml:"interval,omitempty"` // Interval is the time to wait between checks.
	Timeout       time.Duration `yaml:"timeout,omitempty"` // Timeout is the time to wait before considering the check to have hung.
	StartPeriod   time.Duration `yaml:"startperiod,omitempty"` // The start period for the container to initialize before the retries starts to count down.
	StartInterval time.Duration `yaml:"startinterval,omitempty"` // The interval to attempt healthchecks at during the start period

	// Retries is the number of consecutive failures needed to consider a container as unhealthy.
	// Zero means inherit.
	Retries int `yaml:"retries,omitempty"`
}

func (s *Service) Create(cli *client.Client) (Container, error) {
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

  // If a healtcheck in implemented we configure the container with it
  // otherwise we add a base HealthCheck
  config.Healthcheck = s.CreateHealthCheck()
 
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

  if len(s.Volume) > 0 {
    for _, vol := range s.Volume {
      paths := strings.Split(vol, ":")
      source := paths[0]
      target := paths[1]

      hostConfig.Mounts = append(hostConfig.Mounts, mount.Mount{
        Type: mount.TypeBind,
        Source: source,
        Target: target,
      })
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
  name = name + "-" + utils.GenerateName(5) 
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
  // TODO: how to handle replicas??
  return false
}


func (s *Service) CreateHealthCheck() *dockerspec.HealthcheckConfig {
    hcConfig := dockerspec.HealthcheckConfig{}
    if s.HealthCheck.Test != nil {
      hcConfig.Test = s.HealthCheck.Test
      hcConfig.Interval = s.HealthCheck.Interval * time.Second
      hcConfig.Timeout = s.HealthCheck.Timeout * time.Second
      hcConfig.Retries = s.HealthCheck.Retries
    } else {
      hcConfig.Test = []string{"CMD", "true"}
      hcConfig.Interval = 30  * time.Second 
      hcConfig.Timeout = 10 * time.Second 
      hcConfig.Retries =  3
    }
    if s.HealthCheck.StartInterval != 0 {
      hcConfig.StartInterval = s.HealthCheck.StartInterval
    }
    if s.HealthCheck.StartPeriod != 0 {
      hcConfig.StartPeriod = s.HealthCheck.StartPeriod
    }
    return &hcConfig
}
