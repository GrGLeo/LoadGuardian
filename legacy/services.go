package main

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

// Return multiple value or a Struct return?
func CreateBackendServices(cli *client.Client) ([]BackendService, string, string, [2]int, error) {
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
    if v, ok := container.Labels["lg.BalanceAlgorithm"]; ok {
      algo = v
    }
    // Extact info on service that is balance
    if container.Labels["lg.monitor"] == "true" {
      serviceName = container.Labels["com.docker.compose.service"]
      replicas, err := strconv.Atoi(container.Labels["lg.min.replicas"])
      if err != nil {
        fmt.Println("Error getting min replicas, setting min to 1")
        Replicas[0] = 1
      }
      Replicas[0] = replicas
      replicas, err = strconv.Atoi(container.Labels["lg.max.replicas"])
      if err != nil {
        fmt.Println("Error getting max replicas, setting min to 2")
        Replicas[1] = 2
      }
      Replicas[1] = replicas

      backend, err := CreateBackend(cli, container.ID)
      if err != nil {
        fmt.Println(err)
        continue
      }
      Services = append(Services, backend)
    }
  }
  if len(Services) == 0 {
    return []BackendService{}, "", "", [2]int{0,0}, errors.New("Failed to create Services")
  }
  return Services, serviceName, algo, Replicas, nil
}

func CreateBackend(cli *client.Client, containerID string) (BackendService, error) {
  containerInfo, err := cli.ContainerInspect(context.Background(), containerID)
  if err != nil {
    return BackendService{}, errors.New(fmt.Sprintf("Error inspecting container: %v", err))
  }
  // extract memory limit
  Memory := containerInfo.HostConfig.Memory

  // TODO: there sould be a better way to do this
  if len(containerInfo.NetworkSettings.Ports) > 0 {
    for port := range containerInfo.NetworkSettings.Ports {
      backend := BackendService{
        ID: containerID,
        Endpoint: "http:/"+containerInfo.Name+":"+port.Port(),
        Connection: 0,
        MemoryLimit: Memory,
        Healthy: true,
      }
      return backend, nil
    }
  } else {
    return BackendService{}, errors.New("No ports exposed or mapped.")
  }
  return BackendService{}, errors.New("Error while getting service info")
}
