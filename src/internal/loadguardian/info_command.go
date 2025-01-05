package loadguardian

import (
	"context"
	"encoding/json"

	"github.com/GrGLeo/LoadBalancer/src/pkg/utils"
	"github.com/docker/docker/api/types/container"
)

type InfoResponse struct {
  Response []ServiceResponse `json:"response"`
}

type ServiceResponse struct {
  ServiceName string `json:"service"`
  Container []ContainerResponse `json:"container"`
}

type ContainerResponse struct {
  Name string `json:"name"`
  Health string `json:"health"`
  Memory float64 `json:"memory"`
  CPU float64 `json:"cpu"`
}

func InfoProcess() (string, error) {
  lg := GetLoadGuardian()
  var resp InfoResponse
  for name, service := range lg.RunningServices {
    var servResp ServiceResponse
    servResp.ServiceName = name
    for _, cont  := range service {
      var contResp ContainerResponse
      contResp.Name = cont.Name
      stats, err := lg.Client.ContainerStatsOneShot(context.Background(), cont.ID)
      if err != nil {
        lg.Logger.Errorw("Failed to read stats", "container", cont.Name, "error", err.Error())
        return "", err
      }
      defer stats.Body.Close()


      var statsInfo container.StatsResponse
      if err := json.NewDecoder(stats.Body).Decode(&statsInfo); err != nil {
      }
      cpuDelta := float64(statsInfo.CPUStats.CPUUsage.TotalUsage - statsInfo.PreCPUStats.CPUUsage.TotalUsage)
      systemDelta := float64(statsInfo.CPUStats.SystemUsage - statsInfo.PreCPUStats.SystemUsage)
      numCores := float64(len(statsInfo.CPUStats.CPUUsage.PercpuUsage))
      cpuUsage := (cpuDelta / systemDelta) * numCores * 100.0
      contResp.CPU = cpuUsage

      memoryUsage := float64(statsInfo.MemoryStats.Usage) / (1024 * 1024) // Convert to MB
      memoryLimit := float64(statsInfo.MemoryStats.Limit) / (1024 * 1024) // Convert to MB
      memoryPercent := (memoryUsage / memoryLimit) * 100.0
      contResp.Memory = memoryPercent

      servResp.Container = append(servResp.Container, contResp)
    }
    resp.Response = append(resp.Response, servResp)
  }
  encodedResp, err := json.Marshal(resp)
  if err != nil {
    lg.Logger.Errorw("Failed to marshal response", "error", err.Error())
  }
  return string(encodedResp), nil
}


func GenerateTable(resp InfoResponse) string {
  header := []string{"Service", "Container", "Health", "CPU", "Memory"}
  baseLength := utils.GetBaseLength(header)
  for _, service := range resp.Response {
    for j, container := range service.Container {
      var row []string
      var rowLength []int
      if j == 0 {
        row = append(row, service.ServiceName)
        rowLength = append(rowLength, len(service.ServiceName))
      } else {
        row = append(row, "")
        rowLength = append(rowLength, 0)
      }
      row = append(row, container.Name)
      rowLength = append(rowLength, len(container.Name))
      row = append(row, container.Health)
      rowLength = append(rowLength, len(container.Health))
      cpu := utils.ConvertFloatToValue(container.CPU, "%")
      row = append(row, cpu)
      rowLength = append(rowLength, len(cpu))
      memory := utils.ConvertFloatToValue(container.Memory, "MB")
      row = append(row, memory)
      rowLength = append(rowLength, len(memory))
      err := utils.UpdateBaseLength(&baseLength, &rowLength)
      if err != nil {
        return "Failed to prep table"
      }
    }
  }
  return ""
}
