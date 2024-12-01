package main

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type BackendService struct {
  ID string
  Endpoint string
  Connection int32
  MemoryLimit int64
  CPULimit int64
  Healthy bool
}

func (back *BackendService) CheckStatus (cli *client.Client) {
  container, err := cli.ContainerInspect(context.Background(), back.ID)
  if err != nil {
    fmt.Println("Failed to inspect container")
  }
  if container.State.Status == "dead" || container.State.Status == "exited" {
    back.Healthy = false
  }
}

func (back *BackendService) RestartService (cli *client.Client) {
  timeout := 0
  stopOptions := container.StopOptions{Timeout: &timeout}
  err := cli.ContainerRestart(context.Background(), back.ID, stopOptions)
  if err != nil {
    fmt.Println("Failed to restart container.")
  }
}
