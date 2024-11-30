package main

import (
    "context"
    "fmt"
    "github.com/docker/docker/api/types/container"
    "github.com/docker/docker/client"
)


func CreateBackendServices() []BackendService {
  var Services []BackendService

  cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
  if err != nil {
      panic(err)
  }
  defer cli.Close()

  containers, err := cli.ContainerList(context.Background(), container.ListOptions{All: true})
  if err != nil {
      panic(err)
  }

 for _, container := range containers {
    if container.Labels["LoadBalanced"] == "true" {
      containerInfo, err := cli.ContainerInspect(context.Background(), container.ID)
      if err != nil {
        fmt.Printf("Error inspecting container: %v", err)
      }
      fmt.Printf("Container port: %v\n", container.Names[0])

      if len(containerInfo.NetworkSettings.Ports) > 0 {
        for port, _ := range containerInfo.NetworkSettings.Ports {
          fmt.Printf("Container port: %v\n", port)
          Services = append(Services, BackendService{Endpoint: "http:/"+container.Names[0]+":"+port.Port(), Connection: 0})
        }
      } else {
        fmt.Println("No ports exposed or mapped.")
      }
    }
  }
  return Services
}
