package servicemanager

import (
	"bufio"
	"context"
	"fmt"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"go.uber.org/zap"
)

type Container struct {
  ID string
  Name string
  Url string
  HealthCheck bool
  Port int
}

type ContainerRollbackConfig struct {
  PastService Service
  New Container
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

func (c *Container) StartAndFetchLogs(cli *client.Client, logger *zap.SugaredLogger, logChannel chan<- LogMessage) error {
    err := c.Start(cli)
    if err != nil {
        return fmt.Errorf("failed to start container: %w", err)
    }
    logger.Infow("Container started", "container", c.Name)
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

func (c *Container) Stop(cli *client.Client, logger *zap.SugaredLogger, timeout *int) error {
  opt := container.StopOptions{
    Timeout: timeout, 
  }
  err := cli.ContainerStop(context.Background(), c.ID, opt) 
  if err != nil {

    logger.Errorw("Error while stopping container", "container", c.Name)
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

func (c *Container) RunningCheck(cli *client.Client) (bool, error) {
  inspect, err := cli.ContainerInspect(context.Background(), c.ID)
  if err != nil {
    return false, err
  }
  health := inspect.State.Status
  if health == "running" {
    return true, nil
  }
  return false, nil
}

func (c *Container) HealthChecker(cli *client.Client) (string, error) {
  inspect, err := cli.ContainerInspect(context.Background(), c.ID)
  if err != nil {
    return "", err
  }
  if inspect.State.Health == nil {
    return "", ErrContainerNotStarted
  }
  health := inspect.State.Health.Status
  return health, nil
}

func CheckContainerHealth(cli *client.Client, container Container, logger *zap.SugaredLogger) bool {
  retries := 0
  maxRetries := 5
  delay := 5 * time.Second

  for retries < maxRetries {
    status, err := container.HealthChecker(cli)
    if err != nil && err == ErrContainerNotStarted {
      logger.Infow("Container health status is not available yet", "container", container.Name)
			time.Sleep(delay)
			delay *= 2
			retries++
      continue
    } else if err != nil {
      logger.Errorw("Failed to inspect container", "container", container.Name, "error", err)
			return false
    }
    switch status {
		case "healthy":
			logger.Infow("Container is healthy", "container", container.Name)
			return true
		case "starting":
			logger.Infow("Container is starting", "container", container.Name)
		default:
			logger.Warnw("Container is unhealthy, retrying...", "container", container.Name, "status", status)
		}

		time.Sleep(delay)
		delay *= 2
		retries++
	}
  logger.Warnw("Container did not become healthy within the retry limit", "container", container.Name)
	return false
}


