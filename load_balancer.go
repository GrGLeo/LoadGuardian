package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type LoadBalancer struct {
  ServiceName string
  Algorithm string
  Services []BackendService
  index uint8
  DockerClient *client.Client
  mu sync.Mutex
}

func NewLoadBalancer() (*LoadBalancer, error) {
  cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
  if err != nil {
      return nil, err
  }

  BackendServices, ServiceName, algo := CreateBackendServices(cli)
  fmt.Println("Load Balancing started: ", ServiceName, " ", "Algorithm used: ", algo)
  return &LoadBalancer{
    ServiceName: ServiceName,
    Algorithm: algo,
    Services: BackendServices,
    index: 0,
    DockerClient: cli,
    mu: sync.Mutex{},
  }, nil
}

func (lb *LoadBalancer) RemoveDeadServices() {
  var newServices []BackendService
  for _, cont := range lb.Services {
    if cont.Healthy {
      newServices = append(newServices, cont)
    }
  }
  lb.Services = newServices
  fmt.Println(newServices)
}

func (lb *LoadBalancer) ScaleUp() {
  newID := lb.Services[0].ScaleUpService(lb.DockerClient)
  backend, err := CreateBackend(lb.DockerClient, newID)
  if err != nil {
    fmt.Println(err.Error())
  }
  lb.Services = append(lb.Services, backend)
}

func (lb *LoadBalancer) ScaleDown(index int) {
}

func (lb *LoadBalancer) handleRequest(w http.ResponseWriter, r *http.Request) {
  var backend *BackendService
  var resp *http.Response
  var err error

  for len(lb.Services) > 0 {
    backend = lb.getBackend()
    targetURL := backend.Endpoint + r.URL.Path
    resp, err = ForwardRequests(targetURL, w, r)
    // Forwarded to the service
    if err == nil {
      break
    }
    // Verifying container
    if err.Error() == "Container not responding" {
      fmt.Println("Checking container status")
      backend.CheckStatus(lb.DockerClient)
      lb.RemoveDeadServices()
    }
  }

  // No more service available
  if resp == nil {
    fmt.Println("Failed to found avaible service")
    http.Error(w, "Service unavailable", 503)
    return
  }

  atomic.AddInt32(&backend.Connection, 1)
  defer resp.Body.Close()

  // Return response
  for key, values := range resp.Header{
    for _, value := range values {
      w.Header().Add(key, value)
    }
  }
  w.WriteHeader(resp.StatusCode)
  io.Copy(w, resp.Body)
  atomic.AddInt32(&backend.Connection, -1)
}

func (lb *LoadBalancer) Monitor() error {
  err := lb.MonitorStats()
  lb.MonitorHealth()
  return err
}

func (lb *LoadBalancer) MonitorHealth()  {
  for _, cont := range lb.Services {
    cont.CheckStatus(lb.DockerClient)
    if !cont.Healthy {
      cont.RestartService(lb.DockerClient)
    }
  }
}

func (lb *LoadBalancer) MonitorStats() error {
	ctx := context.Background()
  for _, cont := range lb.Services { 
    containerID := cont.ID
    stats, err := lb.DockerClient.ContainerStats(ctx, containerID, false)
    if err != nil {
      return fmt.Errorf("failed to get container stats: %w", err)
    }
    defer stats.Body.Close()

    var statsInfo container.StatsResponse
    if err := json.NewDecoder(stats.Body).Decode(&statsInfo); err != nil {
      return fmt.Errorf("failed to decode container stats: %w", err)
    }

    // Not used yet CPU limit
    //cpuDelta := float64(statsInfo.CPUStats.CPUUsage.TotalUsage - statsInfo.PreCPUStats.CPUUsage.TotalUsage)
    //systemDelta := float64(statsInfo.CPUStats.SystemUsage - statsInfo.PreCPUStats.SystemUsage)
    //numCores := float64(len(statsInfo.CPUStats.CPUUsage.PercpuUsage))
    //cpuUsage := (cpuDelta / systemDelta) * numCores * 100.0

    memoryUsage := float64(statsInfo.MemoryStats.Usage) / (1024 * 1024) // Convert to MB
    memoryLimit := float64(cont.MemoryLimit) / (1024 * 1024) // Convert to MB
    memoryPercent := (memoryUsage / memoryLimit) * 100.0
    if memoryPercent > 80.00{
      fmt.Print("\nMemory Limit close to max\n")
    }

    // fmt.Printf("Container ID: %s\n", containerID)
    // fmt.Printf("Memory Usage: %.2f MB / %.2f MB (%.2f%%)\n", memoryUsage, memoryLimit, memoryPercent)
  }

	return nil
}

func ForwardRequests(targetURL string, w http.ResponseWriter, r *http.Request) (*http.Response, error) {
  var body io.Reader
  if r.Body != nil {
    body = r.Body
  } else {
    body = http.NoBody
  }

  req, err := http.NewRequest(r.Method, targetURL, body)
  if err != nil {
    http.Error(w, "Failed to create request", 500)
    return nil, err
  }
  for key, values := range r.Header {
    for _, value := range values {
      req.Header.Add(key, value)
    }
  }

  client := &http.Client{}
  resp, err := client.Do(req)
  if err != nil {
    return nil, errors.New("Container not responding")
  }
  return resp, nil
}

