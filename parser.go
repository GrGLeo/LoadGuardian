package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
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

const greenCheck = "\033[32m✓\033[0m"

type Service struct {
  Image string `yaml:"image,omitempty"`
  Network []string `yaml:"network,omitempty"`
  Volume []string  `yaml:"volume,omitempty"`
  Port []string `yaml:"ports,omitempty"`
  Dependencies []string `yaml:"dependencies,omitempty"`
}

type Network struct {
  Driver string `yaml:"driver,omitempty"`
}

type Config struct {
  Service map[string]Service `yaml:"service"`
  Network map[string]Network `yaml:"networks,omitempty"`
}

type Container struct {
  ID string
  Name string
  Url string
}

type LoadGuardian struct {
  Client *client.Client
  Config Config
  RunningContainer map[string][]Container
}

type LogMessage struct {
  containerName string
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
  // verify all services have an associated image
  for name, value := range c.Service {
    if value.Image == "" {
      log.Fatalf("Service %s unknown Image name", name)
    }
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
    ReadProgress(reader, func(status string){
      s.Suffix = fmt.Sprintf(" Pulling Service %s - %s", name, status)
    })
    s.Stop()
  }
  return nil
}

func (c *Config) CreateAllService(cli *client.Client) (map[string][]Container, error) {
  runningCont := make(map[string][]Container)
  for name, service := range c.Service {
    container, err := service.Create(cli, 1)
    if err != nil {
      return runningCont, err
    }
    runningCont[name] = append(runningCont[name], container)
  }
  return runningCont, nil
}

func (s *Service) Create(cli *client.Client, n int) (Container, error) {
  var cport string
  var hport string
  
  if len(s.Port) != 0 {
    ports := strings.Split(s.Port[0], ":")
    cport = ports[0]
    hport = ports[1]
  }

  config := &container.Config{
    Image: s.Image,
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

func (c *Container) Start(cli *client.Client) error {
  err := cli.ContainerStart(context.Background(), c.ID, container.StartOptions{})
  if err != nil {
    fmt.Println(err.Error())
    return err
  }
  return nil
}

func (c *Container) FetchLogs(cli *client.Client, logChannel chan<- LogMessage) error {
  logsOpt := container.LogsOptions{
    ShowStdout: true,
    ShowStderr: true,
    Follow: true,
  }

  reader, err := cli.ContainerLogs(context.Background(), c.ID, logsOpt)
  if err != nil {
    fmt.Println(err.Error())
    return err 
  }
  defer reader.Close()
  scanner := bufio.NewScanner(reader)
  buf := make([]byte, 0, 64*1024)
  scanner.Buffer(buf, 1024*1024)
  for scanner.Scan() {
    logChannel <- LogMessage{
      containerName: c.Name,
      Message: scanner.Text(),
    }
  }
  if err := scanner.Err(); err != nil {
    fmt.Println(err.Error())
    return err
  }
  return  nil
}

func (c *Container) StartAndFetchLogs(cli *client.Client, logChannel chan<- LogMessage) error {
    err := c.Start(cli)
    greenCheck := "\033[32m✓\033[0m"
    fmt.Printf("%s %s started.\n", greenCheck, c.Name)
    if err != nil {
        return fmt.Errorf("failed to start container: %w", err)
    }
    // Fetch the logs in a separate goroutine
    go func() {
        err := c.FetchLogs(cli, logChannel)
        if err != nil {
            logChannel <- LogMessage{
                containerName: c.Name,
                Message: fmt.Sprintf("error fetching logs: %s", err),
            }
        }
    }()
    return nil
}

func (c *Container) Stop(cli *client.Client, timeout *int) error {
  opt := container.StopOptions{
    Timeout: timeout, 
  }
  err := cli.ContainerStop(context.Background(), c.ID, opt) 
  if err != nil {
    fmt.Printf("Error while stopping container: %s\n", c.Name)
    return err
  }
  return nil
}

func (lg *LoadGuardian) StopAll(timeout int) error {
  fmt.Println("Stopping all container")
  for name, containers := range lg.RunningContainer {
    fmt.Printf("Stopping services: %s\n", name)
    for _, c := range containers {
      err := c.Stop(lg.Client, &timeout)
      if err != nil {
        return err
      }
    }
  }
  return nil
}

func PrintLogs(logChannel <-chan LogMessage) {
  for logMessage := range logChannel {
    cleanMessage := sanitizeLogMessage(logMessage.Message)
    fmt.Printf("[Container: %s] %s\n",logMessage.containerName, cleanMessage)
  }
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

func sanitizeLogMessage(message string) string {
  // Remove non-printable characters using a regex
  re := regexp.MustCompile(`[[:cntrl:]]`)
  clean := strings.TrimSpace(re.ReplaceAllString(message, ""))
  return stripAnsiCodes(clean)
}

func stripAnsiCodes(input string) string {
  // Strip ANSI codes from a string
  var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)
  return ansiRegex.ReplaceAllString(input, "")
}
