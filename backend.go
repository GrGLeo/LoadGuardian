package main

import (
	"context"
	"fmt"
	"math/rand"
	"strings"

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
  cont, err := cli.ContainerInspect(context.Background(), back.ID)
  if err != nil {
    fmt.Println("Failed to inspect container")
  }
  if cont.State.Status == "dead" || cont.State.Status == "exited" {
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
  fmt.Println("Container started: ", back.ID)
}

func (back *BackendService) ScaleUpService (cli *client.Client) string {
  cont, err := cli.ContainerInspect(context.Background(), back.ID)
  if err != nil {
    fmt.Println("Failed to inspect container")
  }

  resp, err := cli.ContainerCreate(
    context.Background(),
    cont.Config,
    cont.HostConfig,
    nil,
    nil,
    CreateName(5),
  )
  if err != nil {
    fmt.Println("Error while scaling container: ", err.Error())
  }
  newID := resp.ID
  fmt.Println("Scaled up: ", newID)
  cli.ContainerStart(context.Background(), newID, container.StartOptions{})
  fmt.Println("Starting container: ", newID)
  return newID
}

func (back *BackendService) ScaleDownService (cli *client.Client) string {
  timeout := 30
  stopOptions := container.StopOptions{Timeout: &timeout}
  err := cli.ContainerStop(context.Background(), back.ID, stopOptions)
  if err != nil {
    fmt.Println("Failed to stop container")
  }
  back.RemoveContainer(cli, false)
  back.Healthy = false
  fmt.Println("Scaled down.")
  return back.ID
}

func (back *BackendService) RemoveContainer (cli *client.Client, force bool) {
  removeOptions := container.RemoveOptions{false, false, force}
  err := cli.ContainerRemove(context.Background(), back.ID, removeOptions)
  if err != nil {
    fmt.Println("Failed to remove container: ", err.Error())
  }
}

func CreateName(length int) string {
  charset := "qweryuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM123456789"
  var sb strings.Builder
  for i := 0; i < length; i++ {
    sb.WriteByte(charset[rand.Intn(len(charset))])
  }
  return sb.String()
}
