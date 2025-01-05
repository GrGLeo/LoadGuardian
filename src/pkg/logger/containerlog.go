package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"log"

	servicemanager "github.com/GrGLeo/LoadBalancer/src/internal/servicemanager"
	"github.com/GrGLeo/LoadBalancer/src/pkg/cleaner"
	"go.uber.org/zap"
)

func PrintLogs(logChannel <-chan servicemanager.LogMessage, logger *zap.SugaredLogger) {
  for logMessage := range logChannel {
    cleanMessage := cleaner.SanitizeLogMessage(logMessage.Message)
    logger.Infow(cleanMessage, "container", logMessage.ContainerName)
  }
}

func ReadProgress(r io.ReadCloser, updateStatus func(string)) error {
  defer r.Close()
  decoder := json.NewDecoder(r)
  for {
    var msg map[string]interface{}
    if err := decoder.Decode(&msg); err == io.EOF {
      break
    } else if err != nil {
      log.Fatal(err)
    }
    if id, ok := msg["id"]; ok {
      fmt.Printf("Image Id: %s\n", id)
    }
    if status, ok := msg["status"].(string); ok {
      updateStatus(status)
    }
  }
  return nil 
}

