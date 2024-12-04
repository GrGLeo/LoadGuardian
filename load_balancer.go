package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type LoadBalancer struct {
  ServiceName string
  Algorithm string
  Services []BackendService
  index uint8
  DockerClient *client.Client
  LastScaled time.Time
  CooldownPeriod time.Duration
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
    LastScaled: time.Now(),
    CooldownPeriod: 3 * time.Minute,
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
  lb.LastScaled = time.Now()
}

func (lb *LoadBalancer) ScaleDown(index int) {
  // Avoid scaling down newly created container
  if time.Since(lb.LastScaled) > lb.CooldownPeriod {
    lb.Services[index].ScaleDownService(lb.DockerClient)
    if len(lb.Services) == 1 {
      lb.Services = []BackendService{lb.Services[0]}
    } else {
      lb.Services = append(lb.Services[:index], lb.Services[index+1:]...)
    }
  }
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
  for i, cont := range lb.Services { 
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
      fmt.Println("Memory Limit close to max")
      lb.ScaleUp()
    }
    if memoryPercent < 40.0{
      if len(lb.Services) > 1 {
        fmt.Println("Low usage, scaling down")
        lb.ScaleDown(i)
      }
    }

    // fmt.Printf("Container ID: %s\n", containerID)
    // fmt.Printf("Memory Usage: %.2f MB / %.2f MB (%.2f%%)\n", memoryUsage, memoryLimit, memoryPercent)
  }

	return nil
}


