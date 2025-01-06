package servicemanager

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"go.uber.org/zap"
)


type Container struct {
  ID string // Container ID
  Name string // Container Name
  Url string // Container URL
  Health string // Container Health (starting | healthy | unhealthy)
  CpuUsage float64
  Memory float64
  Port int // Container Port
}

type ContainerRollbackConfig struct {
  ServiceName string // Service name for the running services key
  Index int // index of the container in the running list
  PastService Service // store the past service config
  New Container // new container info
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
  logger.Infow("Container stopped", "container", c.Name)
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

func (c *Container) Info(cli *client.Client, logger *zap.SugaredLogger) {
  resp, err := cli.ContainerInspect(context.Background(), c.ID)
  health := resp.State.Health.Status
  c.Health = health
  stats, err := cli.ContainerStatsOneShot(context.Background(), c.ID)
  if err != nil {
    logger.Errorw("Failed to read stats", "container", c.Name, "error", err.Error())
    return
  }
  defer stats.Body.Close()

  var statsInfo container.StatsResponse
  if err := json.NewDecoder(stats.Body).Decode(&statsInfo); err != nil {
  }
  cpuDelta := float64(statsInfo.CPUStats.CPUUsage.TotalUsage - statsInfo.PreCPUStats.CPUUsage.TotalUsage)
  systemDelta := float64(statsInfo.CPUStats.SystemUsage - statsInfo.PreCPUStats.SystemUsage)
  numCores := float64(len(statsInfo.CPUStats.CPUUsage.PercpuUsage))
  cpuUsage := (cpuDelta / systemDelta) * numCores * 100.0
  c.CpuUsage = cpuUsage

  memoryUsage := float64(statsInfo.MemoryStats.Usage) / (1024 * 1024) // Convert to MB
  memoryLimit := float64(statsInfo.MemoryStats.Limit) / (1024 * 1024) // Convert to MB
  memoryPercent := (memoryUsage / memoryLimit) * 100.0
  c.Memory = memoryPercent
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
