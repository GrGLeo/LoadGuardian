package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type Replicas struct {
  MinReplicas int
  MaxReplicas int
  LastScaled time.Time
  CooldownPeriod time.Duration
}

type LoadBalancer struct {
  DockerClient *client.Client
  ServiceName string
  Algorithm string
  Services []BackendService
  index uint8
  mu sync.Mutex
  Replicas Replicas
}

func NewLoadBalancer() (*LoadBalancer, error) {
  cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
  if err != nil {
      return nil, err
  }

  BackendServices, ServiceName, algo, replicasInfo := CreateBackendServices(cli)
  fmt.Println("Load Balancing started: ", ServiceName, " ", "Algorithm used: ", algo)

  replicas := Replicas{
    MinReplicas: replicasInfo[0],
    MaxReplicas: replicasInfo[1],
    LastScaled: time.Now(),
    CooldownPeriod: 3 * time.Minute,
  }

  return &LoadBalancer{
    ServiceName: ServiceName,
    Algorithm: algo,
    Services: BackendServices,
    index: 0,
    DockerClient: cli,
    mu: sync.Mutex{},
    Replicas: replicas,
  }, nil
}

func (lb *LoadBalancer) RemoveDeadServices() {
  fmt.Println("Hey there")
  var newServices []BackendService
  for _, cont := range lb.Services {
    if cont.Healthy {
      newServices = append(newServices, cont)
    }
  }
  lb.mu.Lock()
  lb.Services = newServices
  lb.mu.Unlock()
}

func (lb *LoadBalancer) ScaleUp() error {
  lb.mu.Lock()
  defer lb.mu.Unlock()
   
  canScale := time.Since(lb.Replicas.LastScaled) > lb.Replicas.CooldownPeriod
  maxReach := len(lb.Services) >= lb.Replicas.MaxReplicas
  // Check if we can scale
  fmt.Println(len(lb.Services), lb.Replicas.MaxReplicas)
  if canScale && !maxReach {
    newID := lb.Services[0].ScaleUpService(lb.DockerClient, lb.ServiceName)
    backend, err := CreateBackend(lb.DockerClient, newID)
    if err != nil {
      return err
    }
    lb.Services = append(lb.Services, backend)
    lb.Replicas.LastScaled = time.Now()
  } else {
    // We do nothing is max replicas is reach to avoid conflict with port
    // Or if cooldown period not yet passed
    reason := "Cooldown period not reach."
    if maxReach {
      reason = "Max replica reached."
    }
    return errors.New(reason)
  }
  return nil
}

func (lb *LoadBalancer) ScaleDown(index int) {
  // Avoid scaling down newly created container
  // No scaling down when min replicas reach
  if time.Since(lb.Replicas.LastScaled) > lb.Replicas.CooldownPeriod && len(lb.Services) > lb.Replicas.MinReplicas {
    lb.Services[index].ScaleDownService(lb.DockerClient)
    lb.mu.Lock()
    if len(lb.Services) == 1 {
      lb.Services = []BackendService{lb.Services[0]}
    } else {
      lb.Services = append(lb.Services[:index], lb.Services[index+1:]...)
    }
    lb.mu.Unlock()
  }
}

func (lb *LoadBalancer) Monitor() error {
  fmt.Println(lb.Services)
  lb.MonitorHealth()
  err := lb.MonitorStats()
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
      err := lb.ScaleUp()
      if err != nil {
        fmt.Println(err.Error())
      }
    }
    if memoryPercent < 20.0{
      if len(lb.Services) > 1 {
        lb.ScaleDown(i)
      }
    }

    // fmt.Printf("Container ID: %s\n", containerID)
    // fmt.Printf("Memory Usage: %.2f MB / %.2f MB (%.2f%%)\n", memoryUsage, memoryLimit, memoryPercent)
  }

	return nil
}
