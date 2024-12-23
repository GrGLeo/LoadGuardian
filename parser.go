package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"gopkg.in/yaml.v3"
)

type Service struct {
  Image string `yaml:"image,omitempty"`
  Network []string `yaml:"network,omitempty"`
  Port []string `yaml:"ports,omitempty"`
}

type Config struct {
  Service map[string]Service `yaml:"service"`
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

func PullServices (c *Config) error {
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

func ReadProgress(r io.ReadCloser) error {
  decoder := json.NewDecoder(r)
  for {
    var msg map[string]interface{}
    if err := decoder.Decode(&msg); err == io.EOF {
      break
    } else if err != nil {
      log.Fatal(err)
    }
    fmt.Println(msg)
  }
  return nil 
}
