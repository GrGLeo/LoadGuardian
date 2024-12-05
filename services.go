package main

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)


func CreateBackendServices(cli *client.Client) ([]BackendService, string, string, [2]int) {
  var Services []BackendService
  var serviceName string
  var Replicas [2]int
  algo := "random"
  // TODO: add filter on both labels to avoid looping all container
  // List all container
  containers, err := cli.ContainerList(context.Background(), container.ListOptions{All: true})
  if err != nil {
      panic(err)
  }

  // Loop through the containers & only continue on labelled one
  for _, container := range containers {
    // Get load balancing algorithm
    if v, ok := container.Labels["BalanceAlgorithm"]; ok {
      algo = v
    }
    // Extact info on service that is balance
    if container.Labels["LoadBalanced"] == "true" {
      serviceName = container.Labels["com.docker.compose.service"]
      replicas, err := strconv.Atoi(container.Labels["LoadBalancer.min.replicas"])
      if err != nil {
        fmt.Println("Error getting min replicas, setting min to 1")
        Replicas[0] = 1
      }
      Replicas[0] = replicas
      replicas, err = strconv.Atoi(container.Labels["LoadBalancer.max.replicas"])
      if err != nil {
        fmt.Println("Error getting max replicas, setting min to 2")
        Replicas[1] = 2
      }
      Replicas[1] = replicas

      containerInfo, err := cli.ContainerInspect(context.Background(), container.ID)
      if err != nil {
        fmt.Printf("Error inspecting container: %v", err)
      }
      // extract memory limit
      Memory := containerInfo.HostConfig.Memory

      // TODO: there should be a better way to do this.
      if len(containerInfo.NetworkSettings.Ports) > 0 {
        for port := range containerInfo.NetworkSettings.Ports {
          backend := BackendService{
            ID: container.ID,
            Endpoint: "http:/"+container.Names[0]+":"+port.Port(),
            Connection: 0,
            MemoryLimit: Memory,
            // bold assumption
            Healthy: true,
          }
          Services = append(Services, backend)
        }
      } else {
        fmt.Println("No ports exposed or mapped.")
      }
    }
  }
  return Services, serviceName, algo, Replicas
}

func CreateBackend(cli *client.Client, containerID string) (BackendService, error) {
  containerInfo, err := cli.ContainerInspect(context.Background(), containerID)
  if err != nil {
    fmt.Printf("Error inspecting container: %v", err)
  }
  // extract memory limit
  Memory := containerInfo.HostConfig.Memory

  // TODO: there should be a better way to do this.
  if len(containerInfo.NetworkSettings.Ports) > 0 {
    for port := range containerInfo.NetworkSettings.Ports {
      backend := BackendService{
        ID: containerID,
        Endpoint: "http:/"+containerInfo.Name+":"+port.Port(),
        Connection: 0,
        MemoryLimit: Memory,
        // bold assumption
        Healthy: true,
      }
      return backend, nil
    }
  } else {
    fmt.Println("No ports exposed or mapped.")
  }
  return BackendService{}, errors.New("Error while getting service info")
}
