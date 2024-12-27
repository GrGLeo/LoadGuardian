package config

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync/atomic"

	"gopkg.in/yaml.v3"
)

func ParseYAML(file string) (Config, error) {
  f, err := os.ReadFile(file)
  if err != nil {
    return Config{}, err
  }
  c := Config{}
  yaml.Unmarshal(f, &c)
  // verify all services have an associated image
  // Set the envs variable if needed
  for name, value := range c.Service {
    if value.Image == "" {
      log.Fatalf("Service %s unknown Image name", name)
    }
    value.NextPort = &atomic.Uint32{}

    if len(value.Envs) > 0 {
      setEnvs := ParseEnvs(value.Envs)
      value.Envs = setEnvs
    }
    // Set the new service
    c.Service[name] = value
  }
  
  // Order services based on dependiencies
  c.Service, err = OrderService(c.Service)
  if err != nil {
    fmt.Println(err.Error())
    os.Exit(1)
  }
  return c, nil
}

func ParseEnvs(envs []string) []string {
  var parsedEnvs []string
  for _, env := range envs {
    if strings.HasPrefix(env, "$"){
      name := env[1:]
      setEnvs := os.Getenv(name)
      parsedEnvs = append(parsedEnvs, fmt.Sprintf("%s=%s", name, setEnvs))
    }
  }
  return parsedEnvs 
}
