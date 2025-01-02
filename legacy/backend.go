package main

import (
	"context"
	"errors"
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

func (back *BackendService) RestartService (cli *client.Client) error {
  timeout := 0
  stopOptions := container.StopOptions{Timeout: &timeout}
  err := cli.ContainerRestart(context.Background(), back.ID, stopOptions)
  if err != nil {
    return errors.New(fmt.Sprintf("Failed to restart container: %q", err.Error()))
  }
  fmt.Println("Container restarted: ", back.ID)
  back.Healthy = true
  return nil
}

func (back *BackendService) ScaleUpService (cli *client.Client, name string) (string, error) {
  cont, err := cli.ContainerInspect(context.Background(), back.ID)
  if err != nil {
    newErr := fmt.Sprintf("Failed to inspect container: %q", err.Error())
    return "", errors.New(newErr)
  }

  contName := fmt.Sprintf("%s-%s", name, CreateName(5))
  resp, err := cli.ContainerCreate(
    context.Background(),
    cont.Config,
    cont.HostConfig,
    nil,
    nil,
    contName,
  )
  if err != nil {
    newErr := fmt.Sprintf("Error while scaling container: %q", err.Error())
    return "", errors.New(newErr)
  }
  newID := resp.ID
  cli.ContainerStart(context.Background(), newID, container.StartOptions{})
  fmt.Println("Starting container: ", contName)
  return newID, nil
}

func (back *BackendService) ScaleDownService (cli *client.Client) error {
  timeout := 30
  stopOptions := container.StopOptions{Timeout: &timeout}
  err := cli.ContainerStop(context.Background(), back.ID, stopOptions)
  if err != nil {
    newErr := fmt.Sprintf("Failed to stop container: %q", err.Error())
    return errors.New(newErr)
  }
  err = back.RemoveContainer(cli, false)
  if err != nil {
    return err
  }
  fmt.Println("Scaled down.")
  return  nil
}

func (back *BackendService) RemoveContainer (cli *client.Client, force bool) error {
  removeOptions := container.RemoveOptions{
    RemoveVolumes: false,
    RemoveLinks: false,
    Force: force,
  }

  err := cli.ContainerRemove(context.Background(), back.ID, removeOptions)
  if err != nil {
    newErr := fmt.Sprintf("Failed to remove container: %q", err.Error())
    return errors.New(newErr)
  }
  return nil
}

func CreateName(length int) string {
  charset := "qweryuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM123456789"
  var sb strings.Builder
  for i := 0; i < length; i++ {
    sb.WriteByte(charset[rand.Intn(len(charset))])
  }
  return sb.String()
}
