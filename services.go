package main

import (
    "context"
    "fmt"
    "github.com/docker/docker/api/types/container"
    "github.com/docker/docker/client"
)


func CreateBackendServices() []BackendService {
  var Services []BackendService

  // Create Docker client
  cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
  if err != nil {
      panic(err)
  }
  defer cli.Close()

  // List all container
  containers, err := cli.ContainerList(context.Background(), container.ListOptions{All: true})
  if err != nil {
      panic(err)
  }

  // Loop through the containers & only continue on labelled one
  for _, container := range containers {
    if container.Labels["LoadBalanced"] == "true" {

      containerInfo, err := cli.ContainerInspect(context.Background(), container.ID)
      if err != nil {
        fmt.Printf("Error inspecting container: %v", err)
      }
      // TODO: there should be a better way to do this.
      if len(containerInfo.NetworkSettings.Ports) > 0 {
        for port := range containerInfo.NetworkSettings.Ports {
          Services = append(Services, BackendService{ID: container.ID, Endpoint: "http:/"+container.Names[0]+":"+port.Port(), Connection: 0})
        }
      } else {
        fmt.Println("No ports exposed or mapped.")
      }
    }
  }
  return Services
}
