package servicemanager

import (
	"bufio"
	"context"
	"fmt"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type Container struct {
  ID string
  Name string
  Url string
}

type LogMessage struct {
  ContainerName string
  Message string
}

func (c *Container) Start(cli *client.Client) error {
  err := cli.ContainerStart(context.Background(), c.ID, container.StartOptions{})
  if err != nil {
    fmt.Println(err.Error())
    return err
  }
  return nil
}

func (c *Container) FetchLogs(cli *client.Client, logChannel chan<- LogMessage) error {
  logsOpt := container.LogsOptions{
    ShowStdout: true,
    ShowStderr: true,
    Follow: true,
  }

  reader, err := cli.ContainerLogs(context.Background(), c.ID, logsOpt)
  if err != nil {
    fmt.Println(err.Error())
    return err 
  }
  defer reader.Close()
  scanner := bufio.NewScanner(reader)
  buf := make([]byte, 0, 64*1024)
  scanner.Buffer(buf, 1024*1024)
  for scanner.Scan() {
    logChannel <- LogMessage{
      ContainerName: c.Name,
      Message: scanner.Text(),
    }
  }
  if err := scanner.Err(); err != nil {
    fmt.Println(err.Error())
    return err
  }
  return  nil
}

func (c *Container) StartAndFetchLogs(cli *client.Client, logChannel chan<- LogMessage) error {
    err := c.Start(cli)
    greenCheck := "\033[32m✓\033[0m"
    fmt.Printf("%s %s started.\n", greenCheck, c.Name)
    if err != nil {
        return fmt.Errorf("failed to start container: %w", err)
    }
    // Fetch the logs in a separate goroutine
    go func() {
        err := c.FetchLogs(cli, logChannel)
        if err != nil {
            logChannel <- LogMessage{
                ContainerName: c.Name,
                Message: fmt.Sprintf("error fetching logs: %s", err),
            }
        }
    }()
    return nil
}

func (c *Container) Stop(cli *client.Client, timeout *int) error {
  opt := container.StopOptions{
    Timeout: timeout, 
  }
  err := cli.ContainerStop(context.Background(), c.ID, opt) 
  if err != nil {
    fmt.Printf("Error while stopping container: %s\n", c.Name)
    return err
  }
  return nil
}

func (c *Container) Remove(cli *client.Client) error {
 opt := container.RemoveOptions{}
 err := cli.ContainerRemove(context.Background(), c.ID, opt)
  if err != nil {
    fmt.Printf("Error while removing container: %s\n", c.Name)
    return err
  }
  return nil
}

