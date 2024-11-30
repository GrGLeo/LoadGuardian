package main

import (
	"io"
	"math/rand"
	"net/http"
//  "github.com/docker/docker/api/types/container"
  "github.com/docker/docker/client"
)



type BackendService struct {
  ID string
  Endpoint string
  Connection int
}

type LoadBalancer struct {
  Services []BackendService
  index uint8
  DockerClient *client.Client
}

func NewLoadBalancer() (*LoadBalancer, error) {
  cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
  if err != nil {
      return nil, err
  }
  BackendServices := CreateBackendServices(cli)
  return &LoadBalancer{
    Services: BackendServices,
    index: 0,
    DockerClient: cli,
  }, nil

}
func (lb *LoadBalancer) getAlgorithm() string {
  // TODO: implement loadign algorithm from docker compose label
  return ""
}

func (lb *LoadBalancer) getNextBackend() string {
  mod := uint8(len(lb.Services))
  lb.index = (lb.index + 1) % mod
  return lb.Services[lb.index].Endpoint
}

func (lb *LoadBalancer) getRandomBackend() string {
  index := rand.Intn(len(lb.Services))
  return lb.Services[index].Endpoint
}

func (lb *LoadBalancer) getStats() {
}

func (lb *LoadBalancer) handleRequest(w http.ResponseWriter, r *http.Request) {
  BackendURL := lb.getNextBackend()
  targetURL := BackendURL + r.URL.Path

  var body io.Reader
      if r.Body != nil {
          body = r.Body
      } else {
          body = http.NoBody
      }

  req, err := http.NewRequest(r.Method, targetURL, body)
  if err != nil {
    http.Error(w, "Failed to create request", 500)
    return
  }
  for key, values := range r.Header {
    for _, value := range values {
      req.Header.Add(key, value)
    }
  }

  client := &http.Client{}
  resp, err := client.Do(req)
  if err != nil {
    http.Error(w, "Failed to forward request", 500)
    return
  }
  defer resp.Body.Close()

  for key, values := range resp.Header{
    for _, value := range values {
      w.Header().Add(key, value)
    }
  }
  
  w.WriteHeader(resp.StatusCode)
  io.Copy(w, resp.Body)
}
