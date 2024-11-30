package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)



type BackendService struct {
  ID string
  Endpoint string
  Connection int
  MemoryLimit int64
  CPULimit int64
}

type LoadBalancer struct {
  ServiceName string
  Services []BackendService
  index uint8
  DockerClient *client.Client
}

func NewLoadBalancer() (*LoadBalancer, error) {
  cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
  if err != nil {
      return nil, err
  }

  BackendServices, ServiceName := CreateBackendServices(cli)
  fmt.Printf("service: %q", ServiceName)
  return &LoadBalancer{
    ServiceName: ServiceName,
    Services: BackendServices,
    index: 0,
    DockerClient: cli,
  }, nil
}

func (lb *LoadBalancer) getAlgorithm() string {
  // TODO: implement loading algorithm from docker compose label
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

func (lb *LoadBalancer) getContainerStats() error {
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

    fmt.Printf("Container ID: %s\n", containerID)
    fmt.Printf("Memory Usage: %.2f MB / %.2f MB (%.2f%%)\n", memoryUsage, memoryLimit, memoryPercent)
  }

	return nil
}
